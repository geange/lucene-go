package bkd

import (
	"context"
	"io"

	"github.com/geange/lucene-go/core/types"
)

var _ types.DocIdSetIterator = &readerDocIDSetIterator{}

// readerDocIDSetIterator Reusable DocIdSetIterator to handle low cardinality leaves.
type readerDocIDSetIterator struct {
	idx    int
	length int
	offset int
	docID  int
	docIDs []int
}

func newReaderDocIDSetIterator(maxPointsInLeafNode int) *readerDocIDSetIterator {
	return &readerDocIDSetIterator{
		docIDs: make([]int, maxPointsInLeafNode),
	}
}

func (r *readerDocIDSetIterator) DocID() int {
	return r.docID
}

func (r *readerDocIDSetIterator) NextDoc() (int, error) {
	if r.idx == r.length {
		r.docID = types.NO_MORE_DOCS
		return r.docID, io.EOF
	} else {
		r.docID = r.docIDs[r.offset+r.idx]
		r.idx++
	}
	return r.docID, nil
}

func (r *readerDocIDSetIterator) Advance(ctx context.Context, target int) (int, error) {
	return r.SlowAdvance(ctx, target)
}

func (r *readerDocIDSetIterator) SlowAdvance(ctx context.Context, target int) (int, error) {
	return types.SlowAdvance(r, target)
}

func (r *readerDocIDSetIterator) Cost() int64 {
	return int64(r.length)
}

func (r *readerDocIDSetIterator) reset(offset int, length int) {
	r.offset = offset
	r.length = length
	//assert offset + length <= docIDs.length;
	r.docID = -1
	r.idx = 0
}
