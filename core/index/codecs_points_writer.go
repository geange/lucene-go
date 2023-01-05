package index

import (
	"github.com/geange/lucene-go/core/types"
	"io"
)

// PointsWriter Abstract API to write points
// lucene.experimental
type PointsWriter interface {
	io.Closer

	// WriteField Write all values contained in the provided reader
	WriteField(fieldInfo *types.FieldInfo, values PointsReader) error

	// Finish Called once at the end before close
	Finish() error
}

type PointsWriterDefault struct {
	WriteField func(fieldInfo *types.FieldInfo, values PointsReader) error
	Finish     func() error
}

// MergeOneField Default naive merge implementation for one field: it just re-indexes all
// the values from the incoming segment. The default codec overrides this for 1D fields and
// uses a faster but more complex implementation.
func (p *PointsWriterDefault) MergeOneField(mergeState *MergeState, fieldInfo *types.FieldInfo) error {
	maxPointCount := int64(0)
	docCount := 0

	for i, pointsReader := range mergeState.PointsReaders {
		if pointsReader != nil {
			readerFieldInfo := mergeState.FieldInfos[i].FieldInfo(fieldInfo.Name)
			if readerFieldInfo != nil && readerFieldInfo.GetPointIndexDimensionCount() > 0 {
				values, err := pointsReader.GetValues(fieldInfo.Name)
				if err != nil {
					return err
				}
				if values != nil {
					maxPointCount += values.Size()
					docCount += values.GetDocCount()
				}
			}
		}
	}

	finalMaxPointCount := maxPointCount
	finalDocCount := docCount

	return p.WriteField(fieldInfo, &innerPointsReader{
		size:     finalMaxPointCount,
		docCount: finalDocCount,
	})
}

var _ PointsReader = &innerPointsReader{}

type innerPointsReader struct {
	size     int64
	docCount int
}

func (i innerPointsReader) Close() error {
	return nil
}

func (i innerPointsReader) CheckIntegrity() error {
	//TODO implement me
	panic("implement me")
}

func (i innerPointsReader) GetValues(field string) (PointValues, error) {
	return &innerPointValues{
		size:     i.size,
		docCount: i.docCount,
	}, nil
}

var _ PointValues = &innerPointValues{}

type innerPointValues struct {
	size     int64
	docCount int
}

func (i *innerPointValues) Intersect(visitor IntersectVisitor) error {
	//TODO implement me
	panic("implement me")
}

func (i *innerPointValues) EstimatePointCount(visitor IntersectVisitor) int64 {
	//TODO implement me
	panic("implement me")
}

func (i *innerPointValues) GetMinPackedValue() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (i *innerPointValues) GetMaxPackedValue() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (i *innerPointValues) GetNumDimensions() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (i *innerPointValues) GetNumIndexDimensions() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (i *innerPointValues) GetBytesPerDimension() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (i *innerPointValues) Size() int64 {
	return i.size
}

func (i *innerPointValues) GetDocCount() int {
	return i.docCount
}

// Merge Default merge implementation to merge incoming points readers by visiting all their points and adding to this writer
func (p *PointsWriterDefault) Merge(mergeState *MergeState) error {
	// check each incoming reader
	for _, reader := range mergeState.PointsReaders {
		if reader == nil {
			continue
		}
		if err := reader.CheckIntegrity(); err != nil {
			return err
		}
	}
	// merge field at a time
	for _, fieldInfo := range mergeState.MergeFieldInfos.List() {
		if fieldInfo.GetPointDimensionCount() == 0 {
			continue
		}
		if err := p.MergeOneField(mergeState, fieldInfo); err != nil {
			return err
		}
	}
	return p.Finish()
}
