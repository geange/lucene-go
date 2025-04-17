package lucene84

import (
	"context"
	"io"
	"math"
	"slices"

	"github.com/geange/lucene-go/core/codecs/types"
	"github.com/geange/lucene-go/core/document"
	coreIndex "github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	types2 "github.com/geange/lucene-go/core/types"
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
	payloadByteUpto uint64
	payloadLength   uint64

	lastStartOffset int64
	startOffset     int64
	endOffset       int64

	docBufferUpto int
	posBufferUpto uint64

	skipper *SkipReader
	skipped bool

	startDocIn store.IndexInput

	docIn   store.IndexInput
	posIn   store.IndexInput
	payIn   store.IndexInput
	payload []byte

	indexHasOffsets  bool
	indexHasPayloads bool

	docFreq       int    // number of docs in this posting list
	totalTermFreq uint64 // number of positions in this posting list
	blockUpto     int    // number of docs in or before the current block
	doc           int64  // doc we last read
	accum         uint64 // accumulator for doc deltas
	freq          uint64 // freq we last read
	position      uint64 // current position

	// how many positions "behind" we are; nextPosition must
	// skip these to "catch up":
	posPendingCount uint64

	// Lazy pos seek: if != -1 then we must seek to this FP
	// before reading positions:
	posPendingFP int64

	// Lazy pay seek: if != -1 then we must seek to this FP
	// before reading payloads/offsets:
	payPendingFP int64

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
	lastPosBlockFP int64

	// Where this term's skip data starts (after
	// docTermStartFP) in the .doc file (or -1 if there is
	// no skip data for this term):
	skipOffset int

	nextSkipDoc int

	needsOffsets   bool   // true if we actually need offsets
	needsPayloads  bool   // true if we actually need payloads
	singletonDocID uint64 // docid when there is a single pulsed posting, otherwise -1
}

func (p *PostingsReader) NewEverythingEnum(fieldInfo *document.FieldInfo) (*EverythingEnum, error) {

	indexHasOffsets := fieldInfo.GetIndexOptions() >= document.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS
	indexHasPayloads := fieldInfo.HasPayloads()
	this := &EverythingEnum{
		indexHasOffsets:  indexHasOffsets,
		indexHasPayloads: indexHasPayloads,
	}

	this.startDocIn = p.docIn
	this.docIn = nil
	this.posIn = p.posIn.Clone().(store.IndexInput)
	if indexHasOffsets || indexHasPayloads {
		this.payIn = p.payIn.Clone().(store.IndexInput)
	} else {
		this.payIn = nil
	}

	if indexHasOffsets {
		this.offsetStartDeltaBuffer = make([]uint64, BLOCK_SIZE)
		this.offsetLengthBuffer = make([]uint64, BLOCK_SIZE)
	} else {
		this.offsetStartDeltaBuffer = nil
		this.offsetLengthBuffer = nil
		this.startOffset = -1
		this.endOffset = -1
	}

	if indexHasPayloads {
		this.payloadLengthBuffer = make([]uint64, BLOCK_SIZE)
		this.payloadBytes = make([]byte, 128)
		this.payload = make([]byte, 0)
	} else {
		this.payloadLengthBuffer = nil
		this.payloadBytes = nil
		this.payload = nil
	}
	// We set the last element of docBuffer to NO_MORE_DOCS, it helps save conditionals in advance()
	this.docBuffer[BLOCK_SIZE] = math.MaxInt32

	return this, nil
}

func (e *EverythingEnum) reset(termState *IntBlockTermState, flags int) (*EverythingEnum, error) {
	e.docFreq = termState.DocFreq
	e.docTermStartFP = int(termState.DocStartFP)
	e.posTermStartFP = int(termState.PosStartFP)
	e.payTermStartFP = int(termState.PayStartFP)
	e.skipOffset = int(termState.SkipOffset)
	e.totalTermFreq = uint64(termState.TotalTermFreq)
	e.singletonDocID = uint64(termState.SingletonDocID)
	if e.docFreq > 1 {
		if e.docIn == nil {
			// lazy init
			e.docIn = e.startDocIn.Clone().(store.IndexInput)
		}
		if _, err := e.docIn.Seek(int64(e.docTermStartFP), io.SeekStart); err != nil {
			return nil, err
		}
	}
	e.posPendingFP = int64(e.posTermStartFP)
	e.payPendingFP = int64(e.payTermStartFP)
	e.posPendingCount = 0
	if termState.TotalTermFreq < BLOCK_SIZE {
		e.lastPosBlockFP = int64(e.posTermStartFP)
	} else if termState.TotalTermFreq == BLOCK_SIZE {
		e.lastPosBlockFP = -1
	} else {
		e.lastPosBlockFP = int64(e.posTermStartFP) + termState.LastPosBlockOffset
	}

	e.needsOffsets = featureRequested(flags, coreIndex.POSTINGS_ENUM_OFFSETS)
	e.needsPayloads = featureRequested(flags, coreIndex.POSTINGS_ENUM_PAYLOADS)

	e.doc = -1
	e.accum = 0
	e.blockUpto = 0
	if e.docFreq > BLOCK_SIZE {
		e.nextSkipDoc = BLOCK_SIZE - 1 // we won't skip if target is found in first block
	} else {
		// TODO:
		//		  e. nextSkipDoc = NO_MORE_DOCS; // not enough docs for skipping
	}
	e.docBufferUpto = BLOCK_SIZE
	e.skipped = false
	return e, nil
}

