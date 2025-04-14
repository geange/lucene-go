package lucene84

import (
	"github.com/geange/lucene-go/core/codecs/types"
	"github.com/geange/lucene-go/core/document"
	coreIndex "github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
)

var _ types.PushPostingsWriter = &PostingsWriter{}

type PostingsWriter struct {
	types.PushPostingsWriterBase

	docOut store.IndexOutput
	posOut store.IndexOutput
	payOut store.IndexOutput

	emptyState *IntBlockTermState
	lastState  *IntBlockTermState

	// Holds starting file pointers for current term:
	docStartFP int64
	posStartFP int64
	payStartFP int64

	docDeltaBuffer []uint64
	freqBuffer     []uint64
	docBufferUpto  int

	posDeltaBuffer         []uint64
	payloadLengthBuffer    []uint64
	offsetStartDeltaBuffer []uint64
	offsetLengthBuffer     []uint64
	posBufferUpto          int

	payloadBytes             []byte
	payloadByteUpto          int
	lastBlockDocID           int
	lastBlockPosFP           int64
	lastBlockPayFP           int64
	lastBlockPosBufferUpto   int
	lastBlockPayloadByteUpto int

	lastDocID       int
	lastPosition    int
	lastStartOffset int
	docCount        int

	pforUtil     *PForUtil
	forDeltaUtil *ForDeltaUtil
	skipWriter   skipWriter

	fieldHasNorms                  bool
	norms                          index.NumericDocValues
	competitiveFreqNormAccumulator coreIndex.CompetitiveImpactAccumulator
}

func (p *PostingsWriter) Close() error {
	//TODO implement me
	panic("implement me")
}

func (p *PostingsWriter) Init(termsOut store.IndexOutput, state *index.SegmentWriteState) error {
	//TODO implement me
	panic("implement me")
}

func (p *PostingsWriter) WriteTerm(term []byte, termsEnum index.TermsEnum, docsSeen index.Bits, norms index.NormsProducer) (*types.BlockTermState, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PostingsWriter) EncodeTerm(out store.DataOutput, fieldInfo *document.FieldInfo, state *types.BlockTermState, absolute bool) error {
	//TODO implement me
	panic("implement me")
}

func (p *PostingsWriter) NewTermState() (*types.BlockTermState, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PostingsWriter) StartTerm(norms index.NumericDocValues) error {
	//TODO implement me
	panic("implement me")
}

func (p *PostingsWriter) FinishTerm(state *types.BlockTermState) error {
	//TODO implement me
	panic("implement me")
}

func (p *PostingsWriter) StartDoc(docID, freq int) error {
	//TODO implement me
	panic("implement me")
}

func (p *PostingsWriter) AddPosition(position int, payload []byte, startOffset, endOffset int) error {
	//TODO implement me
	panic("implement me")
}

func (p *PostingsWriter) FinishDoc() error {
	//TODO implement me
	panic("implement me")
}
