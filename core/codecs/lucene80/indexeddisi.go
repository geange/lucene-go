package lucene80

import (
	"context"

	"github.com/geange/lucene-go/core/types"
)

var _ types.DocIdSetIterator = &IndexedDISI{}

type IndexedDISI struct {
}

func (i *IndexedDISI) DocID() int {
	//TODO implement me
	panic("implement me")
}

func (i *IndexedDISI) NextDoc(ctx context.Context) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (i *IndexedDISI) Advance(ctx context.Context, target int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (i *IndexedDISI) SlowAdvance(ctx context.Context, target int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (i *IndexedDISI) Cost() int64 {
	//TODO implement me
	panic("implement me")
}

func (i *IndexedDISI) AdvanceExact(ctx context.Context, target int) (bool, error) {
	panic("implement me")
}
