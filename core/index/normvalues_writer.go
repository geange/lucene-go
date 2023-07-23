package index

import (
	"errors"

	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/util/packed"
)

// NormValuesWriter Buffers up pending long per doc, then flushes when segment flushes.
type NormValuesWriter struct {
	docsWithField *DocsWithFieldSet
	pending       *packed.PackedLongValuesBuilder
	fieldInfo     *document.FieldInfo
	lastDocID     int
}

func (n *NormValuesWriter) AddValue(docID int, value int64) error {
	if n.lastDocID >= docID {
		return errors.New("docID too small")
	}
	n.pending.Add(value)
	n.lastDocID = docID
	return n.docsWithField.Add(docID)
}

func (n *NormValuesWriter) Finish(maxDoc int) {

}

func (n *NormValuesWriter) Flush(state *SegmentWriteState, sortMap *DocMap, normsConsumer NormsConsumer) error {
	//values := n.pending.Build()
	panic("")
}

func NewNormValuesWriter(fieldInfo *document.FieldInfo) *NormValuesWriter {
	return &NormValuesWriter{
		docsWithField: NewDocsWithFieldSet(),
		pending:       packed.NewPackedLongValuesBuilder(make([]uint64, 0)),
		fieldInfo:     fieldInfo,
		lastDocID:     -1,
	}
}
