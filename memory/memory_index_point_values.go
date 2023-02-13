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

func (m *MemoryIndexPointValues) Intersect(visitor *index.IntersectVisitor) error {
	values := m.info.pointValues
	visitor.Grow(m.info.pointValuesCount)
	for i := 0; i < m.info.pointValuesCount; i++ {
		err := visitor.VisitLeaf(0, values[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *MemoryIndexPointValues) EstimatePointCount(visitor *index.IntersectVisitor) int64 {
	return 1
}

func (m *MemoryIndexPointValues) GetMinPackedValue() ([]byte, error) {
	return m.info.minPackedValue, nil
}

func (m *MemoryIndexPointValues) GetMaxPackedValue() ([]byte, error) {
	return m.info.maxPackedValue, nil
}

func (m *MemoryIndexPointValues) GetNumDimensions() (int, error) {
	return m.info.fieldInfo.GetPointIndexDimensionCount(), nil
}

func (m *MemoryIndexPointValues) GetNumIndexDimensions() (int, error) {
	return m.info.fieldInfo.GetPointDimensionCount(), nil
}

func (m *MemoryIndexPointValues) GetBytesPerDimension() (int, error) {
	return m.info.fieldInfo.GetPointNumBytes(), nil
}

func (m *MemoryIndexPointValues) Size() int64 {
	return int64(m.info.pointValuesCount)
}

func (m *MemoryIndexPointValues) GetDocCount() int {
	return 1
}
