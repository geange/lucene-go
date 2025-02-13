package packed

import (
	"context"
	"errors"
	"fmt"
	"math"
	"slices"

	"github.com/geange/lucene-go/core/store"
)

const (
	MIN_BLOCK_SHIFT = 2
	MAX_BLOCK_SHIFT = 22
)

type DirectMonotonicWriter struct {
	meta            store.IndexOutput
	data            store.IndexOutput
	numValues       int64
	baseDataPointer int64
	buffer          []int64
	bufferMaxSize   int
	count           int64
	finished        bool
	previous        int64
}

func NewDirectMonotonicWriter(metaOut, dataOut store.IndexOutput, numValues int64, blockShift int) (*DirectMonotonicWriter, error) {
	if blockShift < MIN_BLOCK_SHIFT || blockShift > MAX_BLOCK_SHIFT {
		return nil, fmt.Errorf("blockShift must be in [%d, %d]", MIN_BLOCK_SHIFT, MAX_BLOCK_SIZE)
	}
	if numValues < 0 {
		return nil, errors.New("numValues can't be negative")
	}
	// TODO: verify numBlocks
	//numBlocks := 0
	//if numValues != 0 {
	//	numBlocks = ((numValues - 1) >> blockShift) + 1
	//}

	blockSize := 1 << blockShift

	bufferMaxSize := int(min(numValues, int64(blockSize)))

	writer := &DirectMonotonicWriter{
		meta:            metaOut,
		data:            dataOut,
		numValues:       numValues,
		baseDataPointer: dataOut.GetFilePointer(),
		buffer:          make([]int64, 0, bufferMaxSize),
		bufferMaxSize:   bufferMaxSize,
		count:           0,
		finished:        false,
		previous:        math.MinInt64,
	}
	return writer, nil
}

func (d *DirectMonotonicWriter) Add(v int64) error {
	if v < d.previous {
		return errors.New("values do not come in order")
	}
	if d.bufferMaxSize == len(d.buffer) {
		if err := d.flush(); err != nil {
			return err
		}
	}
	d.buffer = append(d.buffer, v)
	d.previous = v
	d.count++
	return nil
}

func (d *DirectMonotonicWriter) flush() error {
	avgInc := float64(d.buffer[len(d.buffer)-1]-d.buffer[0]) / float64(max(1, len(d.buffer)-1))
	for i := range d.buffer {
		expectedInc := int64(avgInc * float64(i))
		d.buffer[i] -= expectedInc
	}

	minNum := slices.Min(d.buffer)
	maxDelta := int64(0)

	for i := range d.buffer {
		d.buffer[i] -= minNum
		// use | will change nothing when it comes to computing required bits
		// but has the benefit of working fine with negative values too
		// (in case of overflow)
		maxDelta |= d.buffer[i]
	}

	ctx := context.Background()
	d.meta.WriteUint64(ctx, uint64(minNum))
	d.meta.WriteUint32(ctx, math.Float32bits(float32(avgInc)))
	d.meta.WriteUint64(ctx, uint64(d.data.GetFilePointer()-d.baseDataPointer))

	if maxDelta == 0 {
		d.meta.WriteByte(0)
	} else {
		bitsRequired := UnsignedBitsRequired(uint64(maxDelta))
		writer, err := DirectWriterGetInstance(d.data, len(d.buffer), bitsRequired)
		if err != nil {
			return nil
		}

		for _, n := range d.buffer {
			writer.Add(uint64(n))
		}

		writer.Finish()
		d.meta.WriteByte(byte(bitsRequired))
	}
	d.buffer = d.buffer[:0]
	return nil
}

// Finish
// This must be called exactly once after all values have been added.
func (d *DirectMonotonicWriter) Finish() error {
	if d.count != d.numValues {
		return errors.New("wrong number of values added, expected")
	}
	if d.finished {
		return errors.New("#finish has been called already")
	}
	if len(d.buffer) > 0 {
		if err := d.flush(); err != nil {
			return err
		}
	}
	d.finished = true
	return nil
}

func DirectMonotonicWriterGetInstance(metaOut, dataOut store.IndexOutput, numValues int64, blockShift int) (*DirectMonotonicWriter, error) {
	return NewDirectMonotonicWriter(metaOut, dataOut, numValues, blockShift)
}
