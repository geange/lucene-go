package index

import (
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util"
)

var _ TermsHash = &TermVectorsConsumer{}

type TermVectorsConsumer struct {
	*TermsHashDefault

	directory       store.Directory
	info            *SegmentInfo
	codec           Codec
	writer          TermVectorsWriter
	hasVectors      bool
	numVectorFields int
	lastDocID       int
	perFields       []*TermVectorsConsumerPerField
}

func NewTermVectorsConsumer(intBlockAllocator util.IntsAllocator,
	byteBlockAllocator util.BytesAllocator, directory store.Directory,
	info *SegmentInfo, codec Codec) *TermVectorsConsumer {

	termsHashDefault := NewTermsHashDefault(intBlockAllocator, byteBlockAllocator, nil)
	return &TermVectorsConsumer{
		TermsHashDefault: termsHashDefault,
		directory:        directory,
		info:             info,
		codec:            codec,
	}
}

func (t *TermVectorsConsumer) SetTermBytePool(termBytePool *util.ByteBlockPool) {
	t.termBytePool = termBytePool
}

func (t *TermVectorsConsumer) Flush(fieldsToFlush map[string]TermsHashPerField,
	state *SegmentWriteState, sortMap *DocMap, norms NormsProducer) error {

	if t.writer != nil {
		numDocs, err := state.SegmentInfo.MaxDoc()
		if err != nil {
			return err
		}
		if err := t.fill(numDocs); err != nil {
			return err
		}
		if err := t.writer.Finish(state.FieldInfos, numDocs); err != nil {
			return err
		}
		return t.writer.Close()
	}
	return nil
}

// Fills in no-term-vectors for all docs we haven't seen since the last doc that had term vectors.
func (t *TermVectorsConsumer) fill(docID int) error {
	for t.lastDocID < docID {
		if err := t.writer.StartDocument(0); err != nil {
			return err
		}
		if err := t.writer.FinishDocument(); err != nil {
			return err
		}
		t.lastDocID++
	}
	return nil
}

func (t *TermVectorsConsumer) initTermVectorsWriter() error {
	if t.writer == nil {
		writer, err := t.codec.TermVectorsFormat().VectorsWriter(t.directory, t.info, nil)
		if err != nil {
			return err
		}
		t.writer = writer
		t.lastDocID = 0
	}
	return nil
}

func (t *TermVectorsConsumer) AddField(invertState *FieldInvertState,
	fieldInfo *types.FieldInfo) (TermsHashPerField, error) {

	return NewTermVectorsConsumerPerField(invertState, t, fieldInfo), nil
}

func (t *TermVectorsConsumer) addFieldToFlush(fieldToFlush *TermVectorsConsumerPerField) error {
	t.perFields = append(t.perFields, fieldToFlush)
	return nil
}

func (t *TermVectorsConsumer) FinishDocument(docID int) error {
	if !t.hasVectors {
		return nil
	}

	// Fields in term vectors are UTF16 sorted:
	SortTermVectorsConsumerPerField(t.perFields)

	if err := t.initTermVectorsWriter(); err != nil {
		return err
	}

	if err := t.fill(docID); err != nil {
		return err
	}

	t.writer.StartDocument(t.numVectorFields)
	for i := 0; i < t.numVectorFields; i++ {
		t.perFields[i].FinishDocument()
	}
	t.writer.FinishDocument()

	t.lastDocID++
	t.Reset()
	t.resetFields()
	return nil
}

func (t *TermVectorsConsumer) resetFields() {
	t.perFields = t.perFields[:0]
	t.numVectorFields = 0
}
