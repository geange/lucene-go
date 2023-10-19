package search

import (
	"github.com/geange/lucene-go/core/types"
)

var _ types.DocIdSetIterator = &BlockMaxDISI{}

type BlockMaxDISI struct {
}

func (b *BlockMaxDISI) DocID() int {
	//TODO implement me
	panic("implement me")
}

func (b *BlockMaxDISI) NextDoc() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (b *BlockMaxDISI) Advance(target int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (b *BlockMaxDISI) Cost() int64 {
	//TODO implement me
	panic("implement me")
}

func (b *BlockMaxDISI) SlowAdvance(target int) (int, error) {
	return types.SlowAdvance(b, target)
}
