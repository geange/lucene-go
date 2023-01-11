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
}

func (s *SimpleTextPostingsFormat) GetName() string {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextPostingsFormat) FieldsConsumer(state *index.SegmentWriteState) (index.FieldsConsumer, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextPostingsFormat) FieldsProducer(state *index.SegmentReadState) (index.FieldsProducer, error) {
	//TODO implement me
	panic("implement me")
}

func getPostingsFileName(segment, segmentSuffix string) string {
	return store.SegmentFileName(segment, segmentSuffix, POSTINGS_EXTENSION)
}
