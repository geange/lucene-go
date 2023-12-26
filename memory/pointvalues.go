package memory

import (
	"context"
	"github.com/geange/lucene-go/core/types"
)

var _ types.PointValues = &memPointValues{}

type memPointValues struct {
	info *info
}

func newMemoryIndexPointValues(info *info) *memPointValues {
	return &memPointValues{info: info}
}

func (m *memPointValues) Intersect(ctx context.Context, visitor types.IntersectVisitor) error {
	values := m.info.pointValues
	visitor.Grow(m.info.pointValuesCount)
	for i := 0; i < m.info.pointValuesCount; i++ {
		err := visitor.VisitLeaf(nil, 0, values[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *memPointValues) EstimatePointCount(ctx context.Context, visitor types.IntersectVisitor) (int, error) {
	return 1, nil
}

func (m *memPointValues) EstimateDocCount(visitor types.IntersectVisitor) (int, error) {
	return types.EstimateDocCount(m, visitor)
}

func (m *memPointValues) GetMinPackedValue() ([]byte, error) {
	return m.info.minPackedValue, nil
}

func (m *memPointValues) GetMaxPackedValue() ([]byte, error) {
	return m.info.maxPackedValue, nil
}

func (m *memPointValues) GetNumDimensions() (int, error) {
	return m.info.fieldInfo.GetPointIndexDimensionCount(), nil
}

func (m *memPointValues) GetNumIndexDimensions() (int, error) {
	return m.info.fieldInfo.GetPointDimensionCount(), nil
}

func (m *memPointValues) GetBytesPerDimension() (int, error) {
	return m.info.fieldInfo.GetPointNumBytes(), nil
}

func (m *memPointValues) Size() int {
	return m.info.pointValuesCount
}

func (m *memPointValues) GetDocCount() int {
	return 1
}
