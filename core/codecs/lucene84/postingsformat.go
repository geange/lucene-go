package lucene84

import (
	"context"

	"github.com/geange/lucene-go/core/interface/index"
)

var _ index.PostingsFormat = &PostingsFormat{}

type PostingsFormat struct {
}

func NewPostingsFormat() *PostingsFormat {
	return &PostingsFormat{}
}

func (p *PostingsFormat) GetName() string {
	//TODO implement me
	panic("implement me")
}

func (p *PostingsFormat) FieldsConsumer(ctx context.Context, state *index.SegmentWriteState) (index.FieldsConsumer, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PostingsFormat) FieldsProducer(ctx context.Context, state *index.SegmentReadState) (index.FieldsProducer, error) {
	//TODO implement me
	panic("implement me")
}
