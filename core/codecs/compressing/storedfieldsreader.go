package compressing

import (
	"bytes"
	"context"
	"errors"

	"github.com/geange/lucene-go/core/codecs"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
)

var _ index.StoredFieldsReader = &StoredFieldsReader{}

type StoredFieldsReader struct {
	version           int
	fieldInfos        index.FieldInfos
	indexReader       FieldsIndex
	maxPointer        int64
	fieldsStream      store.IndexInput
	chunkSize         int
	packedIntsVersion int
	compressionMode   CompressionMode
	decompressor      Decompressor
	numDocs           int
	merging           bool
	state             *BlockState
	numChunks         int64 // number of written blocks
	numDirtyChunks    int64 // number of incomplete compressed blocks written
	numDirtyDocs      int64 // cumulative number of docs in incomplete chunks
	closed            bool
}

func NewStoredFieldsReader(ctx context.Context, d store.Directory, si index.SegmentInfo, segmentSuffix string, fn index.FieldInfos,
	context *store.IOContext, formatName string, compressionMode CompressionMode) (*StoredFieldsReader, error) {

	segment := si.Name()

	numDocs, err := si.MaxDoc()
	if err != nil {
		return nil, err
	}

	reader := &StoredFieldsReader{
		compressionMode: compressionMode,
		fieldInfos:      fn,
		numDocs:         numDocs,
	}

	fieldsStreamFN := store.SegmentFileName(segment, segmentSuffix, FIELDS_EXTENSION)

	// Open the data file
	fieldsStream, err := d.OpenInput(ctx, fieldsStreamFN)
	if err != nil {
		return nil, err
	}
	version, err := codecs.CheckIndexHeader(ctx, fieldsStream, formatName, VERSION_START, VERSION_CURRENT, si.GetID(), segmentSuffix)
	if err != nil {
		return nil, err
	}

	var metaIn store.ChecksumIndexInput
	if version >= VERSION_OFFHEAP_INDEX {
		metaStreamFN := store.SegmentFileName(segment, segmentSuffix, META_EXTENSION)
		metaIn, err = store.OpenChecksumInput(ctx, d, metaStreamFN)
		if err != nil {
			return nil, err
		}
		_, err = codecs.CheckIndexHeader(ctx, metaIn, INDEX_CODEC_NAME+"Meta", META_VERSION_START, version, si.GetID(), segmentSuffix)
		if err != nil {
			return nil, err
		}
	}
	if version >= VERSION_META {
		chunkSize, err := metaIn.ReadUvarint(ctx)
		if err != nil {
			return nil, err
		}
		reader.chunkSize = int(chunkSize)

		packedIntsVersion, err := metaIn.ReadUvarint(ctx)
		if err != nil {
			return nil, err
		}
		reader.packedIntsVersion = int(packedIntsVersion)
	} else {
		chunkSize, err := fieldsStream.ReadUvarint(ctx)
		if err != nil {
			return nil, err
		}
		reader.chunkSize = int(chunkSize)

		packedIntsVersion, err := fieldsStream.ReadUvarint(ctx)
		if err != nil {
			return nil, err
		}
		reader.packedIntsVersion = int(packedIntsVersion)
	}

	reader.decompressor = compressionMode.NewDecompressor()
	reader.merging = false
	reader.state = reader.newBlockState()

	// NOTE: data file is too costly to verify checksum against all the bytes on open,
	// but for now we at least verify proper structure of the checksum footer: which looks
	// for FOOTER_MAGIC + algorithmID. This is cheap and can detect some forms of corruption
	// such as file truncation.
	if _, err := codecs.RetrieveChecksum(ctx, fieldsStream); err != nil {
		return nil, err
	}

	maxPointer := int64(-1)
	var indexReader FieldsIndex

	if version < VERSION_OFFHEAP_INDEX {
		// Load the index into memory
		indexName := store.SegmentFileName(segment, segmentSuffix, "fdx")
		indexStream, err := store.OpenChecksumInput(ctx, d, indexName)
		if err != nil {
			return nil, err
		}

		codecNameIdx := formatName[0:len(formatName)-len("Data")] + "Index"
		version2, err := codecs.CheckIndexHeader(ctx, indexStream, codecNameIdx, VERSION_START, VERSION_CURRENT, si.GetID(), segmentSuffix)
		if version != version2 {
			return nil, errors.New("version mismatch between stored fields index and data")
		}
		indexReader, err = NewLegacyFieldsIndexReader(ctx, indexStream, si)
		if err != nil {
			return nil, err
		}
		maxPointerU64, err := indexStream.ReadUvarint(ctx)
		if err != nil {
			return nil, err
		}
		maxPointer = int64(maxPointerU64)

		if _, err := codecs.CheckFooter(ctx, indexStream); err != nil {
			return nil, err
		}
	} else {
		fieldsIndexReader, err := NewFieldsIndexReader(ctx, d, si.Name(), segmentSuffix, INDEX_EXTENSION, INDEX_CODEC_NAME, si.GetID(), metaIn)
		if err != nil {
			return nil, err
		}
		indexReader = fieldsIndexReader
		maxPointer = int64(fieldsIndexReader.GetMaxPointer())
	}

	reader.maxPointer = maxPointer
	reader.indexReader = indexReader

	if version >= VERSION_NUM_CHUNKS {
		numChunks, err := metaIn.ReadUvarint(ctx)
		if err != nil {
			return nil, err
		}
		reader.numChunks = int64(numChunks)

		numDirtyChunks, err := metaIn.ReadUvarint(ctx)
		if err != nil {
			return nil, err
		}
		reader.numDirtyChunks = int64(numDirtyChunks)

		numDirtyDocs, err := metaIn.ReadUvarint(ctx)
		if err != nil {
			return nil, err
		}
		reader.numDirtyDocs = int64(numDirtyDocs)

	} else {
		if version >= VERSION_META {
			// consume dirty chunks/docs stats we wrote
			if _, err := metaIn.ReadUvarint(ctx); err != nil {
				return nil, err
			}
			if _, err := metaIn.ReadUvarint(ctx); err != nil {
				return nil, err
			}
		}
		// Old versions of this format did not record these. Since bulk
		// merges are disabled on version increments anyway, we make no effort
		// to get valid values for these stats.
		reader.numChunks = -1
		reader.numDirtyChunks = -1
		reader.numDirtyDocs = -1
	}

	if reader.numChunks < reader.numDirtyChunks {
		return nil, errors.New("cannot have more dirty chunks than chunks")
	}
	if (reader.numDirtyChunks == 0) != (reader.numDirtyDocs == 0) {
		return nil, errors.New("cannot have dirty chunks without dirty docs or vice-versa")
	}
	if reader.numDirtyDocs < reader.numDirtyChunks {
		return nil, errors.New("cannot have more dirty chunks than documents within dirty chunks")
	}

	if metaIn != nil {
		if _, err := codecs.CheckFooter(ctx, metaIn); err != nil {
			return nil, err
		}
		if err := metaIn.Close(); err != nil {
			return nil, err
		}
	}
	return reader, nil
}

