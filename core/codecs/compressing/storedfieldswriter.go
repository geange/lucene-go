package compressing

import (
	"context"
	"math"

	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/packed"
)

var _ index.StoredFieldsWriter = &StoredFieldsWriter{}

const (
	FIELDS_EXTENSION = "fdt"
	INDEX_EXTENSION  = "fdx"
	META_EXTENSION   = "fdm"
	INDEX_CODEC_NAME = "Lucene85FieldsIndex"

	STRING                = 0x00
	BYTE_ARR              = 0x01
	NUMERIC_INT           = 0x02
	NUMERIC_FLOAT         = 0x03
	NUMERIC_LONG          = 0x04
	NUMERIC_DOUBLE        = 0x05
	VERSION_START         = 1
	VERSION_OFFHEAP_INDEX = 2
	// Version where all metadata were moved to the meta file.
	VERSION_META = 3
	// Version where numChunks is explicitly recorded in meta file and a dirty chunk bit is recorded in each chunk
	VERSION_NUM_CHUNKS = 4
	VERSION_CURRENT    = VERSION_NUM_CHUNKS
	META_VERSION_START = 0

	// for compression of timestamps
	SECOND          = 1000
	HOUR            = 60 * 60 * SECOND
	DAY             = 24 * HOUR
	SECOND_ENCODING = 0x40
	HOUR_ENCODING   = 0x80
	DAY_ENCODING    = 0xC0
)

var (
	TYPE_BITS            = packed.UnsignedBitsRequired(NUMERIC_DOUBLE)
	TYPE_MASK            = packed.MaxValue(TYPE_BITS)
	NEGATIVE_ZERO_FLOAT  = math.Float32bits(-0)
	NEGATIVE_ZERO_DOUBLE = math.Float64bits(-0)
)

type StoredFieldsWriter struct {
	segment         string
	indexWriter     *FieldsIndexWriter
	metaStream      store.IndexOutput
	fieldsStream    store.IndexOutput
	compressor      Compressor
	compressionMode CompressionMode
	chunkSize       int
	maxDocsPerChunk int
	bufferedDocs    store.BufferDataOutput
	numStoredFields []int // number of stored fields
	endOffsets      []int // end offsets in bufferedDocs
	docBase         int   // doc ID at the beginning of the chunk
	numBufferedDocs int   // docBase + numBufferedDocs == current doc ID

	numChunks      int
	numDirtyChunks int // number of incomplete compressed blocks written
	numDirtyDocs   int // cumulative number of missing docs in incomplete chunks
}

func NewStoredFieldsWriter(directory store.Directory, si index.SegmentInfo, segmentSuffix string,
	context *store.IOContext, formatName string, compressionMode CompressionMode,
	chunkSize, maxDocsPerChunk, blockShift int) (*StoredFieldsWriter, error) {

	panic("")
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
