package simpletext

import (
	"github.com/geange/lucene-go/core/index"
)

var _ index.FieldsProducer = &FieldsReader{}

type FieldsReader struct {
}

func (s *FieldsReader) Close() error {
	//TODO implement me
	panic("implement me")
}

func (s *FieldsReader) CheckIntegrity() error {
	//TODO implement me
	panic("implement me")
}

func (s *FieldsReader) GetMergeInstance() index.FieldsProducer {
	//TODO implement me
	panic("implement me")
}
