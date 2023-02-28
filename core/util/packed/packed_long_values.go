package packed

import "github.com/geange/lucene-go/core/types"

var _ types.LongValues = &PackedLongValues{}

// PackedLongValues Utility class to compress integers into a LongValues instance.
// TODO: need to reduce memory
type PackedLongValues struct {
	values []uint64
}

func NewPackedLongValues(values []uint64) *PackedLongValues {
	return &PackedLongValues{values: values}
}

func NewPackedLongValuesV1() *PackedLongValues {
	return &PackedLongValues{values: make([]uint64, 0)}
}

func (p *PackedLongValues) Size() int64 {
	return int64(len(p.values))
}

func (p *PackedLongValues) Get(index int64) int64 {
	return int64(p.values[index])
}

type PackedLongValuesIterator struct {
	p   *PackedLongValues
	pos int
}

// HasNext Whether or not there are remaining values.
func (i *PackedLongValuesIterator) HasNext() bool {
	return i.pos < int(i.p.Size())
}

func (i *PackedLongValuesIterator) Next() int64 {
	if i.HasNext() {
		v := i.p.Get(int64(i.pos))
		i.pos++
		return v
	}
	return 0
}

func (p *PackedLongValues) Iterator() *PackedLongValuesIterator {
	iterator := &PackedLongValuesIterator{
		p:   p,
		pos: -1,
	}
	return iterator
}

type PackedLongValuesBuilder struct {
	pending []uint64
}

func NewPackedLongValuesBuilder(v []uint64) *PackedLongValuesBuilder {
	return &PackedLongValuesBuilder{pending: v}
}

func NewPackedLongValuesBuilderV1() *PackedLongValuesBuilder {
	return &PackedLongValuesBuilder{pending: make([]uint64, 0)}
}

// Add a new element to this builder.
func (p *PackedLongValuesBuilder) Add(value int64) {
	p.pending = append(p.pending, uint64(value))
}

func (p *PackedLongValuesBuilder) Build() *PackedLongValues {
	return NewPackedLongValues(p.pending)
}