func featureRequested(flags int, feature int) bool {
	return (flags & feature) == feature
}

func (e *EverythingEnum) canReuse(docIn store.IndexInput, fieldInfo *document.FieldInfo) bool {
	return docIn == e.startDocIn &&
		e.indexHasOffsets == (fieldInfo.GetIndexOptions() >= document.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS) &&
		e.indexHasPayloads == fieldInfo.HasPayloads()
}

func (e *EverythingEnum) refillDocs(ctx context.Context) error {
	left := e.docFreq - e.blockUpto

	if left >= BLOCK_SIZE {
		if err := e.forDeltaUtil.DecodeAndPrefixSum(ctx, e.docIn, e.accum, e.docBuffer); err != nil {
			return err
		}
		if err := e.pforUtil.Decode(ctx, e.docIn, e.freqBuffer); err != nil {
			return err
		}
		e.blockUpto += BLOCK_SIZE
	} else if e.docFreq == 1 {
		e.docBuffer[0] = e.singletonDocID
		e.freqBuffer[0] = e.totalTermFreq
		//e.docBuffer[1] = NO_MORE_DOCS
		e.blockUpto++
	} else {
		if err := readVIntBlock(ctx, e.docIn, e.docBuffer, e.freqBuffer, left, true); err != nil {
			return err
		}
		prefixSum(e.docBuffer, left, e.accum)
		//e.docBuffer[left] = NO_MORE_DOCS
		e.blockUpto += left
	}
	e.accum = e.docBuffer[BLOCK_SIZE-1]
	e.docBufferUpto = 0
	return nil
}

// TODO: in theory we could avoid loading frq block
// when not needed, ie, use skip data to load how far to
// seek the pos pointer ... instead of having to load frq
// blocks only to sum up how many positions to skip
func (e *EverythingEnum) skipPositions(ctx context.Context) error {
	// Skip positions now:
	toSkip := e.posPendingCount - e.freq
	// if (DEBUG) {
	//   System.out.println("      FPR.skipPositions: toSkip=" + toSkip);
	// }

	leftInBlock := uint64(BLOCK_SIZE - e.posBufferUpto)
	if toSkip < leftInBlock {
		end := e.posBufferUpto + toSkip
		for e.posBufferUpto < end {
			if e.indexHasPayloads {
				e.payloadByteUpto += e.payloadLengthBuffer[e.posBufferUpto]
			}
			e.posBufferUpto++
		}
	} else {
		toSkip -= leftInBlock
		for toSkip >= BLOCK_SIZE {
			if err := e.pforUtil.Skip(ctx, e.posIn); err != nil {
				return err
			}

			if e.indexHasPayloads {
				// Skip payloadLength block:
				if err := e.pforUtil.Skip(ctx, e.payIn); err != nil {
					return err
				}

				// Skip payloadBytes block:
				nBytes, err := e.payIn.ReadUvarint(ctx)
				if err != nil {
					return err
				}
				if _, err := e.payIn.Seek(e.payIn.GetFilePointer()+int64(nBytes), io.SeekStart); err != nil {
					return nil
				}
			}

			if e.indexHasOffsets {
				if err := e.pforUtil.Skip(ctx, e.payIn); err != nil {
					return err
				}
				if err := e.pforUtil.Skip(ctx, e.payIn); err != nil {
					return err
				}
			}
			toSkip -= BLOCK_SIZE
		}
		if err := e.refillPositions(ctx); err != nil {
			return err
		}
		e.payloadByteUpto = 0
		e.posBufferUpto = 0
		for e.posBufferUpto < toSkip {
			if e.indexHasPayloads {
				e.payloadByteUpto += e.payloadLengthBuffer[e.posBufferUpto]
			}
			e.posBufferUpto++
		}
	}

	e.position = 0
	e.lastStartOffset = 0
	return nil
}

