package simpletext

import (
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
)

var _ index.PointValues = &SimpleTextBKDReader{}

type SimpleTextBKDReader struct {
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

func (s *SimpleTextBKDReader) Intersect(visitor index.IntersectVisitor) error {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextBKDReader) EstimatePointCount(visitor index.IntersectVisitor) int64 {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextBKDReader) GetMinPackedValue() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextBKDReader) GetMaxPackedValue() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextBKDReader) GetNumDimensions() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextBKDReader) GetNumIndexDimensions() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextBKDReader) GetBytesPerDimension() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextBKDReader) Size() int64 {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextBKDReader) GetDocCount() int {
	//TODO implement me
	panic("implement me")
}
