package simpletext

import (
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
)

var _ index.TermVectorsFormat = &TermVectorsFormat{}

type TermVectorsFormat struct {
}

func NewTermVectorsFormat() *TermVectorsFormat {
	return &TermVectorsFormat{}
}

func (s *TermVectorsFormat) VectorsReader(dir store.Directory, segmentInfo *index.SegmentInfo,
	fieldInfos *index.FieldInfos, context *store.IOContext) (index.TermVectorsReader, error) {
	return NewTermVectorsReader(dir, segmentInfo, context)
}

func (s *TermVectorsFormat) VectorsWriter(dir store.Directory,
	segmentInfo *index.SegmentInfo, context *store.IOContext) (index.TermVectorsWriter, error) {

	return NewTermVectorsWriter(dir, segmentInfo.Name(), context)
}
