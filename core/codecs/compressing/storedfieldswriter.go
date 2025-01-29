package compressing

import (
	"context"

	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
)

var _ index.StoredFieldsWriter = &StoredFieldsWriter{}

type StoredFieldsWriter struct {
}

func (s *StoredFieldsWriter) Close() error {
	//TODO implement me
	panic("implement me")
}

func (s *StoredFieldsWriter) StartDocument(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (s *StoredFieldsWriter) FinishDocument(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (s *StoredFieldsWriter) WriteField(ctx context.Context, fieldInfo *document.FieldInfo, field document.IndexableField) error {
	//TODO implement me
	panic("implement me")
}

func (s *StoredFieldsWriter) Finish(ctx context.Context, fieldInfos index.FieldInfos, numDocs int) error {
	//TODO implement me
	panic("implement me")
}
