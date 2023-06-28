package search

import (
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/util"
	"io"
)

var _ DocIdSet = &IntArrayDocIdSet{}

type IntArrayDocIdSet struct {
	docs []int
}

func (r *IntArrayDocIdSet) Iterator() index.DocIdSetIterator {
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

var _ index.DocIdSetIterator = &IntArrayDocIdSetIterator{}

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

func (r *IntArrayDocIdSetIterator) NextDoc() (int, error) {
	if r.i == len(r.docs) {
		return -1, io.EOF
	}

	r.doc = r.docs[r.i]
	r.i++
	return r.doc, nil
}

func (r *IntArrayDocIdSetIterator) Advance(target int) (int, error) {
	return r.SlowAdvance(target)
}

func (r *IntArrayDocIdSetIterator) SlowAdvance(target int) (int, error) {
	return index.SlowAdvance(r, target)
}

func (r *IntArrayDocIdSetIterator) Cost() int64 {
	return int64(len(r.docs))
}
