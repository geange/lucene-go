package lucene84

import (
	"context"

	"github.com/geange/lucene-go/core/codecs/blocktree"
	"github.com/geange/lucene-go/core/interface/index"
)

const (
	DOC_EXTENSION   = "doc"
	POS_EXTENSION   = "pos"
	PAY_EXTENSION   = "pay"
	BLOCK_SIZE      = 128
	MAX_SKIP_LEVELS = 10

	TERMS_CODEC = "Lucene84PostingsWriterTerms"
	DOC_CODEC   = "Lucene84PostingsWriterDoc"
	POS_CODEC   = "Lucene84PostingsWriterPos"
	PAY_CODEC   = "Lucene84PostingsWriterPay"

	VERSION_START                     = 0
	VERSION_COMPRESSED_TERMS_DICT_IDS = 1
	VERSION_CURRENT                   = VERSION_COMPRESSED_TERMS_DICT_IDS
)

var _ index.PostingsFormat = &PostingsFormat{}

type PostingsFormat struct {
	name             string
	minTermBlockSize int
	maxTermBlockSize int
}

func NewPostingsFormat() *PostingsFormat {
	return &PostingsFormat{}
}

func (p *PostingsFormat) GetName() string {
	return p.name
}

func (p *PostingsFormat) FieldsConsumer(ctx context.Context, state *index.SegmentWriteState) (index.FieldsConsumer, error) {
	postingsWriter, err := NewPostingsWriter(ctx, state)
	if err != nil {
		return nil, err
	}
	return blocktree.NewTermsWriter(ctx, state, postingsWriter, p.minTermBlockSize, p.maxTermBlockSize)
}

func (p *PostingsFormat) FieldsProducer(ctx context.Context, state *index.SegmentReadState) (index.FieldsProducer, error) {
	postingsReader, err := NewPostingsReader(ctx, state)
	if err != nil {
		return nil, err
	}
	return blocktree.NewTermsReader(ctx, postingsReader, state)
}
