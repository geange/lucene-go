package index

import (
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/bytesref"
	"github.com/geange/lucene-go/core/util/ints"
)

var _ TermsHash = &TermVectorsConsumer{}

type TermVectorsConsumer struct {
	*BaseTermsHash

	directory       store.Directory
	info            *SegmentInfo
	codec           Codec
	writer          TermVectorsWriter
	hasVectors      bool
	numVectorFields int
	lastDocID       int
	perFields       []*TermVectorsConsumerPerField
}

func NewTermVectorsConsumer(intBlockAllocator ints.IntsAllocator,
	byteBlockAllocator bytesref.Allocator, directory store.Directory,
	info *SegmentInfo, codec Codec) *TermVectorsConsumer {

	termsHashDefault := NewTermsHashDefault(intBlockAllocator, byteBlockAllocator, nil)
	return &TermVectorsConsumer{
		BaseTermsHash: termsHashDefault,
		directory:     directory,
		info:          info,
		codec:         codec,
	}
}

func (t *TermVectorsConsumer) SetTermBytePool(termBytePool *bytesref.BlockPool) {
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
		if err := t.writer.Finish(nil, state.FieldInfos, numDocs); err != nil {
			return err
		}
		return t.writer.Close()
	}
	return nil
}

// Fills in no-term-vectors for all docs we haven't seen since the last doc that had term vectors.
func (t *TermVectorsConsumer) fill(docID int) error {
	for t.lastDocID < docID {
		if err := t.writer.StartDocument(nil, 0); err != nil {
			return err
		}
		if err := t.writer.FinishDocument(nil); err != nil {
			return err
		}
		t.lastDocID++
	}
	return nil
}

func (t *TermVectorsConsumer) initTermVectorsWriter() error {
	if t.writer == nil {
		writer, err := t.codec.TermVectorsFormat().VectorsWriter(nil, t.directory, t.info, nil)
		if err != nil {
			return err
		}
		t.writer = writer
		t.lastDocID = 0
	}
	return nil
}

func (t *TermVectorsConsumer) AddField(invertState *index.FieldInvertState,
	fieldInfo *document.FieldInfo) (TermsHashPerField, error) {

	return NewTermVectorsConsumerPerField(invertState, t, fieldInfo)
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

	t.writer.StartDocument(nil, t.numVectorFields)
	for i := 0; i < t.numVectorFields; i++ {
		t.perFields[i].FinishDocument()
	}
	t.writer.FinishDocument(nil)

	t.lastDocID++
	t.Reset()
	t.resetFields()
	return nil
}

func (t *TermVectorsConsumer) resetFields() {
	t.perFields = t.perFields[:0]
	t.numVectorFields = 0
}
