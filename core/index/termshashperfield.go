package index

import (
	"encoding/binary"

	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/util/bytesref"
	"github.com/geange/lucene-go/core/util/ints"
)

const HASH_INIT_SIZE = 4

// TermsHashPerField This class stores streams of information per term without knowing the size of the
// stream ahead of time. Each stream typically encodes one level of information like term frequency
// per document or term proximity. Internally this class allocates a linked list of slices that can
// be read by a ByteSliceReader for each term. Terms are first deduplicated in a BytesHash once
// this is done internal data-structures point to the current offset of each stream that can be written to.
//
// TermsHashPerField 可以存储每个term的信息，而不需要知道数据流的大小。
// 每个数据流通常编码一个级别的信息，如每个文档的术语频率（term's frequency）或术语接近度（term's proximity）。
// 在内部，此类为每个术语（term）分配 ByteSliceReader 可以读取的切片的链接列表。
// 完成后，首先在BytesRefHash中对术语（term）进行重复数据消除。内部数据结构指向可以写入的每个流的当前偏移量。
type TermsHashPerField interface {
	// Start adding a new field instance; first is true if this is the first time this field name was seen in the document.
	Start(field document.IndexableField, first bool) bool

	Add(termBytes []byte, docID int) error

	Add2nd(textStart, docID int) error

	GetNextPerField() TermsHashPerField

	Finish() error

	Reset() error

	// NewTerm Called when a term is seen for the first time.
	NewTerm(termID, docID int) error

	// AddTerm Called when a previously seen term is seen again.
	AddTerm(termID, docID int) error

	// NewPostingsArray Called when the postings array is initialized or resized.
	NewPostingsArray()

	// CreatePostingsArray Creates a new postings array of the specified size.
	CreatePostingsArray(size int) ParallelPostingsArray

	GetPostingsArray() ParallelPostingsArray

	SetPostingsArray(v ParallelPostingsArray)
}

type baseTermsHashPerField struct {
	nextPerField TermsHashPerField
	intPool      *ints.BlockPool
	bytePool     *bytesref.BlockPool

	// for each term we store an integer per stream that points into the bytePool above
	// the address is updated once data is written to the stream to point to the next free offset
	// in the terms stream. The start address for the stream is stored in postingsArray.byteStarts[termId]
	// This is initialized in the #addTerm method, either to a brand new per term stream if the term is new or
	// to the addresses where the term stream was written to when we saw it the last time.
	termStreamAddressBuffer []int

	streamAddressOffset int

	streamCount int

	fieldName string

	indexOptions document.IndexOptions

	// This stores the actual term bytes for postings and offsets into the parent hash in the case that this
	// TermsHashPerField is hashing term vectors.
	bytesHash *bytesref.BytesHash

	postingsArray ParallelPostingsArray

	lastDocID int

	sortedTermIDs []int

	doNextCall bool

	fnNewTerm func(termID, docID int) error // Called when a term is seen for the first time.
	fnAddTerm func(termID, docID int) error // Called when a previously seen term is seen again.
	// Called when the postings array is initialized or resized.
}

func newBaseTermsHashPerField(streamCount int,
	intPool *ints.BlockPool, bytePool, termBytePool *bytesref.BlockPool,
	nextPerField TermsHashPerField, fieldName string, indexOptions document.IndexOptions, perField TermsHashPerField) *baseTermsHashPerField {

	res := &baseTermsHashPerField{
		nextPerField:  nextPerField,
		intPool:       intPool,
		bytePool:      bytePool,
		streamCount:   streamCount,
		fieldName:     fieldName,
		indexOptions:  indexOptions,
		bytesHash:     nil,
		postingsArray: nil,
		lastDocID:     0,
		sortedTermIDs: nil,
		doNextCall:    false,
		fnNewTerm:     perField.NewTerm,
		fnAddTerm:     perField.AddTerm,
	}

	return res
}

func (t *baseTermsHashPerField) GetPostingsArray() ParallelPostingsArray {
	return t.postingsArray
}

func (t *baseTermsHashPerField) SetPostingsArray(v ParallelPostingsArray) {
	t.postingsArray = v
}

