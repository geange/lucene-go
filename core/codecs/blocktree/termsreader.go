package blocktree

import (
	"context"
	"iter"

	"github.com/geange/lucene-go/core/codecs/types"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/fst"
)

const (
	OUTPUT_FLAGS_NUM_BITS = 2
	OUTPUT_FLAGS_MASK     = 0x3
	OUTPUT_FLAG_IS_FLOOR  = 0x1
	OUTPUT_FLAG_HAS_TERMS = 0x2

	TERMS_EXTENSION             = "tim"
	TERMS_CODEC_NAME            = "BlockTreeTermsDict"
	VERSION_START               = 3
	VERSION_META_LONGS_REMOVED  = 4
	VERSION_COMPRESSED_SUFFIXES = 5
	VERSION_META_FILE           = 6
	VERSION_CURRENT             = VERSION_META_FILE
	TERMS_INDEX_EXTENSION       = "tip"
	TERMS_INDEX_CODEC_NAME      = "BlockTreeTermsIndex"
	TERMS_META_EXTENSION        = "tmd"
	TERMS_META_CODEC_NAME       = "BlockTreeTermsMeta"
)

var _ index.FieldsProducer = &TermsReader{}

type TermsReader struct {
	termsIn        store.IndexInput // Open input to the main terms dict file (_X.tib)
	indexIn        store.IndexInput // Open input to the terms index file (_X.tip)
	postingsReader types.PostingsReader
	fieldMap       map[string]*FieldReader
	fieldList      []string
	segment        string
	version        int
}

func NewTermsReader(ctx context.Context, postingsReader types.PostingsReader,
	state index.SegmentReadState) (*TermsReader, error) {
	panic("")
}

func (t *TermsReader) Close() error {
	//TODO implement me
	panic("implement me")
}

func (t *TermsReader) Iterator() iter.Seq[string] {
	//TODO implement me
	panic("implement me")
}

func (t *TermsReader) Names() []string {
	//TODO implement me
	panic("implement me")
}

func (t *TermsReader) Terms(field string) (index.Terms, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TermsReader) Size() int {
	//TODO implement me
	panic("implement me")
}

func (t *TermsReader) CheckIntegrity() error {
	//TODO implement me
	panic("implement me")
}

func (t *TermsReader) GetMergeInstance() index.FieldsProducer {
	//TODO implement me
	panic("implement me")
}

var (
	FST_OUTPUTS = fst.NewByteSequenceOutputs()
)
