package types

import (
	coreindex "github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/interface/index"
)

var _ index.TermState = &BlockTermState{}

type BlockTermState struct {
	coreindex.OrdTermState

	DocFreq          int   // how many docs have this term
	TotalTermFreq    int   // total number of occurrences of this term
	TermBlockOrd     int   // the term's ord in the current block
	BlockFilePointer int64 // fp into the terms dict primary file (_X. tim) that holds this term
}

func (b *BlockTermState) CopyFrom(other index.TermState) {
	state, ok := other.(*BlockTermState)
	if ok {
		b.OrdTermState.CopyFrom(&state.OrdTermState)
		b.DocFreq = state.DocFreq
		b.TotalTermFreq = state.TotalTermFreq
		b.TermBlockOrd = state.TermBlockOrd
		b.BlockFilePointer = state.BlockFilePointer
	}
}
