package lucene84

import (
	"context"
	"io"

	"github.com/geange/lucene-go/core/codecs/types"
	"github.com/geange/lucene-go/core/document"
	coreIndex "github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
)

type IntBlockTermState struct {
	types.BlockTermStateBase

	DocStartFP         int64
	PosStartFP         int64
	PayStartFP         int64
	SkipOffset         int64
	LastPosBlockOffset int64
	SingletonDocID     int
}

func NewIntBlockTermState() *IntBlockTermState {
	return &IntBlockTermState{
		SkipOffset:         -1,
		LastPosBlockOffset: -1,
		SingletonDocID:     -1,
	}
}

func (s *IntBlockTermState) Clone() *IntBlockTermState {
	state := &IntBlockTermState{
		BlockTermStateBase: types.BlockTermStateBase{},
		DocStartFP:         0,
		PosStartFP:         0,
		PayStartFP:         0,
		SkipOffset:         0,
		LastPosBlockOffset: 0,
		SingletonDocID:     0,
	}
	state.CopyFrom(s)
	return state
}

func (s *IntBlockTermState) CopyFrom(other index.TermState) {
	state, ok := other.(*IntBlockTermState)
	if ok {
		state.BlockTermStateBase.CopyFrom(&state.BlockTermStateBase)

		s.DocStartFP = state.DocStartFP
		s.PosStartFP = state.PosStartFP
		s.PayStartFP = state.PayStartFP
		s.SkipOffset = state.SkipOffset
		s.LastPosBlockOffset = state.LastPosBlockOffset
		s.SingletonDocID = state.SingletonDocID
	}
}

var _ index.PostingsEnum = &BlockDocsEnum{}

type BlockDocsEnum struct {
	docBuffer     []int
	freqBuffer    []int
	docBufferUpto int

	//private Lucene84SkipReader skipper;
	skipped bool

	startDocIn store.IndexInput

	docIn            store.IndexInput
	indexHasFreq     bool
	indexHasPos      bool
	indexHasOffsets  bool
	indexHasPayloads bool

	docFreq       int   // number of docs in this posting list
	totalTermFreq int   // sum of freqBuffer in this posting list (or docFreq when omitted)
	blockUpto     int   // number of docs in or before the current block
	doc           int   // doc we last read
	accum         int64 // accumulator for doc deltas

	// Where this term's postings start in the .doc file:
	docTermStartFP int64

	// Where this term's skip data starts (after
	// docTermStartFP) in the .doc file (or -1 if there is
	// no skip data for this term):
	skipOffset int64

	// docID for next skip point, we won't use skipper if
	// target docID is not larger than this
	nextSkipDoc int

	needsFreq bool // true if the caller actually needs frequencies
	// as we read freqBuffer lazily, isFreqsRead shows if freqBuffer are read for the current block
	// always true when we don't have freqBuffer (indexHasFreq=false) or don't need freqBuffer (needsFreq=false)
	isFreqsRead    bool
	singletonDocID int // docid when there is a single pulsed posting, otherwise -1
}

func NewBlockDocsEnum(fieldInfo *document.FieldInfo) *BlockDocsEnum {
	panic("")
}

func (b *BlockDocsEnum) CanReuse(docIn store.IndexInput, fieldInfo *document.FieldInfo) bool {
	return docIn == b.startDocIn &&
		b.indexHasFreq == (fieldInfo.GetIndexOptions() >= document.INDEX_OPTIONS_DOCS_AND_FREQS) &&
		b.indexHasPos == (fieldInfo.GetIndexOptions() >= document.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS) &&
		b.indexHasPayloads == fieldInfo.HasPayloads()
}

