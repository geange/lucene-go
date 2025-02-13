package compressing

import (
	"context"

	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/packed"
)

var _ FieldsIndex = &LegacyFieldsIndexReader{}

type LegacyFieldsIndexReader struct {
	maxDoc              int
	docBases            []int
	startPointers       []int64
	avgChunkDocs        []int
	avgChunkSizes       []int
	docBasesDeltas      []packed.Reader // delta from the avg
	startPointersDeltas []packed.Reader // delta from the avg
}

// NewLegacyFieldsIndexReader
// It is the responsibility of the caller to close fieldsIndexIn after this constructor
// has been called
func NewLegacyFieldsIndexReader(ctx context.Context, fieldsIndexIn store.IndexInput, si index.SegmentInfo) (*LegacyFieldsIndexReader, error) {
	panic("")
}

func (r *LegacyFieldsIndexReader) Close() error {
	//TODO implement me
	panic("implement me")
}

func (r *LegacyFieldsIndexReader) GetStartPointer(docId int) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (r *LegacyFieldsIndexReader) CheckIntegrity() error {
	//TODO implement me
	panic("implement me")
}

func (r *LegacyFieldsIndexReader) Clone() (FieldsIndex, error) {
	//TODO implement me
	panic("implement me")
}
