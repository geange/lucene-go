package simpletext

import (
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
)

var _ index.PointValues = &TextBKDReader{}

type TextBKDReader struct {
	splitPackedValues      []byte
	leafBlockFPs           []int64
	leafNodeOffset         int
	numDims                int
	numIndexDims           int
	bytesPerDim            int
	bytesPerIndexEntry     int
	in                     store.IndexInput
	maxPointsInLeafNode    int
	minPackedValue         []byte
	maxPackedValue         []byte
	pointCount             int64
	docCount               int
	version                int
	packedBytesLength      int
	packedIndexBytesLength int
}

func (s *TextBKDReader) Intersect(visitor index.IntersectVisitor) error {
	//TODO implement me
	panic("implement me")
}

func (s *TextBKDReader) EstimatePointCount(visitor index.IntersectVisitor) int64 {
	//TODO implement me
	panic("implement me")
}

func (s *TextBKDReader) GetMinPackedValue() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (s *TextBKDReader) GetMaxPackedValue() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (s *TextBKDReader) GetNumDimensions() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *TextBKDReader) GetNumIndexDimensions() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *TextBKDReader) GetBytesPerDimension() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *TextBKDReader) Size() int64 {
	//TODO implement me
	panic("implement me")
}

func (s *TextBKDReader) GetDocCount() int {
	//TODO implement me
	panic("implement me")
}