func (b *BlockDocsEnum) Reset(termState *IntBlockTermState, flags int) (*BlockDocsEnum, error) {
	b.docFreq = termState.DocFreq
	b.totalTermFreq = b.docFreq
	if b.indexHasFreq {
		b.totalTermFreq = termState.TotalTermFreq
	}

	b.docTermStartFP = termState.DocStartFP
	b.skipOffset = termState.SkipOffset
	b.singletonDocID = termState.SingletonDocID
	if b.docFreq > 1 {
		if b.docIn == nil {
			// lazy init
			b.docIn = b.startDocIn.Clone().(store.IndexInput)
		}
		if _, err := b.docIn.Seek(b.docTermStartFP, io.SeekStart); err != nil {
			return nil, err
		}
	}

	b.doc = -1
	b.needsFreq = coreIndex.FeatureRequested(flags, coreIndex.POSTINGS_ENUM_FREQS)
	b.isFreqsRead = true
	if b.indexHasFreq == false || b.needsFreq == false {
		for i := 0; i < FOR_UTIL_BLOCK_SIZE; i++ {
			b.freqBuffer[i] = 1
		}
	}
	b.accum = 0
	b.blockUpto = 0
	b.nextSkipDoc = BLOCK_SIZE - 1 // we won't skip if target is found in first block
	b.docBufferUpto = BLOCK_SIZE
	b.skipped = false
	return b, nil
}

func (b *BlockDocsEnum) DocID() int {
	//TODO implement me
	panic("implement me")
}

func (b *BlockDocsEnum) NextDoc(ctx context.Context) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (b *BlockDocsEnum) Advance(ctx context.Context, target int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (b *BlockDocsEnum) SlowAdvance(ctx context.Context, target int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (b *BlockDocsEnum) Cost() int64 {
	//TODO implement me
	panic("implement me")
}

func (b *BlockDocsEnum) Freq() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (b *BlockDocsEnum) NextPosition() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (b *BlockDocsEnum) StartOffset() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (b *BlockDocsEnum) EndOffset() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (b *BlockDocsEnum) GetPayload() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

var _ index.PostingsEnum = &EverythingEnum{}

// EverythingEnum
// Also handles payloads + offsets
type EverythingEnum struct {
	docBuffer      []int64
	freqBuffer     []int64
	posDeltaBuffer []int64

	payloadLengthBuffer    []int64
	offsetStartDeltaBuffer []int64
	offsetLengthBuffer     []int64

	payloadBytes    []byte
	payloadByteUpto int
	payloadLength   int

	lastStartOffset int
	startOffset     int
	endOffset       int

	docBufferUpto int
	posBufferUpto int

	//private Lucene84SkipReader skipper;
	skipped bool

	startDocIn store.IndexInput

	docIn   store.IndexInput
	posIn   store.IndexInput
	payIn   store.IndexInput
	payload []byte

	indexHasOffsets  bool
	indexHasPayloads bool

	docFreq       int // number of docs in this posting list
	totalTermFreq int // number of positions in this posting list
	blockUpto     int // number of docs in or before the current block
	doc           int // doc we last read
	accum         int // accumulator for doc deltas
	freq          int // freq we last read
	position      int // current position

	// how many positions "behind" we are; nextPosition must
	// skip these to "catch up":
	posPendingCount int

	// Lazy pos seek: if != -1 then we must seek to this FP
	// before reading positions:
	posPendingFP int

	// Lazy pay seek: if != -1 then we must seek to this FP
	// before reading payloads/offsets:
	payPendingFP int

	// Where this term's postings start in the .doc file:
	docTermStartFP int

	// Where this term's postings start in the .pos file:
	posTermStartFP int

	// Where this term's payloads/offsets start in the .pay
	// file:
	payTermStartFP int

	// File pointer where the last (vInt encoded) pos delta
	// block is.  We need this to know whether to bulk
	// decode vs vInt decode the block:
	lastPosBlockFP int

	// Where this term's skip data starts (after
	// docTermStartFP) in the .doc file (or -1 if there is
	// no skip data for this term):
	skipOffset int

	nextSkipDoc int

	needsOffsets   bool // true if we actually need offsets
	needsPayloads  bool // true if we actually need payloads
	singletonDocID int  // docid when there is a single pulsed posting, otherwise -1
}

func (e *EverythingEnum) DocID() int {
	//TODO implement me
	panic("implement me")
}

func (e *EverythingEnum) NextDoc(ctx context.Context) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (e *EverythingEnum) Advance(ctx context.Context, target int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (e *EverythingEnum) SlowAdvance(ctx context.Context, target int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (e *EverythingEnum) Cost() int64 {
	//TODO implement me
	panic("implement me")
}

func (e *EverythingEnum) Freq() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (e *EverythingEnum) NextPosition() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (e *EverythingEnum) StartOffset() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (e *EverythingEnum) EndOffset() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (e *EverythingEnum) GetPayload() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}
