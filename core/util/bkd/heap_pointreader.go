package bkd

import (
	"encoding/binary"
	"github.com/geange/lucene-go/core/util"
)

var _ PointReader = &HeapPointReader{}

// HeapPointReader Utility class to read buffered points from in-heap arrays.
// lucene.internal
type HeapPointReader struct {
	curRead    int
	block      []byte
	config     *Config
	end        int
	pointValue *HeapPointValue
}

func NewHeapPointReader(config *Config, block []byte, start, end int) *HeapPointReader {
	reader := &HeapPointReader{
		curRead: start - 1,
		block:   block,
		config:  config,
		end:     end,
	}

	if start < end {
		reader.pointValue = NewHeapPointValue(config, block)
	}
	return reader
}

func (h *HeapPointReader) Close() error {
	return nil
}

func (h *HeapPointReader) Next() (bool, error) {
	h.curRead++
	return h.curRead < h.end, nil
}

func (h *HeapPointReader) PointValue() PointValue {
	h.pointValue.SetOffset(h.curRead * h.config.BytesPerDoc())
	return h.pointValue
}

var _ PointValue = &HeapPointValue{}

// HeapPointValue Reusable implementation for a point value on-heap
type HeapPointValue struct {
	packedValue       *util.BytesRef
	packedValueDocID  *util.BytesRef
	packedValueLength int
}

func NewHeapPointValue(config *Config, value []byte) *HeapPointValue {
	return &HeapPointValue{
		packedValue:       util.NewBytesRef(value, 0, config.PackedBytesLength()),
		packedValueDocID:  util.NewBytesRef(value, 0, config.BytesPerDoc()),
		packedValueLength: config.PackedBytesLength(),
	}
}

func (h *HeapPointValue) SetOffset(offset int) {
	h.packedValue.Offset = offset
	h.packedValueDocID.Offset = offset
}

func (h *HeapPointValue) PackedValue() []byte {
	return h.packedValue.GetBytes()
}

func (h *HeapPointValue) DocID() int {
	return int(binary.BigEndian.Uint32(h.packedValueDocID.GetBytes()[h.packedValueLength:]))
}

func (h *HeapPointValue) PackedValueDocIDBytes() []byte {
	return h.packedValueDocID.GetBytes()
}