func newStoredFieldsReader(reader *StoredFieldsReader, merging bool) (*StoredFieldsReader, error) {
	this := &StoredFieldsReader{}
	this.version = reader.version
	this.fieldInfos = reader.fieldInfos
	this.fieldsStream = reader.fieldsStream.Clone().(store.IndexInput)
	indexReader, err := reader.indexReader.Clone()
	if err != nil {
		return nil, err
	}
	this.indexReader = indexReader
	this.maxPointer = reader.maxPointer
	this.chunkSize = reader.chunkSize
	this.packedIntsVersion = reader.packedIntsVersion
	this.compressionMode = reader.compressionMode
	this.decompressor = reader.decompressor.Clone()
	this.numDocs = reader.numDocs
	this.numChunks = reader.numChunks
	this.numDirtyChunks = reader.numDirtyChunks
	this.numDirtyDocs = reader.numDirtyDocs
	this.merging = merging
	this.state = this.newBlockState()
	this.closed = false

	return this, nil
}

func (s *StoredFieldsReader) Close() error {
	if !s.closed {
		if err := s.indexReader.Close(); err != nil {
			return err
		}
		if err := s.fieldsStream.Close(); err != nil {
			return err
		}
		s.closed = true
	}
	return nil
}

func (s *StoredFieldsReader) VisitDocument(ctx context.Context, docID int, visitor document.StoredFieldVisitor) error {
	doc, err := s.document(docID)
	if err != nil {
		return err
	}

	for fieldIDX := 0; fieldIDX < doc.numStoredFields; fieldIDX++ {
		infoAndBits, err := doc.in.ReadUvarint(ctx)
		if err != nil {
			return err
		}
		fieldNumber := infoAndBits >> TYPE_BITS
		fieldInfo := s.fieldInfos.FieldInfoByNumber(int(fieldNumber))

		bits := infoAndBits & TYPE_MASK

		needsField, err := visitor.NeedsField(fieldInfo)
		if err != nil {
			return err
		}

		switch needsField {
		case document.STORED_FIELD_VISITOR_YES:
			if err := readField(ctx, doc.in, visitor, fieldInfo, bits); err != nil {
				return err
			}
			break
		case document.STORED_FIELD_VISITOR_NO:
			if fieldIDX == doc.numStoredFields-1 { // don't skipField on last field value; treat like STOP
				return nil
			}
			if err := skipField(ctx, doc.in, bits); err != nil {
				return err
			}
			break
		case document.STORED_FIELD_VISITOR_STOP:
			return nil
		}
	}
	return nil
}

func (s *StoredFieldsReader) Clone(ctx context.Context) index.StoredFieldsReader {
	reader, _ := newStoredFieldsReader(s, false)
	return reader
}

