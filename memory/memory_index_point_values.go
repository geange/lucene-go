package memory

import (
	"github.com/geange/lucene-go/core/index"
)

type memoryIndexPointValues struct {
	info *Info
}

func newMemoryIndexPointValues(info *Info) *memoryIndexPointValues {
	return &memoryIndexPointValues{info: info}
}

func (m *memoryIndexPointValues) Intersect(visitor index.IntersectVisitor) error {
	//TODO implement me
	panic("implement me")
}

func (m *memoryIndexPointValues) EstimatePointCount(visitor index.IntersectVisitor) int64 {
	//TODO implement me
	panic("implement me")
}

func (m *memoryIndexPointValues) GetMinPackedValue() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (m *memoryIndexPointValues) GetMaxPackedValue() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (m *memoryIndexPointValues) GetNumDimensions() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (m *memoryIndexPointValues) GetNumIndexDimensions() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (m *memoryIndexPointValues) GetBytesPerDimension() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (m *memoryIndexPointValues) Size() int64 {
	//TODO implement me
	panic("implement me")
}

func (m *memoryIndexPointValues) GetDocCount() int {
	//TODO implement me
	panic("implement me")
}
