package index

import (
	"fmt"
	"sort"
	"strings"

	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/util/attribute"
	"github.com/geange/lucene-go/core/util/bytesref"
)

var _ TermsHashPerField = &TermVectorsConsumerPerField{}

type TermVectorsConsumerPerField struct {
	*baseTermsHashPerField

	termVectorsPostingsArray *TermVectorsPostingsArray

	termsWriter       *TermVectorsConsumer
	fieldState        *FieldInvertState
	fieldInfo         *document.FieldInfo
	doVectors         bool
	doVectorPositions bool
	doVectorOffsets   bool
	doVectorPayloads  bool
	offsetAttribute   attribute.OffsetAttr
	payloadAttribute  attribute.PayloadAttr
	termFreqAtt       attribute.TermFreqAttr
	termBytePool      *bytesref.BlockPool
	hasPayloads       bool // if enabled, and we actually saw any for this field
}

func NewTermVectorsConsumerPerField(invertState *FieldInvertState,
	termsHash *TermVectorsConsumer, fieldInfo *document.FieldInfo) (*TermVectorsConsumerPerField, error) {

	indexOptions := fieldInfo.GetIndexOptions()
	termBytePool := termsHash.GetTermBytePool()

	perfield := &TermVectorsConsumerPerField{
		termsWriter:  termsHash,
		fieldState:   invertState,
		fieldInfo:    fieldInfo,
		termBytePool: termsHash.termBytePool,
	}

	perfield.baseTermsHashPerField = newBaseTermsHashPerField(2,
		termsHash.GetIntPool(), termsHash.GetBytePool(), termsHash.GetTermBytePool(),
		nil, fieldInfo.Name(), indexOptions, perfield)

	byteStarts := NewPostingsBytesStartArray(perfield)
	bytesHash, err := bytesref.NewBytesHash(termBytePool, bytesref.WithCapacity(HASH_INIT_SIZE), bytesref.WithStartArray(byteStarts))
	if err != nil {
		return nil, err
	}
	perfield.bytesHash = bytesHash
	return perfield, nil
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

func (t *TermVectorsConsumerPerField) Start(field document.IndexableField, first bool) bool {
	t.baseTermsHashPerField.Start(field, first)
	t.termFreqAtt = t.fieldState.termFreqAttribute

	if first {

		if t.getNumTerms() != 0 {
			// Only necessary if previous doc hit a
			// non-aborting exception while writing vectors in
			// this field:
			t.Reset()
		}

		t.hasPayloads = false

		t.doVectors = field.FieldType().StoreTermVectors()

		if t.doVectors {

			t.termsWriter.hasVectors = true

			t.doVectorPositions = field.FieldType().StoreTermVectorPositions()

			// Somewhat confusingly, unlike postings, you are
			// allowed to index TV offsets without TV positions:
			t.doVectorOffsets = field.FieldType().StoreTermVectorOffsets()

			if t.doVectorPositions {
				t.doVectorPayloads = field.FieldType().StoreTermVectorPayloads()
			} else {
				t.doVectorPayloads = false
				if field.FieldType().StoreTermVectorPayloads() {
					// TODO: move this check somewhere else, and impl the other missing ones
					panic(fmt.Sprintf("cannot index term vector payloads without term vector positions (field='%s')", field.Name()))
					//throw new IllegalArgumentException("cannot index term vector payloads without term vector positions (field=\"" + field.name() + "\")");
				}
			}

		} else {
			if field.FieldType().StoreTermVectorOffsets() {
				panic(fmt.Sprintf("cannot index term vector offsets when term vectors are not indexed (field='%s')", field.Name()))
				//throw new IllegalArgumentException("cannot index term vector offsets when term vectors are not indexed (field=\"" + field.name() + "\")");
			}
			if field.FieldType().StoreTermVectorPositions() {
				panic(fmt.Sprintf("cannot index term vector positions when term vectors are not indexed (field='%s')", field.Name()))
				//throw new IllegalArgumentException("cannot index term vector positions when term vectors are not indexed (field=\"" + field.name() + "\")");
			}
			if field.FieldType().StoreTermVectorPayloads() {
				panic(fmt.Sprintf("cannot index term vector payloads when term vectors are not indexed (field='%s')", field.Name()))
				//throw new IllegalArgumentException("cannot index term vector payloads when term vectors are not indexed (field=\"" + field.name() + "\")");
			}
		}
	} else {
		if t.doVectors != field.FieldType().StoreTermVectors() {
			panic(fmt.Sprintf("all instances of a given field name must have the same term vectors settings (storeTermVectors changed for field='%s'", field.Name()))
			//throw new IllegalArgumentException("all instances of a given field name must have the same term vectors settings (storeTermVectors changed for field=\"" + field.name() + "\")");
		}
		if t.doVectorPositions != field.FieldType().StoreTermVectorPositions() {
			panic(fmt.Sprintf("all instances of a given field name must have the same term vectors settings (storeTermVectorPositions changed for field='%s'", field.Name()))
			//throw new IllegalArgumentException("all instances of a given field name must have the same term vectors settings (storeTermVectorPositions changed for field=\"" + field.name() + "\")");
		}
		if t.doVectorOffsets != field.FieldType().StoreTermVectorOffsets() {
			panic(fmt.Sprintf("all instances of a given field name must have the same term vectors settings (storeTermVectorOffsets changed for field='%s'", field.Name()))
			//throw new IllegalArgumentException("all instances of a given field name must have the same term vectors settings (storeTermVectorOffsets changed for field=\"" + field.name() + "\")");
		}
		if t.doVectorPayloads != field.FieldType().StoreTermVectorPayloads() {
			panic(fmt.Sprintf("all instances of a given field name must have the same term vectors settings (storeTermVectorPayloads changed for field='%s'", field.Name()))
			//throw new IllegalArgumentException("all instances of a given field name must have the same term vectors settings (storeTermVectorPayloads changed for field=\"" + field.name() + "\")");
		}
	}

	if t.doVectors {
		if t.doVectorOffsets {
			t.offsetAttribute = t.fieldState.offsetAttribute
			//assert offsetAttribute != null;
		}

		if t.doVectorPayloads {
			// Can be null:
			t.payloadAttribute = t.fieldState.payloadAttribute
		} else {
			t.payloadAttribute = nil
		}
	}
	return t.doVectors
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

func (t *TermVectorsConsumerPerField) Reset() error {
	t.bytesHash.Clear(false)
	t.sortedTermIDs = nil
	if t.nextPerField != nil {
		return t.nextPerField.Reset()
	}
	return nil
}

func SortTermVectorsConsumerPerField(fields []*TermVectorsConsumerPerField) {
	sort.Sort(TermVectorsConsumerPerFields(fields))
}

var _ sort.Interface = TermVectorsConsumerPerFields{}

type TermVectorsConsumerPerFields []*TermVectorsConsumerPerField

func (p TermVectorsConsumerPerFields) Len() int {
	return len(p)
}

func (p TermVectorsConsumerPerFields) Less(i, j int) bool {
	return strings.Compare(p[i].fieldName, p[j].fieldName) < 0
}

func (p TermVectorsConsumerPerFields) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
