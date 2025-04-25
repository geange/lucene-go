package lucene84

import (
	"context"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
	"io"
)

var _ index.ImpactsEnum = &BlockImpactsPostingsEnum{}

type BlockImpactsPostingsEnum struct {
	forUtil      *ForUtil
	forDeltaUtil *ForDeltaUtil
	pforUtil     *PForUtil

	docBuffer      []uint64
	freqBuffer     []uint64
	posDeltaBuffer []uint64

	docBufferUpto int
	posBufferUpto int

	skipper *ScoreSkipReader

	docIn store.IndexInput
	posIn store.IndexInput

	indexHasOffsets  bool
	indexHasPayloads bool

	docFreq       int    // number of docs in this posting list
	totalTermFreq uint64 // number of positions in this posting list
	docUpto       int    // how many docs we've read
	doc           int    // doc we last read
	accum         uint64 // accumulator for doc deltas
	freq          uint64 // freq we last read
	position      int64  // current position

	// how many positions "behind" we are; nextPosition must
	// skip these to "catch up":
	posPendingCount int

	// Lazy pos seek: if != -1 then we must seek to this FP
	// before reading positions:
	posPendingFP uint64

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

	seekTo int64

	// as we read freqBuffer lazily, isFreqsRead shows if freqBuffer are read for the current block
	// always true when we don't have freqBuffer (indexHasFreq=false) or don't need freqBuffer (needsFreq=false)
	isFreqsRead bool
}

func NewBlockImpactsDocsEnum( fieldInfo *document.FieldInfo,  termState *IntBlockTermState) *BlockImpactsDocsEnum {
	
}

func (b *BlockImpactsPostingsEnum) DocID() int {
	return b.doc
}

func (b *BlockImpactsPostingsEnum) NextDoc(ctx context.Context) (int, error) {
	return b.Advance(ctx, b.doc+1)
}

func (b *BlockImpactsPostingsEnum) Advance(ctx context.Context, target int) (int, error) {
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
	b.docBufferUpto = next + 1
	return b.doc, nil
}

func (b *BlockImpactsPostingsEnum) SlowAdvance(ctx context.Context, target int) (int, error) {
	return types.SlowAdvanceWithContext(ctx, b, target)
}

func (b *BlockImpactsPostingsEnum) Cost() int64 {
	return int64(b.docFreq)
}

func (b *BlockImpactsPostingsEnum) Freq() (int, error) {
	return b.docFreq, nil
}

func (b *BlockImpactsPostingsEnum) NextPosition() (int, error) {
	return -1, nil
}

func (b *BlockImpactsPostingsEnum) StartOffset() (int, error) {
	return -1, nil
}

func (b *BlockImpactsPostingsEnum) EndOffset() (int, error) {
	return -1, nil
}

func (b *BlockImpactsPostingsEnum) GetPayload() ([]byte, error) {
	return []byte{}, nil
}

func (b *BlockImpactsPostingsEnum) AdvanceShallow(ctx context.Context, target int) error {
	if target > b.nextSkipDoc {
		// always plus one to fix the result, since skip position in Lucene84SkipReader
		// is a little different from MultiLevelSkipListReader
		off, err := b.skipper.SkipTo(ctx, target)
		if err != nil {
			return err
		}
		newDocUpto := off + 1

		if newDocUpto > b.docUpto {
			// Skipper moved
			b.docUpto = newDocUpto

			// Force to read next block
			b.docBufferUpto = BLOCK_SIZE
			b.accum = uint64(b.skipper.GetDoc())
			b.posPendingFP = uint64(b.skipper.GetPosPointer())
			b.posPendingCount = b.skipper.GetPosBufferUpto()
			b.seekTo = b.skipper.GetDocPointer() // delay the seek
		}
		// next time we call advance, this is used to
		// foresee whether skipper is necessary.
		b.nextSkipDoc = b.skipper.GetNextSkipDoc()
	}
	return nil
}

func (b *BlockImpactsPostingsEnum) GetImpacts() (index.Impacts, error) {
	if err := b.AdvanceShallow(context.Background(), b.doc); err != nil {
		return nil, err
	}
	return b.skipper.GetImpacts()
}

func (b *BlockImpactsPostingsEnum) refillDocs(ctx context.Context) error {
	left := b.docFreq - b.docUpto

	if left >= BLOCK_SIZE {
		if err := b.forDeltaUtil.DecodeAndPrefixSum(ctx, b.docIn, b.accum, b.docBuffer); err != nil {
			return err
		}
		if err := b.pforUtil.Decode(ctx, b.docIn, b.freqBuffer); err != nil {
			return err
		}
	} else {
		if err := readVIntBlock(ctx, b.docIn, b.docBuffer, b.freqBuffer, left, true); err != nil {
			return err
		}
		prefixSum(b.docBuffer, left, b.accum)
		b.docBuffer[left] = types.NO_MORE_DOCS
	}
	b.accum = b.docBuffer[BLOCK_SIZE-1]
	b.docBufferUpto = 0
	return nil
}

func (b *BlockImpactsPostingsEnum) refillPositions(ctx context.Context) error {
	if b.posIn.GetFilePointer() == int64(b.lastPosBlockFP) {
		count := int(b.totalTermFreq % BLOCK_SIZE)
		payloadLength := 0
		for i := 0; i < count; i++ {
			code, err := b.posIn.ReadUvarint(ctx)
			if err != nil {
				return err
			}
			if b.indexHasPayloads {
				if (code & 1) != 0 {
					length, err := b.posIn.ReadUvarint(ctx)
					if err != nil {
						return err
					}
					payloadLength = int(length)
				}
				b.posDeltaBuffer[i] = code >> 1
				if payloadLength != 0 {
					if _, err := b.posIn.Seek(b.posIn.GetFilePointer()+int64(payloadLength), io.SeekStart); err != nil {
						return err
					}
				}
			} else {
				b.posDeltaBuffer[i] = code
			}
			if b.indexHasOffsets {
				offfset, err := b.posIn.ReadUvarint(ctx)
				if err != nil {
					return err
				}
				if (offfset & 1) != 0 {
					// offset length changed
					if _, err := b.posIn.ReadUvarint(ctx); err != nil {
						return err
					}
				}
			}
		}
	} else {
		if err := b.pforUtil.Decode(ctx, b.posIn, b.posDeltaBuffer); err != nil {
			return err
		}
	}
	return nil
}
