package search

import (
	"context"
	"io"

	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util"
)

var _ DocIdSet = &IntArrayDocIdSet{}

type IntArrayDocIdSet struct {
	docs []int
}

func (r *IntArrayDocIdSet) Iterator() types.DocIdSetIterator {
	//TODO implement me
	panic("implement me")
}

func (r *IntArrayDocIdSet) Bits() util.Bits {
	//TODO implement me
	panic("implement me")
}

func NewIntArrayDocIdSet(docs []int) *IntArrayDocIdSet {
	return &IntArrayDocIdSet{docs: docs}
}

var _ types.DocIdSetIterator = &IntArrayDocIdSetIterator{}

type IntArrayDocIdSetIterator struct {
	docs []int
	i    int
	doc  int
}

func NewIntArrayDocIdSetIterator(docs []int) *IntArrayDocIdSetIterator {
	return &IntArrayDocIdSetIterator{
		docs: docs,
		doc:  -1,
	}
}

func (r *IntArrayDocIdSetIterator) DocID() int {
	return r.doc
}

func (r *IntArrayDocIdSetIterator) NextDoc(context.Context) (int, error) {
	if r.i == len(r.docs) {
		return -1, io.EOF
	}

	r.doc = r.docs[r.i]
	r.i++
	return r.doc, nil
}

func (r *IntArrayDocIdSetIterator) Advance(ctx context.Context, target int) (int, error) {
	return r.SlowAdvance(nil, target)
}

func (r *IntArrayDocIdSetIterator) SlowAdvance(ctx context.Context, target int) (int, error) {
	return types.SlowAdvanceWithContext(ctx, r, target)
}

func (r *IntArrayDocIdSetIterator) Cost() int64 {
	return int64(len(r.docs))
}
