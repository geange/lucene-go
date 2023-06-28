package search

import (
	"github.com/bits-and-blooms/bitset"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/util"
)

var _ DocIdSet = &BitDocIdSet{}

type BitDocIdSet struct {
	set  *bitset.BitSet
	cost int64
}

func (b BitDocIdSet) Iterator() index.DocIdSetIterator {
	return index.NewBitSetIterator(b.set, b.cost)
}

func (b BitDocIdSet) Bits() util.Bits {
	return b.set
}

func NewBitDocIdSet(set *bitset.BitSet, cost int64) *BitDocIdSet {
	return &BitDocIdSet{set: set, cost: cost}
}