func (t *baseTermsHashPerField) Reset() error {
	t.bytesHash.Clear(false)
	t.sortedTermIDs = t.sortedTermIDs[:0]
	if t.nextPerField != nil {
		return t.nextPerField.Reset()
	}
	return nil
}

func (t *baseTermsHashPerField) initReader(reader *ByteSliceReader, termID, stream int) error {
	streamStartOffset := t.postingsArray.GetAddressOffset(termID)
	streamAddressBuffer := t.intPool.Get(streamStartOffset >> ints.INT_BLOCK_SHIFT)
	offsetInAddressBuffer := streamStartOffset & ints.INT_BLOCK_MASK

	startIndex := t.postingsArray.GetByteStarts(termID) + stream*bytesref.FIRST_LEVEL_SIZE
	endIndex := streamAddressBuffer[offsetInAddressBuffer+stream]

	return reader.init(t.bytePool, startIndex, endIndex)
}

// Collapse the hash table and sort in-place; also sets this.sortedTermIDs to the results This method must not be called twice unless reset() or reinitHash() was called.
func (t *baseTermsHashPerField) sortTerms() {
	t.sortedTermIDs = t.bytesHash.Sort()
}

// Returns the sorted term IDs. sortTerms() must be called before
func (t *baseTermsHashPerField) getSortedTermIDs() []int {
	return t.sortedTermIDs
}

func (t *baseTermsHashPerField) reinitHash() {
	t.bytesHash.ReInit()
}

// Secondary entry point (for 2nd & subsequent TermsHash),
// because token text has already been "interned" into
// textStart, so we hash by textStart.  term vectors use
// this API.
func (t *baseTermsHashPerField) add(textStart, docID int) error {
	termID := t.bytesHash.AddByPoolOffset(uint32(textStart))
	if termID >= 0 {
		// First time we are seeing this token since we last
		// flushed the hash.
		return t.initStreamSlices(termID, docID)
	}
	_, err := t.positionStreamSlice(termID, docID)
	return err
}

func (t *baseTermsHashPerField) initStreamSlices(termID, docID int) error {
	// Init stream slices
	if t.streamCount+t.intPool.IntUpto() > ints.INT_BLOCK_SIZE {
		// not enough space remaining in this buffer -- jump to next buffer and lose this remaining
		// piece
		t.intPool.NextBuffer()
	}

	if bytesref.BYTE_BLOCK_SIZE-t.bytePool.ByteUpto() < (2*t.streamCount)*bytesref.FIRST_LEVEL_SIZE {
		// can we fit at least one byte per stream in the current buffer, if not allocate a new one
		t.bytePool.NextBuffer()
	}

	t.termStreamAddressBuffer = t.intPool.Buffer()
	t.streamAddressOffset = t.intPool.IntUpto()
	t.intPool.AddIntUpto(t.streamCount) // advance the pool to reserve the N streams for this term

	t.postingsArray.SetAddressOffset(termID, t.streamAddressOffset+t.intPool.IntOffset())

	for i := 0; i < t.streamCount; i++ {
		// initialize each stream with a slice we start with ByteBlockPool.FIRST_LEVEL_SIZE)
		// and grow as we need more space. see ByteBlockPool.LEVEL_SIZE_ARRAY
		upto := t.bytePool.NewSlice(bytesref.FIRST_LEVEL_SIZE)
		t.termStreamAddressBuffer[t.streamAddressOffset+i] = upto + t.bytePool.ByteOffset()
	}
	t.postingsArray.SetByteStarts(termID, t.termStreamAddressBuffer[t.streamAddressOffset])
	return t.fnNewTerm(termID, docID)
}

func (t *baseTermsHashPerField) positionStreamSlice(termID, docID int) (int, error) {
	termID = (-termID) - 1
	intStart := t.postingsArray.GetAddressOffset(termID)
	t.termStreamAddressBuffer = t.intPool.Get(intStart >> ints.INT_BLOCK_SHIFT)
	t.streamAddressOffset = intStart & ints.INT_BLOCK_MASK
	if err := t.fnAddTerm(termID, docID); err != nil {
		return 0, err
	}
	return termID, nil
}

