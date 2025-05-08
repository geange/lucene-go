package lucene84

import (
	"context"

	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
)

var _ index.ImpactsEnum = &BlockImpactsEverythingEnum{}

type BlockImpactsEverythingEnum struct {
	forUtil      *ForUtil
	forDeltaUtil *ForDeltaUtil
	pforUtil     *PForUtil

	docBuffer      []uint64
	freqBuffer     []uint64
	posDeltaBuffer []uint64

	payloadLengthBuffer    []uint64
	offsetStartDeltaBuffer []uint64
	offsetLengthBuffer     []uint64

	payloadBytes    []byte
	payloadByteUpto int
	payloadLength   int

	lastStartOffset int
	startOffset     int
	endOffset       int

	docBufferUpto int
	posBufferUpto int

	skipper *ScoreSkipReader

	docIn   store.IndexInput
	posIn   store.IndexInput
	payIn   store.IndexInput
	payload []byte

	indexHasFreq     bool
	indexHasPos      bool
	indexHasOffsets  bool
	indexHasPayloads bool

	docFreq       int // number of docs in this posting list
	totalTermFreq int // number of positions in this posting list
	docUpto       int // how many docs we've read
	posDocUpTo    int // for how many docs we've read positions, offsets, and payloads
	doc           int // doc we last read
	accum         int // accumulator for doc deltas
	position      int // current position

	// how many positions "behind" we are; nextPosition must
	// skip these to "catch up":
	posPendingCount int

	// Lazy pos seek: if != -1 then we must seek to this FP
	// before reading positions:
	posPendingFP uint64

	// Lazy pay seek: if != -1 then we must seek to this FP
	// before reading payloads/offsets:
	payPendingFP uint64

	// Where this term's postings start in the .doc file:
	docTermStartFP uint64

	// Where this term's postings start in the .pos file:
	posTermStartFP uint64

	// Where this term's payloads/offsets start in the .pay
	// file:
	payTermStartFP uint64

	// File pointer where the last (vInt encoded) pos delta
	// block is.  We need this to know whether to bulk
	// decode vs vInt decode the block:
	lastPosBlockFP uint64

	nextSkipDoc int

	needsPositions bool
	needsOffsets   bool // true if we actually need offsets
	needsPayloads  bool // true if we actually need payloads

	isFreqsRead bool // shows if freqBuffer for the current doc block are read into freqBuffer

	seekTo int64
}

func (b *BlockImpactsEverythingEnum) DocID() int {
	//TODO implement me
	panic("implement me")
}

func (b *BlockImpactsEverythingEnum) NextDoc(ctx context.Context) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (b *BlockImpactsEverythingEnum) Advance(ctx context.Context, target int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (b *BlockImpactsEverythingEnum) SlowAdvance(ctx context.Context, target int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (b *BlockImpactsEverythingEnum) Cost() int64 {
	//TODO implement me
	panic("implement me")
}

func (b *BlockImpactsEverythingEnum) Freq() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (b *BlockImpactsEverythingEnum) NextPosition() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (b *BlockImpactsEverythingEnum) StartOffset() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (b *BlockImpactsEverythingEnum) EndOffset() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (b *BlockImpactsEverythingEnum) GetPayload() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (b *BlockImpactsEverythingEnum) AdvanceShallow(ctx context.Context, target int) error {
	//TODO implement me
	panic("implement me")
}

func (b *BlockImpactsEverythingEnum) GetImpacts() (index.Impacts, error) {
	//TODO implement me
	panic("implement me")
}
