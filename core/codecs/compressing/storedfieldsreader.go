package compressing

import (
	"context"

	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
)

var _ index.StoredFieldsReader = &StoredFieldsReader{}

type StoredFieldsReader struct {
}

func (s *StoredFieldsReader) Close() error {
	//TODO implement me
	panic("implement me")
}

func (s *StoredFieldsReader) VisitDocument(ctx context.Context, docID int, visitor document.StoredFieldVisitor) error {
	//TODO implement me
	panic("implement me")
}

func (s *StoredFieldsReader) Clone(ctx context.Context) index.StoredFieldsReader {
	//TODO implement me
	panic("implement me")
}

func (s *StoredFieldsReader) CheckIntegrity() error {
	//TODO implement me
	panic("implement me")
}

func (s *StoredFieldsReader) GetMergeInstance() index.StoredFieldsReader {
	//TODO implement me
	panic("implement me")
}