func (e *EverythingEnum) refillPositions(ctx context.Context) error {
	if e.posIn.GetFilePointer() == e.lastPosBlockFP {
		count := e.totalTermFreq % BLOCK_SIZE
		payloadLength := uint64(0)
		offsetLength := uint64(0)
		payloadByteUpto := uint64(0)
		for i := 0; i < int(count); i++ {
			code, err := e.posIn.ReadUvarint(ctx)
			if err != nil {
				return err
			}
			if e.indexHasPayloads {
				if (code & 1) != 0 {
					payloadLength, err = e.posIn.ReadUvarint(ctx)
					if err != nil {
						return err
					}
				}
				e.payloadLengthBuffer[i] = payloadLength
				e.posDeltaBuffer[i] = code >> 1
				if payloadLength != 0 {
					size := int(payloadByteUpto + payloadLength)
					slices.Grow(e.payloadBytes, size)

					if _, err := e.posIn.Read(e.payloadBytes[payloadByteUpto : payloadByteUpto+payloadLength]); err != nil {
						return err
					}
					payloadByteUpto += payloadLength
				}
			} else {
				e.posDeltaBuffer[i] = code
			}

			if e.indexHasOffsets {
				deltaCode, err := e.posIn.ReadUvarint(ctx)
				if err != nil {
					return err
				}
				if (deltaCode & 1) != 0 {
					offsetLength, err = e.posIn.ReadUvarint(ctx)
					if err != nil {
						return err
					}
				}
				e.offsetStartDeltaBuffer[i] = deltaCode >> 1
				e.offsetLengthBuffer[i] = offsetLength
			}
		}
		payloadByteUpto = 0
		return nil
	}

	if err := e.pforUtil.Decode(ctx, e.posIn, e.posDeltaBuffer); err != nil {
		return err
	}

	if e.indexHasPayloads {
		if e.needsPayloads {
			if err := e.pforUtil.Decode(ctx, e.payIn, e.payloadLengthBuffer); err != nil {
				return err
			}
			nBytes, err := e.payIn.ReadUvarint(ctx)
			if err != nil {
				return err
			}
			slices.Grow(e.payloadBytes, int(nBytes))
			if _, err := e.payIn.Read(e.payloadBytes[:nBytes]); err != nil {
				return nil
			}
		} else {
			// this works, because when writing a vint block we always force the first length to be written
			if err := e.pforUtil.Skip(ctx, e.payIn); err != nil { // skip over lengths
				return err
			}
			nBytes, err := e.payIn.ReadUvarint(ctx) // read length of payloadBytes
			if err != nil {
				return err
			}
			if _, err := e.payIn.Seek(e.payIn.GetFilePointer()+int64(nBytes), io.SeekStart); err != nil { // skip over payloadBytes
				return err
			}
		}
		e.payloadByteUpto = 0
	}

	if e.indexHasOffsets {
		if e.needsOffsets {
			if err := e.pforUtil.Decode(ctx, e.payIn, e.offsetStartDeltaBuffer); err != nil {
				return err
			}
			if err := e.pforUtil.Decode(ctx, e.payIn, e.offsetLengthBuffer); err != nil {
				return err
			}
		} else {
			// this works, because when writing a vint block we always force the first length to be written
			if err := e.pforUtil.Skip(ctx, e.payIn); err != nil { // skip over starts
				return err
			}
			if err := e.pforUtil.Skip(ctx, e.payIn); err != nil { // skip over lengths
				return err
			}
		}
	}
	return nil
}

// Read values that have been written using variable-length encoding instead of bit-packing.
func readVIntBlock(ctx context.Context, docIn store.IndexInput, docBuffer, freqBuffer []uint64, num int, indexHasFreq bool) error {
	if indexHasFreq {
		for i := 0; i < num; i++ {
			code, err := docIn.ReadUvarint(ctx)
			if err != nil {
				return err
			}
			docBuffer[i] = code >> 1
			if (code & 1) != 0 {
				freqBuffer[i] = 1
			} else {
				freq, err := docIn.ReadUvarint(ctx)
				if err != nil {
					return err
				}
				freqBuffer[i] = freq
			}
		}
		return nil
	}

	for i := 0; i < num; i++ {
		doc, err := docIn.ReadUvarint(ctx)
		if err != nil {
			return err
		}
		docBuffer[i] = doc
	}
	return nil
}

func prefixSum(buffer []uint64, count int, base uint64) {
	buffer[0] += base
	for i := 1; i < count; i++ {
		buffer[i] += buffer[i-1]
	}
}

func (e *EverythingEnum) DocID() int {
	return int(e.doc)
}

func (e *EverythingEnum) NextDoc(ctx context.Context) (int, error) {
	if e.docBufferUpto == BLOCK_SIZE {
		if err := e.refillDocs(ctx); err != nil {
			return 0, err
		}
	}

	e.doc = int64(e.docBuffer[e.docBufferUpto])
	e.freq = e.freqBuffer[e.docBufferUpto]
	e.posPendingCount += e.freq
	e.docBufferUpto++

	e.position = 0
	e.lastStartOffset = 0
	return int(e.doc), nil
}

