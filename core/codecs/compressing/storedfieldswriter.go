package compressing

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"slices"

	"github.com/geange/lucene-go/core/codecs"
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
	bufferedDocs    *store.BufferDataOutput
	numStoredFields []int // number of stored fields
	endOffsets      []int // end offsets in bufferedDocs
	docBase         int   // doc ID at the beginning of the chunk
	numBufferedDocs int   // docBase + numBufferedDocs == current doc ID

	numChunks      int
	numDirtyChunks int // number of incomplete compressed blocks written
	numDirtyDocs   int // cumulative number of missing docs in incomplete chunks

	numStoredFieldsInDoc int
}

func NewStoredFieldsWriter(ctx context.Context, directory store.Directory, si index.SegmentInfo, segmentSuffix string,
	context *store.IOContext, formatName string, compressionMode CompressionMode,
	chunkSize, maxDocsPerChunk, blockShift int) (*StoredFieldsWriter, error) {

	writer := &StoredFieldsWriter{
		segment:         si.Name(),
		compressionMode: compressionMode,
		compressor:      compressionMode.NewCompressor(),
		chunkSize:       chunkSize,
		maxDocsPerChunk: maxDocsPerChunk,
		docBase:         0,
		bufferedDocs:    store.NewBufferDataOutput(),
		numStoredFields: make([]int, 16),
		endOffsets:      make([]int, 16),
		numBufferedDocs: 0,
	}

	metaStream, err := directory.CreateOutput(ctx, store.SegmentFileName(writer.segment, segmentSuffix, META_EXTENSION))
	if err != nil {
		return nil, err
	}
	writer.metaStream = metaStream
	err = codecs.WriteIndexHeader(ctx, metaStream, INDEX_CODEC_NAME+"Meta", VERSION_CURRENT, si.GetID(), segmentSuffix)
	if err != nil {
		return nil, err
	}

	fieldsStream, err := directory.CreateOutput(ctx, store.SegmentFileName(writer.segment, segmentSuffix, FIELDS_EXTENSION))
	if err != nil {
		return nil, err
	}
	writer.fieldsStream = fieldsStream
	err = codecs.WriteIndexHeader(ctx, fieldsStream, formatName, VERSION_CURRENT, si.GetID(), segmentSuffix)
	if err != nil {
		return nil, err
	}
	indexWriter := NewFieldsIndexWriter(ctx, directory, writer.segment, segmentSuffix, INDEX_EXTENSION, INDEX_CODEC_NAME,
		si.GetID(), blockShift, context)
	writer.indexWriter = indexWriter

	err = metaStream.WriteUvarint(ctx, uint64(chunkSize))
	if err != nil {
		return nil, err
	}
	err = metaStream.WriteUvarint(ctx, uint64(packed.VERSION_CURRENT))
	if err != nil {
		return nil, err
	}

	return writer, nil
}

func (s *StoredFieldsWriter) Close() error {
	//TODO implement me
	panic("implement me")
}

func (s *StoredFieldsWriter) StartDocument(ctx context.Context) error {
	return nil
}

func (s *StoredFieldsWriter) FinishDocument(ctx context.Context) error {
	size := s.numBufferedDocs + 1
	slices.Grow(s.numStoredFields, size)
	slices.Grow(s.endOffsets, size)

	s.numStoredFields[s.numBufferedDocs] = s.numStoredFieldsInDoc
	s.numStoredFieldsInDoc = 0
	s.endOffsets[s.numBufferedDocs] = s.bufferedDocs.Size()
	s.numBufferedDocs++
	if s.triggerFlush() {
		if err := s.flush(ctx, false); err != nil {
			return err
		}
	}
	return nil
}

func (s *StoredFieldsWriter) triggerFlush() bool {
	return s.bufferedDocs.Size() >= s.chunkSize || // chunks of at least chunkSize bytes
		s.numBufferedDocs >= s.maxDocsPerChunk
}

func (s *StoredFieldsWriter) flush(ctx context.Context, force bool) error {
	s.numChunks++
	if force {
		s.numDirtyChunks++ // incomplete: we had to force this flush
		s.numDirtyDocs += s.numBufferedDocs
	}
	if err := s.indexWriter.writeIndex(ctx, s.numBufferedDocs, s.fieldsStream.GetFilePointer()); err != nil {
		return err
	}

	// transform end offsets into lengths
	lengths := s.endOffsets
	for i := s.numBufferedDocs - 1; i > 0; i-- {
		lengths[i] = s.endOffsets[i] - s.endOffsets[i-1]
	}

	sliced := s.bufferedDocs.Size() >= 2*s.chunkSize
	dirtyChunk := force
	if err := s.writeHeader(ctx, s.docBase, s.numBufferedDocs, s.numStoredFields, lengths, sliced, dirtyChunk); err != nil {
		return err
	}

	// compress stored fields to fieldsStream
	//
	// TODO: do we need to slice it since we already have the slices in the buffer? Perhaps
	// we should use max-block-bits restriction on the buffer itself, then we won't have to check it here.
	content := slices.Clone(s.bufferedDocs.Bytes())
	s.bufferedDocs.Reset()

	if sliced {
		// big chunk, slice it
		for compressed := 0; compressed < len(content); compressed += s.chunkSize {
			data := content[compressed:min(s.chunkSize, len(content)-compressed)]
			if err := s.compressor.Compress(nil, data, s.fieldsStream); err != nil {
				return err
			}
		}
	} else {
		if err := s.compressor.Compress(nil, content, s.fieldsStream); err != nil {
			return err
		}
	}

	// reset
	s.docBase += s.numBufferedDocs
	s.numBufferedDocs = 0
	s.bufferedDocs.Reset()
	return nil
}

