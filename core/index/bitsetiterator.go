package index

import (
	"context"
	"io"

	"github.com/bits-and-blooms/bitset"

	"github.com/geange/lucene-go/core/types"
)

var _ types.DocIdSetIterator = &BitSetIterator{}

type BitSetIterator struct {
	bits *bitset.BitSet
	cost int64
	doc  int
}

func NewBitSetIterator(bits *bitset.BitSet, cost int64) *BitSetIterator {
	it := &BitSetIterator{
		bits: bits,
		cost: cost,
		doc:  -1,
	}

	return it
}

func (b *BitSetIterator) GetBitSet() *bitset.BitSet {
	return b.bits
}

func (b *BitSetIterator) DocID() int {
	return b.doc
}

func (b *BitSetIterator) NextDoc(ctx context.Context) (int, error) {
	return b.Advance(ctx, b.doc+1)
}

func (b *BitSetIterator) Advance(ctx context.Context, target int) (int, error) {
	value, ok := b.bits.NextSet(uint(target))
	if !ok {
		return 0, io.EOF
	}

	b.doc = int(value)
	return b.doc, nil
}

func (b *BitSetIterator) SlowAdvance(ctx context.Context, target int) (int, error) {
	return types.SlowAdvanceWithContext(ctx, b, target)
}

func (b *BitSetIterator) Cost() int64 {
	return b.cost
}
