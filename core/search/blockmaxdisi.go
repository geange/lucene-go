package search

import (
	"context"

	"github.com/geange/lucene-go/core/types"
)

var _ types.DocIdSetIterator = &BlockMaxDISI{}

type BlockMaxDISI struct {
}

func (b *BlockMaxDISI) DocID() int {
	//TODO implement me
	panic("implement me")
}

func (b *BlockMaxDISI) NextDoc(context.Context) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (b *BlockMaxDISI) Advance(ctx context.Context, target int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (b *BlockMaxDISI) Cost() int64 {
	//TODO implement me
	panic("implement me")
}

func (b *BlockMaxDISI) SlowAdvance(ctx context.Context, target int) (int, error) {
	return types.SlowAdvanceWithContext(ctx, b, target)
}
