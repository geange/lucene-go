package index

import (
	"bytes"
	"context"
	"fmt"
	"github.com/geange/lucene-go/core/interface/index"

	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
)

// PointValuesWriter Buffers up pending byte[][] item(s) per doc, then flushes when segment flushes.
type PointValuesWriter struct {
	fieldInfo         *document.FieldInfo
	bytes             *PagedBytes
	bytesOut          store.DataOutput
	docIDs            []int
	numPoints         int
	numDocs           int
	lastDocID         int
	packedBytesLength int
}

func NewPointValuesWriter(fieldInfo *document.FieldInfo) *PointValuesWriter {
	pages := NewPagedBytes(16)
	return &PointValuesWriter{
		fieldInfo:         fieldInfo,
		bytes:             pages,
		bytesOut:          pages.GetDataOutput(),
		docIDs:            make([]int, 0, 16),
		packedBytesLength: fieldInfo.GetPointDimensionCount() * fieldInfo.GetPointNumBytes(),
	}
}

// AddPackedValue
// TODO: if exactly the same item is added to exactly the same doc, should we dedup?
func (p *PointValuesWriter) AddPackedValue(docID int, value []byte) error {
	if len(value) != p.packedBytesLength {
		return fmt.Errorf("this field's item has length=%d", len(value))
	}

	if _, err := p.bytesOut.Write(value); err != nil {
		return err
	}
	p.docIDs = append(p.docIDs, docID)
	if docID != p.lastDocID {
		p.numDocs++
		p.lastDocID = docID
	}
	p.numPoints++
	return nil
}

func (p *PointValuesWriter) Flush(ctx context.Context, state *index.SegmentWriteState, docMap index.DocMap, writer index.PointsWriter) error {
	bytesReader, err := p.bytes.Freeze(false)
	if err != nil {
		return err
	}

	points := &innerMutablePointValues{
		bytesReader: bytesReader,
		pw:          p,
		numDocs:     p.numDocs,
		numPoints:   p.numPoints,
		ords:        make([]int, p.numPoints),
		temp:        nil,
	}

	for i := 0; i < p.numPoints; i++ {
		points.ords[i] = i
	}
	reader := &innerReader{values: points}
	return writer.WriteField(ctx, p.fieldInfo, reader)
}

var _ index.PointsReader = &innerReader{}

type innerReader struct {
	values types.PointValues
}

func (i *innerReader) Close() error {
	return nil
}

func (i *innerReader) CheckIntegrity() error {
	//TODO implement me
	panic("implement me")
}

func (i *innerReader) GetValues(ctx context.Context, field string) (types.PointValues, error) {
	return i.values, nil
}

func (i *innerReader) GetMergeInstance() index.PointsReader {
	return i
}

var _ types.MutablePointValues = &innerMutablePointValues{}

type innerMutablePointValues struct {
	bytesReader *PagedBytesReader
	pw          *PointValuesWriter
	numDocs     int
	numPoints   int
	ords        []int
	temp        []int
}

func (r *innerMutablePointValues) Intersect(ctx context.Context, visitor types.IntersectVisitor) error {
	scratch := new(bytes.Buffer)
	packedValue := make([]byte, r.pw.packedBytesLength)

	for i := 0; i < r.numPoints; i++ {
		r.GetValue(i, scratch)
		copy(packedValue, scratch.Bytes())
		if err := visitor.VisitLeaf(ctx, r.GetDocID(i), packedValue); err != nil {
			return err
		}
	}
	return nil
}

func (r *innerMutablePointValues) EstimatePointCount(ctx context.Context, visitor types.IntersectVisitor) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (r *innerMutablePointValues) EstimateDocCount(visitor types.IntersectVisitor) (int, error) {
	return types.EstimateDocCount(r, visitor)
}

func (r *innerMutablePointValues) GetMinPackedValue() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (r *innerMutablePointValues) GetMaxPackedValue() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (r *innerMutablePointValues) GetNumDimensions() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (r *innerMutablePointValues) GetNumIndexDimensions() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (r *innerMutablePointValues) GetBytesPerDimension() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (r *innerMutablePointValues) Size() int {
	return r.numPoints
}

func (r *innerMutablePointValues) GetDocCount() int {
	return r.numDocs
}

func (r *innerMutablePointValues) GetValue(i int, packedValue *bytes.Buffer) {
	offset := r.pw.packedBytesLength * r.ords[i]
	r.bytesReader.FillSlice(packedValue, offset, r.pw.packedBytesLength)
}

func (r *innerMutablePointValues) GetByteAt(i, k int) byte {
	offset := r.pw.packedBytesLength*r.ords[i] + k
	return r.bytesReader.GetByte(int64(offset))
}

func (r *innerMutablePointValues) GetDocID(i int) int {
	return r.pw.docIDs[r.ords[i]]
}

func (r *innerMutablePointValues) Swap(i, j int) {
	r.ords[i], r.ords[j] = r.ords[j], r.ords[i]
}

func (r *innerMutablePointValues) Save(i, j int) {
	if len(r.temp) == 0 {
		r.temp = make([]int, len(r.ords))
	}
	r.temp[j] = r.ords[i]
}

func (r *innerMutablePointValues) Restore(i, j int) {
	if len(r.temp) != 0 {
		copy(r.ords[i:], r.temp[i:j-i])
	}
}
