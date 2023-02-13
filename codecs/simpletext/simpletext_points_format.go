package simpletext

import "github.com/geange/lucene-go/core/index"

var _ index.PointsFormat = &SimpleTextPointsFormat{}

// SimpleTextPointsFormat For debugging, curiosity, transparency only!! Do not use this codec in production.
// This codec stores all dimensional data in a single human-readable text file (_N.dim).
// You can view this in any text editor, and even edit it to alter your index.
// lucene.experimental
type SimpleTextPointsFormat struct {
}

func NewSimpleTextPointsFormat() *SimpleTextPointsFormat {
	return &SimpleTextPointsFormat{}
}

func (s *SimpleTextPointsFormat) FieldsWriter(state *index.SegmentWriteState) (index.PointsWriter, error) {
	return NewSimpleTextPointsWriter(state)
}

func (s *SimpleTextPointsFormat) FieldsReader(state *index.SegmentReadState) (index.PointsReader, error) {
	return NewSimpleTextPointsReader(state)
}

const (
	// POINT_EXTENSION Extension of points data file
	POINT_EXTENSION = "dim"

	// POINT_INDEX_EXTENSION Extension of points index file
	POINT_INDEX_EXTENSION = "dii"
)