func (s *StoredFieldsReader) CheckIntegrity() error {
	if err := s.indexReader.CheckIntegrity(); err != nil {
		return err
	}
	if _, err := codecs.ChecksumEntireFile(context.Background(), s.fieldsStream); err != nil {
		return err
	}
	return nil
}

func (s *StoredFieldsReader) GetMergeInstance() index.StoredFieldsReader {
	reader, _ := newStoredFieldsReader(s, true)
	return reader
}

func (s *StoredFieldsReader) document(docID int) (*SerializedDocument, error) {
	if s.state.contains(docID) == false {
		startPointer, err := s.indexReader.GetStartPointer(docID)
		if err != nil {
			return nil, err
		}
		if _, err := s.fieldsStream.Seek(startPointer, 0); err != nil {
			return nil, err
		}
		if err := s.state.reset(docID); err != nil {
			return nil, err
		}
	}
	return s.state.document(docID)
}

// BlockState
// Keeps state about the current block of documents.
type BlockState struct {
	docBase         int
	chunkDocs       int
	sliced          bool
	offsets         []int
	numStoredFields []int
	startPointer    int
	spare           *bytes.Buffer
	bytes           *bytes.Buffer
}

// Get the serialized representation of the given docID. This docID has to be contained in the current block.
func (s *BlockState) document(docID int) (*SerializedDocument, error) {
	panic("")
}

func (s *BlockState) contains(docID int) bool {
	return docID >= s.docBase && docID < s.docBase+s.chunkDocs
}

// Reset this block so that it stores state for the block that contains the given doc id.
func (s *BlockState) reset(docID int) error {
	return s.doReset(docID)
}

func (s *BlockState) doReset(docID int) error {
	panic("")
}

func (s *StoredFieldsReader) newBlockState() *BlockState {
	state := &BlockState{}

	if s.merging {
		state.spare = new(bytes.Buffer)
		state.bytes = new(bytes.Buffer)
	}
	return state
}

func readField(ctx context.Context, in store.DataInput, visitor document.StoredFieldVisitor,
	info *document.FieldInfo, bits uint64) error {
	switch bits & TYPE_MASK {
	case BYTE_ARR:
		length, err := in.ReadUvarint(ctx)
		if err != nil {
			return err
		}
		data := make([]byte, length)
		if _, err := in.Read(data); err != nil {
			return err
		}
		if err := visitor.BinaryField(info, data); err != nil {
			return err
		}
	case STRING:
		length, err := in.ReadUvarint(ctx)
		if err != nil {
			return err
		}
		data := make([]byte, length)
		if _, err := in.Read(data); err != nil {
			return err
		}
		if err := visitor.StringField(info, data); err != nil {
			return err
		}
	case NUMERIC_INT:
		num, err := in.ReadZInt32(ctx)
		if err != nil {
			return err
		}
		if err := visitor.Int32Field(info, int32(num)); err != nil {
			return err
		}
	case NUMERIC_FLOAT:
		num, err := ReadZFloat(ctx, in)
		if err != nil {
			return err
		}
		if err := visitor.Float32Field(info, num); err != nil {
			return err
		}
	case NUMERIC_LONG:
		num, err := ReadTLong(ctx, in)
		if err != nil {
			return err
		}
		if err := visitor.Int64Field(info, int64(num)); err != nil {
			return err
		}
	case NUMERIC_DOUBLE:
		num, err := ReadZDouble(ctx, in)
		if err != nil {
			return err
		}
		if err := visitor.Float64Field(info, num); err != nil {
			return err
		}
	default:
		return errors.New("unknown type flag")
	}
	return nil
}

func skipField(ctx context.Context, in store.DataInput, bits uint64) error {
	switch bits & TYPE_MASK {
	case BYTE_ARR, STRING:
		length, err := in.ReadUvarint(ctx)
		if err != nil {
			return err
		}
		if err := in.SkipBytes(ctx, int(length)); err != nil {
			return err
		}
	case NUMERIC_INT:
		if _, err := in.ReadZInt32(ctx); err != nil {
			return err
		}
	case NUMERIC_FLOAT:
		if _, err := ReadZFloat(ctx, in); err != nil {
			return err
		}
	case NUMERIC_LONG:
		if _, err := ReadTLong(ctx, in); err != nil {
			return err
		}
	case NUMERIC_DOUBLE:
		if _, err := ReadZDouble(ctx, in); err != nil {
			return err
		}
	default:
		return errors.New("unknown type flag")
	}
	return nil
}

// SerializedDocument
// A serialized document, you need to decode its input in order to get an actual Document.
type SerializedDocument struct {
	in              store.DataInput // the serialized data
	length          int             // the number of bytes on which the document is encoded
	numStoredFields int             // the number of stored fields
}

func NewSerializedDocument(in store.DataInput, length int, numStoredFields int) *SerializedDocument {
	return &SerializedDocument{in: in, length: length, numStoredFields: numStoredFields}
}
