package simpletext

import (
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
)

var _ index.TermVectorsFormat = &SimpleTextTermVectorsFormat{}

type SimpleTextTermVectorsFormat struct {
}

func NewSimpleTextTermVectorsFormat() *SimpleTextTermVectorsFormat {
	return &SimpleTextTermVectorsFormat{}
}

func (s *SimpleTextTermVectorsFormat) VectorsReader(dir store.Directory, segmentInfo *index.SegmentInfo,
	fieldInfos *index.FieldInfos, context *store.IOContext) (index.TermVectorsReader, error) {
	return NewSimpleTextTermVectorsReader(dir, segmentInfo, context)
}

func (s *SimpleTextTermVectorsFormat) VectorsWriter(dir store.Directory,
	segmentInfo *index.SegmentInfo, context *store.IOContext) (index.TermVectorsWriter, error) {

	return NewSimpleTextTermVectorsWriter(dir, segmentInfo.Name(), context)
}
