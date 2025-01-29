package compressing

import (
	"context"

	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
)

var _ index.TermVectorsWriter = &TermVectorsWriter{}

type TermVectorsWriter struct {
}

func (t *TermVectorsWriter) Close() error {
	//TODO implement me
	panic("implement me")
}

func (t *TermVectorsWriter) StartDocument(ctx context.Context, numVectorFields int) error {
	//TODO implement me
	panic("implement me")
}

func (t *TermVectorsWriter) FinishDocument(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (t *TermVectorsWriter) StartField(ctx context.Context, fieldInfo *document.FieldInfo, numTerms int, positions, offsets, payloads bool) error {
	//TODO implement me
	panic("implement me")
}

func (t *TermVectorsWriter) FinishField(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (t *TermVectorsWriter) StartTerm(ctx context.Context, term []byte, freq int) error {
	//TODO implement me
	panic("implement me")
}

func (t *TermVectorsWriter) FinishTerm(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (t *TermVectorsWriter) AddPosition(ctx context.Context, position, startOffset, endOffset int, payload []byte) error {
	//TODO implement me
	panic("implement me")
}

func (t *TermVectorsWriter) Finish(ctx context.Context, fieldInfos index.FieldInfos, numDocs int) error {
	//TODO implement me
	panic("implement me")
}
