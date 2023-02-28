package index

import (
	"encoding/binary"
	"fmt"
	"github.com/geange/lucene-go/core/tokenattributes"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util"
	"sort"
	"strings"
)

var _ TermsHashPerField = &TermVectorsConsumerPerField{}

type TermVectorsConsumerPerField struct {
	*TermsHashPerFieldDefault

	termVectorsPostingsArray *TermVectorsPostingsArray

	termsWriter       *TermVectorsConsumer
	fieldState        *FieldInvertState
	fieldInfo         *types.FieldInfo
	doVectors         bool
	doVectorPositions bool
	doVectorOffsets   bool
	doVectorPayloads  bool
	offsetAttribute   tokenattributes.OffsetAttribute
	payloadAttribute  tokenattributes.PayloadAttribute
	termFreqAtt       tokenattributes.TermFrequencyAttribute
	termBytePool      *util.ByteBlockPool
	hasPayloads       bool // if enabled, and we actually saw any for this field
}

func NewTermVectorsConsumerPerField(invertState *FieldInvertState,
	termsHash *TermVectorsConsumer, fieldInfo *types.FieldInfo) *TermVectorsConsumerPerField {

	perfield := &TermVectorsConsumerPerField{
		TermsHashPerFieldDefault: &TermsHashPerFieldDefault{
			nextPerField:            nil,
			intPool:                 termsHash.intPool,
			bytePool:                termsHash.bytePool,
			termStreamAddressBuffer: nil,
			streamAddressOffset:     0,
			streamCount:             2,
			fieldName:               fieldInfo.Name(),
			indexOptions:            fieldInfo.GetIndexOptions(),
			bytesHash:               nil,
			postingsArray:           nil,
			lastDocID:               0,
			sortedTermIDs:           nil,
			doNextCall:              false,
			fnNewTerm:               nil,
			fnAddTerm:               nil,
		},
		termsWriter:  termsHash,
		fieldState:   invertState,
		fieldInfo:    fieldInfo,
		termBytePool: termsHash.termBytePool,
	}

	byteStarts := NewPostingsBytesStartArray(perfield)
	HASH_INIT_SIZE := 4
	perfield.bytesHash = util.NewBytesRefHashV1(termsHash.termBytePool, HASH_INIT_SIZE, byteStarts)
	perfield.fnNewTerm = perfield.NewTerm
	perfield.fnAddTerm = perfield.AddTerm
	return perfield
}

func (t *TermVectorsConsumerPerField) NewTerm(termID, docID int) error {
	postings := t.termVectorsPostingsArray

	freq, err := t.getTermFreq()
	if err != nil {
		return err
	}
	postings.SetFreqs(termID, freq)
	postings.SetLastOffsets(termID, 0)
	postings.SetLastPositions(termID, 0)
	return t.writeProx(postings, termID)
}

func (t *TermVectorsConsumerPerField) AddTerm(termID, docID int) error {
	postings := t.termVectorsPostingsArray

	freq, err := t.getTermFreq()
	if err != nil {
		return err
	}
	postings.SetFreqs(termID, postings.freqs[termID]+freq)
	return t.writeProx(postings, termID)
}

func (t *TermVectorsConsumerPerField) Finish() error {
	if !t.doVectors || t.getNumTerms() == 0 {
		return nil
	}
	return t.termsWriter.addFieldToFlush(t)
}

func (t *TermVectorsConsumerPerField) FinishDocument() error {
	panic("")
}

func (t *TermVectorsConsumerPerField) NewPostingsArray() {
	t.termVectorsPostingsArray = t.postingsArray.(*TermVectorsPostingsArray)
}

func (t *TermVectorsConsumerPerField) CreatePostingsArray(size int) ParallelPostingsArray {
	return NewTermVectorsPostingsArray()
}

func (t *TermVectorsConsumerPerField) getFieldName() string {
	return t.fieldName
}

func (t *TermVectorsConsumerPerField) getTermFreq() (int, error) {
	freq := t.termFreqAtt.GetTermFrequency()
	if freq != 1 {
		if t.doVectorPositions {
			return 0, fmt.Errorf(
				"field %s: cannot index term vector positions while using custom TermFrequencyAttribute",
				t.getFieldName(),
			)
		}

		if t.doVectorOffsets {
			return 0, fmt.Errorf(
				"field %s: cannot index term vector offsets while using custom TermFrequencyAttribute",
				t.getFieldName(),
			)
		}
	}
	return freq, nil
}

func (t *TermVectorsConsumerPerField) writeProx(postings *TermVectorsPostingsArray, termID int) error {
	if t.doVectorOffsets {
		startOffset := t.fieldState.offset + t.offsetAttribute.StartOffset()
		endOffset := t.fieldState.offset + t.offsetAttribute.EndOffset()

		t.writeVInt(1, startOffset-postings.lastOffsets[termID])
		t.writeVInt(1, endOffset-startOffset)
		postings.SetLastOffsets(termID, endOffset)
	}

	if t.doVectorPositions {
		var payload []byte
		if t.payloadAttribute != nil {
			payload = t.payloadAttribute.GetPayload()
		}

		pos := t.fieldState.position - postings.lastPositions[termID]
		if len(payload) > 0 {
			t.writeVInt(0, (pos<<1)|1)
			t.writeVInt(0, len(payload))
			t.writeBytes(0, payload)
			t.hasPayloads = true
		} else {
			t.writeVInt(0, pos<<1)
		}
		postings.SetLastPositions(termID, t.fieldState.position)
	}
	return nil
}

func (t *TermVectorsConsumerPerField) writeBytes(stream int, bs []byte) {
	for _, b := range bs {
		t.writeByte(stream, b)
	}
}

func (t *TermVectorsConsumerPerField) writeVInt(stream, i int) {
	buf := make([]byte, 10)
	num := binary.PutUvarint(buf, uint64(i))

	for _, b := range buf[:num] {
		t.writeByte(stream, b)
	}
}

func (t *TermVectorsConsumerPerField) writeByte(stream int, b byte) {
	streamAddress := t.streamAddressOffset + stream
	upto := t.termStreamAddressBuffer[streamAddress]
	bytes := t.bytePool.Get(upto >> util.BYTE_BLOCK_SHIFT)
	offset := upto & util.BYTE_BLOCK_MASK
	if bytes[offset] != 0 {
		// End of slice; allocate a new one
		offset = t.bytePool.AllocSlice(bytes, offset)
		bytes = t.bytePool.Current()
		t.termStreamAddressBuffer[streamAddress] = offset + t.bytePool.ByteOffset
	}
	bytes[offset] = b
	t.termStreamAddressBuffer[streamAddress]++
}

func SortTermVectorsConsumerPerField(fields []*TermVectorsConsumerPerField) {
	sort.Sort(PerFields(fields))
}

var _ sort.Interface = PerFields{}

type PerFields []*TermVectorsConsumerPerField

func (p PerFields) Len() int {
	return len(p)
}

func (p PerFields) Less(i, j int) bool {
	return strings.Compare(p[i].fieldName, p[j].fieldName) < 0
}

func (p PerFields) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