func (e *EverythingEnum) Advance(ctx context.Context, target int) (int, error) {
	if target > e.nextSkipDoc {
		if e.skipper == nil {
			// Lazy init: first time this enum has ever been used for skipping
			var err error
			e.skipper, err = NewSkipReader(e.docIn.Clone().(store.IndexInput),
				MAX_SKIP_LEVELS, true, e.indexHasOffsets, e.indexHasPayloads)
			if err != nil {
				return 0, err
			}
		}

		if !e.skipped {
			// This is the first time this enum has skipped
			// since reset() was called; load the skip data:
			e.skipper.Init(ctx, e.docTermStartFP+e.skipOffset, e.docTermStartFP, e.posTermStartFP, e.payTermStartFP, e.docFreq)
			e.skipped = true
		}

		skipTo, err := e.skipper.SkipTo(ctx, target)
		if err != nil {
			return 0, err
		}
		newDocUpto := skipTo + 1

		if newDocUpto > e.blockUpto-BLOCK_SIZE+e.docBufferUpto {
			// Skipper moved
			e.blockUpto = newDocUpto

			// Force to read next block
			e.docBufferUpto = BLOCK_SIZE
			e.accum = uint64(e.skipper.GetDoc())
			e.docIn.Seek(e.skipper.GetDocPointer(), io.SeekStart)
			e.posPendingFP = e.skipper.GetPosPointer()
			e.payPendingFP = e.skipper.GetPayPointer()
			e.posPendingCount = uint64(e.skipper.GetPosBufferUpto())
			e.lastStartOffset = 0 // new document
			e.payloadByteUpto = uint64(e.skipper.GetPayloadByteUpto())
		}
		e.nextSkipDoc = e.skipper.GetNextSkipDoc()
	}
	if e.docBufferUpto == BLOCK_SIZE {
		e.refillDocs(ctx)
	}

	// Now scan:
	var doc uint64
	for {
		doc = e.docBuffer[e.docBufferUpto]
		e.freq = e.freqBuffer[e.docBufferUpto]
		e.posPendingCount += e.freq
		e.docBufferUpto++

		if int(doc) >= target {
			break
		}
	}

	e.position = 0
	e.lastStartOffset = 0
	e.doc = int64(doc)
	return int(doc), nil
}

func (e *EverythingEnum) SlowAdvance(ctx context.Context, target int) (int, error) {
	return types2.SlowAdvanceWithContext(ctx, e, target)
}

func (e *EverythingEnum) Cost() int64 {
	return int64(e.docFreq)
}

func (e *EverythingEnum) Freq() (int, error) {
	return int(e.freq), nil
}

func (e *EverythingEnum) NextPosition() (int, error) {
	if e.posPendingFP != -1 {
		if _, err := e.posIn.Seek(e.posPendingFP, io.SeekStart); err != nil {
			return 0, err
		}
		e.posPendingFP = -1

		if e.payPendingFP != -1 && e.payIn != nil {
			if _, err := e.payIn.Seek(e.payPendingFP, io.SeekStart); err != nil {
				return 0, err
			}
			e.payPendingFP = -1
		}

		// Force buffer refill:
		e.posBufferUpto = BLOCK_SIZE
	}

	// FIXME
	ctx := context.Background()

	if e.posPendingCount > (e.freq) {
		if err := e.skipPositions(ctx); err != nil {
			return 0, err
		}
		e.posPendingCount = e.freq
	}

	if e.posBufferUpto == BLOCK_SIZE {
		if err := e.refillPositions(ctx); err != nil {
			return 0, err
		}
		e.posBufferUpto = 0
	}
	e.position += e.posDeltaBuffer[e.posBufferUpto]

	if e.indexHasPayloads {
		e.payloadLength = e.payloadLengthBuffer[e.posBufferUpto]
		e.payload = e.payloadBytes[e.payloadByteUpto : e.payloadByteUpto+e.payloadLength]
		e.payloadByteUpto += e.payloadLength
	}

	if e.indexHasOffsets {
		e.startOffset = e.lastStartOffset + int64(e.offsetStartDeltaBuffer[e.posBufferUpto])
		e.endOffset = e.startOffset + int64(e.offsetLengthBuffer[e.posBufferUpto])
		e.lastStartOffset = e.startOffset
	}

	e.posBufferUpto++
	e.posPendingCount--
	return int(e.position), nil
}

func (e *EverythingEnum) StartOffset() (int, error) {
	return int(e.startOffset), nil
}

func (e *EverythingEnum) EndOffset() (int, error) {
	return int(e.endOffset), nil
}

func (e *EverythingEnum) GetPayload() ([]byte, error) {
	return e.payload, nil
}
