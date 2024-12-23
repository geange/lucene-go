package index

import (
	"context"

	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/types"
)

type BasePointsWriter struct {
	WriteField func(ctx context.Context, fieldInfo *document.FieldInfo, values index.PointsReader) error
	Finish     func() error
}

// MergeOneField
// Default naive merge implementation for one field: it just re-indexes all
// the values from the incoming segment. The default codec overrides this for 1D fields and
// uses a faster but more complex implementation.
func (p *BasePointsWriter) MergeOneField(ctx context.Context, mergeState *MergeState, fieldInfo *document.FieldInfo) error {
	maxPointCount := 0
	docCount := 0

	for i, pointsReader := range mergeState.PointsReaders {
		if pointsReader != nil {
			readerFieldInfo := mergeState.FieldInfos[i].FieldInfo(fieldInfo.Name())
			if readerFieldInfo != nil && readerFieldInfo.GetPointIndexDimensionCount() > 0 {
				values, err := pointsReader.GetValues(ctx, fieldInfo.Name())
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

	return p.WriteField(ctx, fieldInfo, newInnerPointsReader(finalMaxPointCount, finalDocCount))
}

var _ index.PointsReader = &innerPointsReader{}

type innerPointsReader struct {
	size     int
	docCount int
}

func newInnerPointsReader(size int, docCount int) *innerPointsReader {
	return &innerPointsReader{size: size, docCount: docCount}
}

func (i *innerPointsReader) Close() error {
	return nil
}

func (i *innerPointsReader) CheckIntegrity() error {
	//TODO implement me
	panic("implement me")
}

func (i *innerPointsReader) GetValues(ctx context.Context, field string) (types.PointValues, error) {
	return &innerPointValues{
		size:     i.size,
		docCount: i.docCount,
	}, nil
}

func (i *innerPointsReader) GetMergeInstance() index.PointsReader {
	return i
}

var _ types.PointValues = &innerPointValues{}

type innerPointValues struct {
	size     int
	docCount int
}

func (i *innerPointValues) Intersect(ctx context.Context, visitor types.IntersectVisitor) error {
	//TODO implement me
	panic("implement me")
}

func (i *innerPointValues) EstimatePointCount(ctx context.Context, visitor types.IntersectVisitor) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (i *innerPointValues) EstimateDocCount(ctx context.Context, visitor types.IntersectVisitor) (int, error) {
	return types.EstimateDocCount(ctx, i, visitor)
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

func (i *innerPointValues) Size() int {
	return i.size
}

func (i *innerPointValues) GetDocCount() int {
	return i.docCount
}

// Merge Default merge implementation to merge incoming points readers by visiting all their points and adding to this writer
func (p *BasePointsWriter) Merge(mergeState *MergeState) error {
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
		if err := p.MergeOneField(nil, mergeState, fieldInfo); err != nil {
			return err
		}
	}
	return p.Finish()
}
