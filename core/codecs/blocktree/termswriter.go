package blocktree

import (
	"context"

	"github.com/geange/lucene-go/core/codecs/types"
	"github.com/geange/lucene-go/core/interface/index"
	coreIndex "github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
)

var _ index.FieldsConsumer = &TermsWriter{}

type TermsWriter struct {
	metaOut         store.IndexOutput
	termsOut        store.IndexOutput
	indexOut        store.IndexOutput
	maxDoc          int
	minItemsInBlock int
	maxItemsInBlock int
	postingsWriter  types.PostingsWriter
	fieldInfos      coreIndex.FieldInfos
	fields          []*store.BufferDataOutput
}

func NewTermsWriter(ctx context.Context, state *index.SegmentWriteState, postingsWriter types.PostingsWriter,
	minItemsInBlock, maxItemsInBlock int) (*TermsWriter, error) {

	panic("")
}

func (t *TermsWriter) Close() error {
	//TODO implement me
	panic("implement me")
}

func (t *TermsWriter) Write(ctx context.Context, fields coreIndex.Fields, norms coreIndex.NormsProducer) error {
	//TODO implement me
	panic("implement me")
}
