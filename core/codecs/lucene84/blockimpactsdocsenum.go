package lucene84

import (
	"context"
	"io"
	"math"

	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
)

var _ index.ImpactsEnum = &BlockImpactsDocsEnum{}

type BlockImpactsDocsEnum struct {
	forUtil      *ForUtil
	forDeltaUtil *ForDeltaUtil
	pforUtil     *PForUtil

	docBuffer  []uint64
	freqBuffer []uint64

	docBufferUpto int

	skipper *ScoreSkipReader

	docIn store.IndexInput

	indexHasFreqs bool

	docFreq   int   // number of docs in this posting list
	blockUpto int   // number of documents in or before the current block
	doc       int   // doc we last read
	accum     int64 // accumulator for doc deltas

	nextSkipDoc int

	seekTo int64

	// as we read freqBuffer lazily, isFreqsRead shows if freqBuffer are read for the current block
	// always true when we don't have freqBuffer (indexHasFreq=false) or don't need freqBuffer (needsFreq=false)
	isFreqsRead bool
}

func (p *PostingsReader) NewBlockImpactsDocsEnum(ctx context.Context, fieldInfo *document.FieldInfo, termState *IntBlockTermState) (*BlockImpactsDocsEnum, error) {
	forUtil := NewForUtil()
	enum := &BlockImpactsDocsEnum{
		forUtil:      forUtil,
		forDeltaUtil: NewForDeltaUtil(forUtil),
		pforUtil:     NewFromForUtil(forUtil),
	}

	indexHasFreqs := fieldInfo.GetIndexOptions() >= document.INDEX_OPTIONS_DOCS_AND_FREQS
	indexHasPositions := fieldInfo.GetIndexOptions() >= document.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS
	indexHasOffsets := fieldInfo.GetIndexOptions() >= document.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS
	indexHasPayloads := fieldInfo.HasPayloads()

	enum.docIn = p.docIn.Clone().(store.IndexInput)

	enum.docFreq = termState.DocFreq
	if _, err := enum.docIn.Seek(termState.DocStartFP, io.SeekStart); err != nil {
		return nil, err
	}

	enum.doc = -1
	enum.accum = 0
	enum.blockUpto = 0
	enum.docBufferUpto = BLOCK_SIZE

	enum.skipper = NewScoreSkipReader(enum.docIn.Clone().(store.IndexInput),
		MAX_SKIP_LEVELS,
		indexHasPositions,
		indexHasOffsets,
		indexHasPayloads)
	if err := enum.skipper.Init(ctx, int(termState.DocStartFP+termState.SkipOffset),
		int(termState.DocStartFP), int(termState.PosStartFP), int(termState.PayStartFP), enum.docFreq); err != nil {
		return nil, err
	}

	// We set the last element of docBuffer to NO_MORE_DOCS, it helps save conditionals in advance()
	enum.docBuffer[BLOCK_SIZE] = math.MaxInt32
	enum.isFreqsRead = true
	if indexHasFreqs == false {
		for i := range enum.freqBuffer {
			enum.freqBuffer[i] = 1
		}
	}
	return enum, nil
}

func (b *BlockImpactsDocsEnum) DocID() int {
	return b.doc
}

func (b *BlockImpactsDocsEnum) NextDoc(ctx context.Context) (int, error) {
	return b.Advance(ctx, b.doc+1)
}

func (b *BlockImpactsDocsEnum) Advance(ctx context.Context, target int) (int, error) {
	if target > b.nextSkipDoc {
		if err := b.AdvanceShallow(ctx, target); err != nil {
			return 0, err
		}
	}
	if b.docBufferUpto == BLOCK_SIZE {
		if b.seekTo >= 0 {
			if _, err := b.docIn.Seek(b.seekTo, io.SeekStart); err != nil {
				return 0, err
			}
			b.isFreqsRead = true // reset isFreqsRead
			b.seekTo = -1
		}
		if err := b.refillDocs(ctx); err != nil {
			return 0, err
		}
	}

	next := findFirstGreater(b.docBuffer, target, b.docBufferUpto)
	b.doc = int(b.docBuffer[next])
	b.docBufferUpto = int(next + 1)
	return b.doc, nil
}

