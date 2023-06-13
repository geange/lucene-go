package search

import "github.com/geange/lucene-go/core/index"

var _ index.DocIdSetIterator = &BlockMaxDISI{}

type BlockMaxDISI struct {
	*index.DocIdSetIteratorDefault
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