func (t *baseTermsHashPerField) getNumTerms() int {
	return t.bytesHash.Size()
}

func (t *baseTermsHashPerField) GetNextPerField() TermsHashPerField {
	return t.nextPerField
}

// Start adding a new field instance; first is true if this is the first time this field
// name was seen in the document.
func (t *baseTermsHashPerField) Start(field document.IndexableField, first bool) bool {
	if t.nextPerField != nil {
		t.doNextCall = t.nextPerField.Start(field, first)
	}
	return true
}

// Add2nd Secondary entry point (for 2nd & subsequent TermsHash),
// because token text has already been "interned" into
// textStart, so we hash by textStart.  term vectors use
// this API.
func (t *baseTermsHashPerField) Add2nd(textStart, docID int) error {
	termID := t.bytesHash.AddByPoolOffset(uint32(textStart))
	if termID >= 0 {
		// First time we are seeing this token since we last
		// flushed the hash.
		return t.initStreamSlices(termID, docID)
	}

	_, err := t.positionStreamSlice(termID, docID)
	return err
}

// Add Called once per inverted token. This is the primary entry point (for first TermsHash); postings use this API.
func (t *baseTermsHashPerField) Add(termBytes []byte, docID int) error {
	// We are first in the chain so we must "intern" the
	// term text into textStart address
	// Get the text & hash of this term.
	termID, err := t.bytesHash.Add(termBytes)
	if err != nil {
		return err
	}
	if termID >= 0 {
		if err = t.initStreamSlices(termID, docID); err != nil {
			return err
		}
	} else {
		termID, err = t.positionStreamSlice(termID, docID)
		if err != nil {
			return err
		}
	}
	if t.doNextCall {
		return t.nextPerField.Add2nd(termID, docID)
	}
	return nil
}

// Finish adding all instances of this field to the
// current document.
func (t *baseTermsHashPerField) Finish() error {
	if t.nextPerField != nil {
		return t.nextPerField.Finish()
	}
	return nil
}

func (t *baseTermsHashPerField) writeBytes(stream int, bs []byte) {
	for _, b := range bs {
		t.writeByte(stream, b)
	}
}

func (t *baseTermsHashPerField) writeVInt(stream, i int) {
	buf := make([]byte, 10)
	num := binary.PutUvarint(buf, uint64(i))

	for _, b := range buf[:num] {
		t.writeByte(stream, b)
	}
}

func (t *baseTermsHashPerField) writeByte(stream int, b byte) {
	streamAddress := t.streamAddressOffset + stream
	upto := t.termStreamAddressBuffer[streamAddress]
	bytes := t.bytePool.Get(upto >> bytesref.BYTE_BLOCK_SHIFT)
	offset := upto & bytesref.BYTE_BLOCK_MASK
	if bytes[offset] != 0 {
		// End of slice; allocate a new one
		offset = t.bytePool.AllocSlice(bytes, offset)
		bytes = t.bytePool.Current()
		t.termStreamAddressBuffer[streamAddress] = offset + t.bytePool.ByteOffset()
	}
	bytes[offset] = b
	t.termStreamAddressBuffer[streamAddress]++
}

var _ bytesref.StartArray = &PostingsBytesStartArray{}

type PostingsBytesStartArray struct {
	perField TermsHashPerField
}

func NewPostingsBytesStartArray(perField TermsHashPerField) *PostingsBytesStartArray {
	return &PostingsBytesStartArray{perField: perField}
}

func (p *PostingsBytesStartArray) Init() []uint32 {
	if p.perField.GetPostingsArray() == nil {
		p.perField.SetPostingsArray(p.perField.CreatePostingsArray(2))
		p.perField.NewPostingsArray()
	}
	return p.perField.GetPostingsArray().TextStarts()
}

func (p *PostingsBytesStartArray) Grow() []uint32 {
	p.perField.GetPostingsArray().Grow()
	p.perField.NewPostingsArray()
	return p.perField.GetPostingsArray().TextStarts()
}

func (p *PostingsBytesStartArray) Clear() []uint32 {
	if p.perField.GetPostingsArray() != nil {
		p.perField.SetPostingsArray(nil)
		p.perField.NewPostingsArray()
	}
	return nil
}