func findFirstGreater(buffer []uint64, target, from int) int {
	for i := from; i < BLOCK_SIZE; i++ {
		if buffer[i] >= uint64(target) {
			return i
		}
	}
	return BLOCK_SIZE
}

func (b *BlockImpactsDocsEnum) refillDocs(ctx context.Context) error {
	// Check if we skipped reading the previous block of freqBuffer, and if yes, position docIn after it
	if b.isFreqsRead == false {
		if err := b.pforUtil.Skip(ctx, b.docIn); err != nil {
			return err
		}
		b.isFreqsRead = true
	}

	left := b.docFreq - b.blockUpto

	if left >= BLOCK_SIZE {
		if err := b.forDeltaUtil.DecodeAndPrefixSum(ctx, b.docIn, uint64(b.accum), b.docBuffer); err != nil {
			return err
		}
		if b.indexHasFreqs {
			if err := b.pforUtil.Decode(ctx, b.docIn, b.freqBuffer); err != nil {
				return err
			}
		}
		b.blockUpto += BLOCK_SIZE
	} else {
		if err := readVIntBlock(ctx, b.docIn, b.docBuffer, b.freqBuffer, left, b.indexHasFreqs); err != nil {
			return err
		}
		prefixSum(b.docBuffer, left, uint64(b.accum))
		b.docBuffer[left] = types.NO_MORE_DOCS
		b.blockUpto += left
	}
	b.accum = int64(b.docBuffer[BLOCK_SIZE-1])
	b.docBufferUpto = 0
	return nil
}

func (b *BlockImpactsDocsEnum) SlowAdvance(ctx context.Context, target int) (int, error) {
	return types.SlowAdvanceWithContext(ctx, b, target)
}

func (b *BlockImpactsDocsEnum) Cost() int64 {
	return int64(b.docFreq)
}

func (b *BlockImpactsDocsEnum) Freq() (int, error) {
	if b.isFreqsRead == false {
		if err := b.pforUtil.Decode(context.Background(), b.docIn, b.freqBuffer); err != nil { // read freqBuffer for this block
			return 0, err
		}
		b.isFreqsRead = true
	}
	return int(b.freqBuffer[b.docBufferUpto-1]), nil
}

func (b *BlockImpactsDocsEnum) NextPosition() (int, error) {
	return -1, nil
}

func (b *BlockImpactsDocsEnum) StartOffset() (int, error) {
	return -1, nil
}

func (b *BlockImpactsDocsEnum) EndOffset() (int, error) {
	return -1, nil
}

func (b *BlockImpactsDocsEnum) GetPayload() ([]byte, error) {
	return []byte{}, nil
}

func (b *BlockImpactsDocsEnum) AdvanceShallow(ctx context.Context, target int) error {
	if target > b.nextSkipDoc {
		// always plus one to fix the result, since skip position in Lucene84SkipReader
		// is a little different from MultiLevelSkipListReader
		upto, err := b.skipper.SkipTo(ctx, target)
		if err != nil {
			return err
		}
		newDocUpto := upto + 1

		if newDocUpto >= b.blockUpto {
			// Skipper moved
			b.blockUpto = newDocUpto

			// Force to read next block
			b.docBufferUpto = BLOCK_SIZE
			b.accum = int64(b.skipper.GetDoc())
			b.seekTo = b.skipper.GetDocPointer() // delay the seek
		}
		// next time we call advance, this is used to
		// foresee whether skipper is necessary.
		b.nextSkipDoc = b.skipper.GetNextSkipDoc()
	}
	return nil
}

func (b *BlockImpactsDocsEnum) GetImpacts() (index.Impacts, error) {
	if err := b.AdvanceShallow(context.Background(), b.doc); err != nil {
		return nil, err
	}
	return b.skipper.GetImpacts()
}
