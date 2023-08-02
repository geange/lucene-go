package bkd

import (
	"encoding/binary"
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

	if len(block) > 0 {
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
	config *Config
	bytes  []byte
	offset int
}

func NewHeapPointValue(config *Config, value []byte) *HeapPointValue {
	return &HeapPointValue{
		config: config,
		bytes:  value,
	}
}

func (h *HeapPointValue) SetOffset(offset int) {
	h.offset = offset
}

func (h *HeapPointValue) PackedValue() []byte {
	return h.bytes[h.offset : h.offset+h.config.PackedBytesLength()]

}

func (h *HeapPointValue) DocID() int {
	bs := h.bytes[h.offset+h.config.PackedBytesLength():]
	return int(binary.BigEndian.Uint32(bs))
}

func (h *HeapPointValue) PackedValueDocIDBytes() []byte {
	return h.bytes[h.offset : h.offset+h.config.BytesPerDoc()]
}
