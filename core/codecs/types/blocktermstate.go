package types

import (
	coreindex "github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/interface/index"
)

//var _ index.TermState = &BlockTermState{}

type BlockTermState interface {
	SetDocFreq(DocFreq int)
	GetDocFreq() int
	SetTotalTermFreq(TotalTermFreq int)
	GetTotalTermFreq() int
	SetTermBlockOrd(TermBlockOrd int)
	GetTermBlockOrd() int
	SetBlockFilePointer(BlockFilePointer int64)
	GetBlockFilePointer() int64
	Clone() BlockTermState
}

var _ BlockTermState = &BlockTermStateBase{}

type BlockTermStateBase struct {
	coreindex.OrdTermState

	DocFreq          int   // how many docs have this term
	TotalTermFreq    int   // total number of occurrences of this term
	TermBlockOrd     int   // the term's ord in the current block
	BlockFilePointer int64 // fp into the terms dict primary file (_X. tim) that holds this term
}

func (b *BlockTermStateBase) Clone() BlockTermState {
	return &BlockTermStateBase{
		OrdTermState:     *b.OrdTermState.Clone(),
		DocFreq:          b.DocFreq,
		TotalTermFreq:    b.TotalTermFreq,
		TermBlockOrd:     b.TermBlockOrd,
		BlockFilePointer: b.BlockFilePointer,
	}
}

func (b *BlockTermStateBase) SetDocFreq(DocFreq int) {
	b.DocFreq = DocFreq
}

func (b *BlockTermStateBase) GetDocFreq() int {
	return b.DocFreq
}

func (b *BlockTermStateBase) SetTotalTermFreq(TotalTermFreq int) {
	b.TotalTermFreq = TotalTermFreq
}

func (b *BlockTermStateBase) GetTotalTermFreq() int {
	return b.TotalTermFreq
}

func (b *BlockTermStateBase) SetTermBlockOrd(TermBlockOrd int) {
	b.TermBlockOrd = TermBlockOrd
}

func (b *BlockTermStateBase) GetTermBlockOrd() int {
	return b.TermBlockOrd
}

func (b *BlockTermStateBase) SetBlockFilePointer(BlockFilePointer int64) {
	b.BlockFilePointer = BlockFilePointer
}

func (b *BlockTermStateBase) GetBlockFilePointer() int64 {
	return b.BlockFilePointer
}

func (b *BlockTermStateBase) CopyFrom(other index.TermState) {
	state, ok := other.(*BlockTermStateBase)
	if ok {
		b.OrdTermState.CopyFrom(&state.OrdTermState)
		b.DocFreq = state.DocFreq
		b.TotalTermFreq = state.TotalTermFreq
		b.TermBlockOrd = state.TermBlockOrd
		b.BlockFilePointer = state.BlockFilePointer
	}
}
