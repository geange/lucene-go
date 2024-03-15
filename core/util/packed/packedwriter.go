package packed

import (
	"errors"

	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/packed/bulkoperation"
	"github.com/geange/lucene-go/core/util/packed/common"
)

var _ Writer = &PackedWriter{}

type PackedWriter struct {
	*BaseWriter

	finished   bool
	format     Format
	encoder    common.BulkOperation
	nextBlocks []byte
	nextValues []uint64
	iterations int
	off        int
	written    int
}

func NewPackedWriter(format Format, out store.DataOutput, valueCount, bitsPerValue, mem int) *PackedWriter {
	encoder, err := Of(format, bitsPerValue)
	if err != nil {
		return nil
	}

	iterations := bulkoperation.ComputeIterations(encoder, valueCount, mem)

	packedWriter := &PackedWriter{
		finished:   false,
		format:     format,
		encoder:    encoder,
		nextBlocks: make([]byte, iterations*encoder.ByteBlockCount()),
		nextValues: make([]uint64, iterations*encoder.ByteValueCount()),
		iterations: iterations,
		off:        0,
		written:    0,
	}
	packedWriter.BaseWriter = newBaseWriter(packedWriter, out, valueCount, bitsPerValue)

	return packedWriter
}

func (p *PackedWriter) GetFormat() Format {
	return p.format
}

func (p *PackedWriter) Add(v uint64) error {
	if p.valueCount != -1 && p.written >= p.valueCount {
		return errors.New("writing past end of stream")
	}
	p.nextValues[p.off] = v
	p.off++
	if p.off == len(p.nextValues) {
		if err := p.flush(); err != nil {
			return err
		}
	}
	p.written++
	return nil
}

func (p *PackedWriter) Finish() error {
	if p.valueCount != -1 {
		for p.written < p.valueCount {
			err := p.Add(0)
			if err != nil {
				return err
			}
		}
	}
	if err := p.flush(); err != nil {
		return err
	}
	p.finished = true
	return nil
}

func (p *PackedWriter) flush() error {
	p.encoder.EncodeBytes(p.nextValues[0:], p.nextBlocks[0:], p.iterations)
	blockCount := p.format.ByteCount(VERSION_CURRENT, p.off, p.bitsPerValue)
	if _, err := p.out.Write(p.nextBlocks[:blockCount]); err != nil {
		return err
	}

	for i := range p.nextValues {
		p.nextValues[i] = 0
	}

	p.off = 0
	return nil
}

func (p *PackedWriter) Ord() int {
	return p.written - 1
}
