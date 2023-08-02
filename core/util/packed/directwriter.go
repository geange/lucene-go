package packed

import (
	"errors"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/packed/common"
	"math"
)

// DirectWriter
// Class for writing packed integers to be directly read from Directory.
// Integers can be read on-the-fly via DirectReader.
// Unlike PackedInts, it optimizes for read i/o operations and supports > 2B values.
type DirectWriter struct {
	bitsPerValue int
	numValues    int
	output       store.DataOutput
	count        int
	finished     bool

	// for now, just use the existing writer under the hood
	off               int
	nextBlocks        []byte
	nextValues        []uint64
	nextValuesMaxSize int
	encoder           common.BulkOperation
	iterations        int
}

func NewDirectWriter(output store.DataOutput, numValues, bitsPerValue int) (*DirectWriter, error) {
	encoder, err := Of(FormatPacked, bitsPerValue)
	if err != nil {
		return nil, err
	}

	iterations := encoder.(interface {
		ComputeIterations(valueCount, ramBudget int) int
	}).ComputeIterations(min(numValues, math.MaxInt32), DEFAULT_BUFFER_SIZE)

	writer := &DirectWriter{
		bitsPerValue: bitsPerValue,
		numValues:    numValues,
		output:       output,
		nextBlocks:   make([]byte, iterations*encoder.ByteBlockCount()),
		nextValues:   make([]uint64, iterations*encoder.ByteValueCount()),
		encoder:      encoder,
		iterations:   iterations,
	}
	return writer, nil
}

func (d *DirectWriter) Add(v uint64) error {
	if d.count >= d.numValues {
		return errors.New("writing past end of stream")
	}

	d.nextValues[d.off] = v
	d.off++

	if len(d.nextValues) == d.off {
		if err := d.flush(); err != nil {
			return err
		}
	}
	d.count++
	return nil
}

func (d *DirectWriter) flush() error {
	d.encoder.EncodeBytes(d.nextValues, d.nextBlocks, d.iterations)
	blockCount := FormatPacked.ByteCount(VERSION_CURRENT, len(d.nextValues), d.bitsPerValue)
	if _, err := d.output.Write(d.nextBlocks[:blockCount]); err != nil {
		return err
	}
	d.off = 0
	return nil
}

// Finish
// finishes writing
func (d *DirectWriter) Finish() error {
	if d.count != d.numValues {
		return errors.New("wrong number of values added")
	}

	if err := d.flush(); err != nil {
		return err
	}
	// pad for fast io: we actually only need this for certain BPV, but its just 3 bytes...
	for i := 0; i < 3; i++ {
		if err := d.output.WriteByte(0); err != nil {
			return err
		}
	}
	d.finished = true
	return nil
}

// GetInstance
// Returns an instance suitable for encoding numValues using bitsPerValue
//func GetInstance(output store.DataOutput, numValues, bitsPerValue int) (*DirectWriter, error) {
//	if _, ok := slices.BinarySearch(SUPPORTED_BITS_PER_VALUE, bitsPerValue); !ok {
//		return nil, errors.New("unsupported bitsPerValue")
//	}
//	return NewDirectWriter(output, numValues, bitsPerValue)
//}

// Round a number of bits per value to the next amount of bits per value that is supported by this writer.
// bitsRequired – the amount of bits required
// the next number of bits per value that is gte the provided value and supported by this writer
//func roundBits(bitsRequired int) int {
//	index, _ := slices.BinarySearch(SUPPORTED_BITS_PER_VALUE, bitsRequired)
//	if index < 0 {
//		return SUPPORTED_BITS_PER_VALUE[-index-1]
//	} else {
//		return bitsRequired
//	}
//}
