package index

import (
	"github.com/bits-and-blooms/bitset"
	"io"
)

var _ DocIdSetIterator = &BitSetIterator{}

type BitSetIterator struct {
	*DocIdSetIteratorDefault
	bits *bitset.BitSet
	cost int
	doc  int
}

func NewBitSetIterator(bits *bitset.BitSet, cost int) *BitSetIterator {
	it := &BitSetIterator{
		bits: bits,
		cost: cost,
		doc:  -1,
	}
	it.DocIdSetIteratorDefault = NewDocIdSetIteratorDefault(&DocIdSetIteratorDefaultConfig{
		NextDoc: it.NextDoc,
	})
	return it
}

func (b *BitSetIterator) DocID() int {
	return int(b.doc)
}

func (b *BitSetIterator) NextDoc() (int, error) {
	return b.Advance(int(b.doc + 1))
}

func (b *BitSetIterator) Advance(target int) (int, error) {
	value, ok := b.bits.NextSet(uint(target))
	if !ok {
		return 0, io.EOF
	}

	b.doc = int(value)
	return int(b.doc), nil
}

func (b *BitSetIterator) Cost() int64 {
	return int64(b.cost)
}
