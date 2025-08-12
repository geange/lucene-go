package packed

import (
	"iter"

	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/packed/bulkoperation"
)

type PackedReaderIterator struct {
	in           store.DataInput
	format       Format
	valueCount   int
	bitsPerValue int
	mem          int
}

func NewPackedReaderIterator(in store.DataInput, format Format,
	valueCount int, bitsPerValue int, mem int) *PackedReaderIterator {
	return &PackedReaderIterator{in: in, format: format, valueCount: valueCount, bitsPerValue: bitsPerValue, mem: mem}
}

func (p *PackedReaderIterator) Iterator() (iter.Seq[uint64], error) {
	bulkOperation, err := Of(p.format, p.bitsPerValue)
	if err != nil {
		return nil, err
	}
	iterations := bulkoperation.ComputeIterations(bulkOperation, p.valueCount, p.mem)

	blocks := make([]byte, bulkOperation.ByteBlockCount()*iterations)
	values := make([]uint64, bulkOperation.ByteValueCount()*iterations)

	if _, err := p.in.Read(blocks); err != nil {
		return nil, err
	}
	bulkOperation.DecodeBytes(blocks, values, iterations)

	return func(yield func(uint64) bool) {
		for _, value := range values {
			if !yield(value) {
				return
			}
		}
	}, nil
}
