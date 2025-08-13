package lucene80

import (
	"context"

	"github.com/geange/lucene-go/core/interface/index"
)

var _ index.DocValuesFormat = &DocValuesFormat{}

type DocValuesFormat struct {
	mode DocValuesConsumerMode
	name string
}

func NewDocValuesFormat() *DocValuesFormat {
	return NewDocValuesFormatWithMode(BEST_SPEED)
}

func NewDocValuesFormatWithMode(mode DocValuesConsumerMode) *DocValuesFormat {
	return &DocValuesFormat{
		mode: mode,
		name: "Lucene80",
	}
}

func (d *DocValuesFormat) GetName() string {
	return d.name
}

func (d *DocValuesFormat) FieldsConsumer(ctx context.Context, state *index.SegmentWriteState) (index.DocValuesConsumer, error) {
	return NewDocValuesConsumer(ctx, state, DV_DATA_CODEC, DV_DATA_EXTENSION, DV_META_CODEC, DV_META_EXTENSION, d.mode)
}

func (d *DocValuesFormat) FieldsProducer(ctx context.Context, state *index.SegmentReadState) (index.DocValuesProducer, error) {
	return NewDocValuesProducer(ctx, state, DV_DATA_CODEC, DV_DATA_EXTENSION, DV_META_CODEC, DV_META_EXTENSION)
}

const (
	DV_DATA_CODEC                       = "Lucene80DocValuesData"
	DV_DATA_EXTENSION                   = "dvd"
	DV_META_CODEC                       = "Lucene80DocValuesMetadata"
	DV_META_EXTENSION                   = "dvm"
	DV_VERSION_START                    = 0
	DV_VERSION_BIN_COMPRESSED           = 1
	DV_VERSION_CONFIGURABLE_COMPRESSION = 2
	DV_VERSION_CURRENT                  = DV_VERSION_CONFIGURABLE_COMPRESSION

	// indicates docvalues type
	DV_NUMERIC                                = 0
	DV_BINARY                                 = 1
	DV_SORTED                                 = 2
	DV_SORTED_SET                             = 3
	DV_SORTED_NUMERIC                         = 4
	DV_DIRECT_MONOTONIC_BLOCK_SHIFT           = 16
	DV_NUMERIC_BLOCK_SHIFT                    = 14
	DV_NUMERIC_BLOCK_SIZE                     = 1 << DV_NUMERIC_BLOCK_SHIFT
	DV_BINARY_BLOCK_SHIFT                     = 5
	DV_BINARY_DOCS_PER_COMPRESSED_BLOCK       = 1 << DV_BINARY_BLOCK_SHIFT
	DV_TERMS_DICT_BLOCK_SHIFT                 = 4
	DV_TERMS_DICT_BLOCK_SIZE                  = 1 << DV_TERMS_DICT_BLOCK_SHIFT
	DV_TERMS_DICT_BLOCK_MASK                  = DV_TERMS_DICT_BLOCK_SIZE - 1
	DV_TERMS_DICT_BLOCK_COMPRESSION_THRESHOLD = 32
	DV_TERMS_DICT_BLOCK_LZ4_SHIFT             = 6
	DV_TERMS_DICT_BLOCK_LZ4_SIZE              = 1 << DV_TERMS_DICT_BLOCK_LZ4_SHIFT
	DV_TERMS_DICT_BLOCK_LZ4_MASK              = DV_TERMS_DICT_BLOCK_LZ4_SIZE - 1
	DV_TERMS_DICT_COMPRESSOR_LZ4_CODE         = 1
	DV_                                       // Writing a special code so we know this is a LZ4-compressed block.
	DV_TERMS_DICT_BLOCK_LZ4_CODE              = DV_TERMS_DICT_BLOCK_LZ4_SHIFT<<16 | DV_TERMS_DICT_COMPRESSOR_LZ4_CODE
	DV_TERMS_DICT_REVERSE_INDEX_SHIFT         = 10
	DV_TERMS_DICT_REVERSE_INDEX_SIZE          = 1 << DV_TERMS_DICT_REVERSE_INDEX_SHIFT
	DV_TERMS_DICT_REVERSE_INDEX_MASK          = DV_TERMS_DICT_REVERSE_INDEX_SIZE - 1
)
