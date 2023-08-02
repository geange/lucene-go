package simpletext

import (
	"context"

	"github.com/geange/lucene-go/core/interface/index"
)

var _ index.PointsFormat = &PointsFormat{}

// PointsFormat For debugging, curiosity, transparency only!! Do not use this codec in production.
// This codec stores all dimensional data in a single human-readable text file (_N.dim).
// You can view this in any text editor, and even edit it to alter your index.
// lucene.experimental
type PointsFormat struct {
}

func NewPointsFormat() *PointsFormat {
	return &PointsFormat{}
}

func (s *PointsFormat) FieldsWriter(ctx context.Context, state *index.SegmentWriteState) (index.PointsWriter, error) {
	return NewSimpleTextPointsWriter(ctx, state)
}

func (s *PointsFormat) FieldsReader(ctx context.Context, state *index.SegmentReadState) (index.PointsReader, error) {
	return NewPointsReader(ctx, state)
}

const (
	// POINT_EXTENSION Extension of points data file
	POINT_EXTENSION = "dim"

	// POINT_INDEX_EXTENSION Extension of points index file
	POINT_INDEX_EXTENSION = "dii"
)