func (s *StoredFieldsWriter) WriteField(ctx context.Context, info *document.FieldInfo, field document.IndexableField) error {
	s.numStoredFieldsInDoc++

	bits := 0
	var buf *bytes.Buffer
	var strVar *string

	obj := field.Get()
	switch value := obj.(type) {
	case int32:
		bits = NUMERIC_INT
	case int64:
		bits = NUMERIC_LONG
	case float32:
		bits = NUMERIC_FLOAT
	case float64:
		bits = NUMERIC_DOUBLE
	case []byte:
		bits = BYTE_ARR
		strVar = nil
		buf = new(bytes.Buffer)
		buf.Write(value)
	case string:
		bits = STRING
		strVar = &value
	default:
		return errors.New("cannot store type")
	}

	infoAndBits := uint64(info.Number())<<TYPE_BITS | uint64(bits)
	if err := s.bufferedDocs.WriteUvarint(ctx, infoAndBits); err != nil {
		return err
	}

	switch value := obj.(type) {
	case int32:
		if err := s.bufferedDocs.WriteZInt32(ctx, value); err != nil {
			return err
		}
	case int64:
		if err := WriteTLong(ctx, s.bufferedDocs, value); err != nil {
			return err
		}
	case float32:
		if err := WriteZFloat(ctx, s.bufferedDocs, value); err != nil {
			return err
		}
	case float64:
		if err := WriteZDouble(ctx, s.bufferedDocs, value); err != nil {
			return err
		}
	case []byte:
		if err := s.bufferedDocs.WriteUvarint(ctx, uint64(buf.Len())); err != nil {
			return err
		}
		if _, err := s.bufferedDocs.Write(buf.Bytes()); err != nil {
			return err
		}
	case string:
		if err := s.bufferedDocs.WriteString(ctx, *strVar); err != nil {
			return err
		}
	default:
		return errors.New("cannot store type")
	}
	return nil
}

func (s *StoredFieldsWriter) Finish(ctx context.Context, fieldInfos index.FieldInfos, numDocs int) error {
	if s.numBufferedDocs > 0 {
		if err := s.flush(ctx, true); err != nil {
			return err
		}
	}

	if s.docBase != numDocs {
		return fmt.Errorf("wrote %d docs, finish called with numDocs=%d", s.docBase, numDocs)
	}

	if err := s.indexWriter.finish(ctx, numDocs, s.fieldsStream.GetFilePointer(), s.metaStream); err != nil {
		return err
	}
	if err := s.metaStream.WriteUvarint(ctx, uint64(s.numChunks)); err != nil {
		return err
	}
	if err := s.metaStream.WriteUvarint(ctx, uint64(s.numDirtyChunks)); err != nil {
		return err
	}
	if err := s.metaStream.WriteUvarint(ctx, uint64(s.numDirtyDocs)); err != nil {
		return err
	}
	if err := codecs.WriteFooter(ctx, s.metaStream); err != nil {
		return err
	}
	if err := codecs.WriteFooter(ctx, s.fieldsStream); err != nil {
		return err
	}
	return nil
}

func (s *StoredFieldsWriter) writeHeader(ctx context.Context, docBase int, numBufferedDocs int, fields []int, lengths []int, sliced bool, dirtyChunk bool) error {
	slicedBit := 0
	if sliced {
		slicedBit = 1
	}
	dirtyBit := 0
	if dirtyChunk {
		dirtyBit = 2
	}

	// save docBase and numBufferedDocs
	if err := s.fieldsStream.WriteUvarint(ctx, uint64(docBase)); err != nil {
		return err
	}
	if err := s.fieldsStream.WriteUvarint(ctx, uint64((numBufferedDocs<<2)|dirtyBit|slicedBit)); err != nil {
		return err
	}

	// save numStoredFields
	if err := s.saveInts(ctx, s.numStoredFields, numBufferedDocs, s.fieldsStream); err != nil {
		return err
	}

	// save lengths
	if err := s.saveInts(ctx, lengths, numBufferedDocs, s.fieldsStream); err != nil {
		return err
	}
	return nil
}

func (s *StoredFieldsWriter) saveInts(ctx context.Context, values []int, length int, out store.IndexOutput) error {
	if length == 1 {
		return out.WriteUvarint(ctx, uint64(values[0]))
	}

	allEqual := true
	for i := 0; i < length; i++ {
		if values[i] != values[0] {
			allEqual = false
			break
		}
	}

	if allEqual {
		if err := out.WriteUvarint(ctx, 0); err != nil {
			return err
		}
		if err := out.WriteUvarint(ctx, uint64(values[0])); err != nil {
			return err
		}
		return nil
	}

	maxValue := 0
	for i := 0; i < length; i++ {
		maxValue |= values[i]
	}
	bitsRequired, err := packed.BitsRequired(int64(maxValue))
	if err != nil {
		return err
	}
	if err := out.WriteUvarint(ctx, uint64(bitsRequired)); err != nil {
		return err
	}
	w := packed.GetWriterNoHeader(out, packed.FormatPacked, length, bitsRequired, 1)
	for i := 0; i < length; i++ {
		if err := w.Add(uint64(values[i])); err != nil {
			return err
		}
	}
	return w.Finish()
}
