package lucene84

import (
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
)

type PostingsWriter struct {
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

	fieldHasNorms bool
	norms         index.NumericDocValues
	//private final CompetitiveImpactAccumulator competitiveFreqNormAccumulator = new CompetitiveImpactAccumulator();
}
