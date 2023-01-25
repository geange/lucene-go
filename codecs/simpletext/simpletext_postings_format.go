package simpletext

import (
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
)

const (
	POSTINGS_EXTENSION = "pst"
)

var _ index.PostingsFormat = &SimpleTextPostingsFormat{}

// SimpleTextPostingsFormat For debugging, curiosity, transparency only!! Do not use this codec in production.
// This codec stores all postings data in a single human-readable text file (_N.pst). You can view this in any text editor, and even edit it to alter your index.
// lucene.experimental
type SimpleTextPostingsFormat struct {
	name string
}

func NewSimpleTextPostingsFormat() *SimpleTextPostingsFormat {
	return &SimpleTextPostingsFormat{name: "SimpleText"}
}

func (s *SimpleTextPostingsFormat) GetName() string {
	return s.name
}

func (s *SimpleTextPostingsFormat) FieldsConsumer(state *index.SegmentWriteState) (index.FieldsConsumer, error) {
	return NewSimpleTextFieldsWriter(state)
}

func (s *SimpleTextPostingsFormat) FieldsProducer(state *index.SegmentReadState) (index.FieldsProducer, error) {
	return NewSimpleTextFieldsReader(state)
}

func getPostingsFileName(segment, segmentSuffix string) string {
	return store.SegmentFileName(segment, segmentSuffix, POSTINGS_EXTENSION)
}
