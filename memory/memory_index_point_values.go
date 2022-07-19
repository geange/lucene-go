package memory

import (
	"github.com/geange/lucene-go/core/index"
)

var _ index.PointValues = &MemoryIndexPointValues{}

type MemoryIndexPointValues struct {
	info *Info
}

func newMemoryIndexPointValues(info *Info) *MemoryIndexPointValues {
	return &MemoryIndexPointValues{info: info}
}

func (m *MemoryIndexPointValues) Intersect(visitor index.IntersectVisitor) error {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryIndexPointValues) EstimatePointCount(visitor index.IntersectVisitor) int64 {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryIndexPointValues) GetMinPackedValue() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryIndexPointValues) GetMaxPackedValue() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryIndexPointValues) GetNumDimensions() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryIndexPointValues) GetNumIndexDimensions() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryIndexPointValues) GetBytesPerDimension() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryIndexPointValues) Size() int64 {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryIndexPointValues) GetDocCount() int {
	//TODO implement me
	panic("implement me")
}
