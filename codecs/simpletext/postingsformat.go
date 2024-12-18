package simpletext

import (
	"context"

	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
)

const (
	POSTINGS_EXTENSION = "pst"
)

var _ index.PostingsFormat = &PostingsFormat{}

// PostingsFormat For debugging, curiosity, transparency only!! Do not use this codec in production.
// This codec stores all postings data in a single human-readable text file (_N.pst). You can view this in any text editor, and even edit it to alter your index.
// lucene.experimental
type PostingsFormat struct {
	name string
}

func NewPostingsFormat() *PostingsFormat {
	return &PostingsFormat{name: "SimpleText"}
}

func (s *PostingsFormat) GetName() string {
	return s.name
}

func (s *PostingsFormat) FieldsConsumer(ctx context.Context, state *index.SegmentWriteState) (index.FieldsConsumer, error) {
	return NewFieldsWriter(ctx, state)
}

func (s *PostingsFormat) FieldsProducer(ctx context.Context, state *index.SegmentReadState) (index.FieldsProducer, error) {
	return NewSimpleTextFieldsReader(state)
}

func getPostingsFileName(segment, segmentSuffix string) string {
	return store.SegmentFileName(segment, segmentSuffix, POSTINGS_EXTENSION)
}
