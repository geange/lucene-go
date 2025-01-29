package compressing

import (
	"context"

	"github.com/geange/lucene-go/core/interface/index"
)

var _ index.TermVectorsReader = &TermVectorsReader{}

type TermVectorsReader struct {
}

func (t *TermVectorsReader) Close() error {
	//TODO implement me
	panic("implement me")
}

func (t *TermVectorsReader) Get(ctx context.Context, doc int) (index.Fields, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TermVectorsReader) CheckIntegrity() error {
	//TODO implement me
	panic("implement me")
}

func (t *TermVectorsReader) Clone(ctx context.Context) index.TermVectorsReader {
	//TODO implement me
	panic("implement me")
}

func (t *TermVectorsReader) GetMergeInstance() index.TermVectorsReader {
	//TODO implement me
	panic("implement me")
}
