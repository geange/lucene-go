package index

import (
	"context"
	"errors"
	"io"

	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util"
)

// Codec
// Encodes/decodes an inverted index segment.
// Note, when extending this class, the name (getName) is written into the index.
// In order for the segment to be read, the name must resolve to your implementation via forName(String).
// This method uses Java's Service SortFieldProvider Interface (SPI) to resolve codec names.
// If you implement your own codec, make sure that it has a no-arg constructor so SPI can load it.
// See Also: ServiceLoader
type Codec interface {
	Named

	// PostingsFormat
	// Encodes/decodes postings
	PostingsFormat() PostingsFormat

	// DocValuesFormat
	// Encodes/decodes docvalues
	DocValuesFormat() DocValuesFormat

	// StoredFieldsFormat
	// Encodes/decodes stored fields
	StoredFieldsFormat() StoredFieldsFormat

	// TermVectorsFormat
	// Encodes/decodes term vectors
	TermVectorsFormat() TermVectorsFormat

	// FieldInfosFormat
	// Encodes/decodes field infos file
	FieldInfosFormat() FieldInfosFormat

	// SegmentInfoFormat
	// Encodes/decodes segment info file
	SegmentInfoFormat() SegmentInfoFormat

	// NormsFormat
	// Encodes/decodes document normalization values
	NormsFormat() NormsFormat

	// LiveDocsFormat
	// Encodes/decodes live docs
	LiveDocsFormat() LiveDocsFormat

	// CompoundFormat
	// Encodes/decodes compound files
	CompoundFormat() CompoundFormat

	// PointsFormat
	// Encodes/decodes points index
	PointsFormat() PointsFormat
}

type Named interface {
	GetName() string
}

var codesPool = make(map[string]Codec)

func RegisterCodec(codec Codec) {
	codesPool[codec.GetName()] = codec
}

func GetCodecByName(name string) (Codec, bool) {
	codec, exist := codesPool[name]
	return codec, exist
}

// CompoundDirectory
// A read-only Directory that consists of a view over a compound file.
// See Also: CompoundFormat
// lucene.experimental
type CompoundDirectory interface {
	store.Directory

	// CheckIntegrity Checks consistency of this directory.
	// Note that this may be costly in terms of I/O, e.g. may involve computing a checksum
	// item against large data files.
	CheckIntegrity() error
}

type BaseCompoundDirectory struct {
}

var (
	ErrUnsupportedOperation = errors.New("unsupported operation exception")
)

func (*BaseCompoundDirectory) DeleteFile(ctx context.Context, name string) error {
	return ErrUnsupportedOperation
}

func (*BaseCompoundDirectory) Rename(ctx context.Context, source, dest string) error {
	return ErrUnsupportedOperation
}

func (*BaseCompoundDirectory) SyncMetaData(ctx context.Context) error {
	return nil
}

func (*BaseCompoundDirectory) CreateOutput(ctx context.Context, name string) (store.IndexOutput, error) {
	return nil, ErrUnsupportedOperation
}

func (*BaseCompoundDirectory) CreateTempOutput(ctx context.Context, prefix, suffix string) (store.IndexOutput, error) {
	return nil, ErrUnsupportedOperation
}

func (*BaseCompoundDirectory) Sync(ctx context.Context, names []string) error {
	return ErrUnsupportedOperation
}

func (*BaseCompoundDirectory) ObtainLock(name string) (store.Lock, error) {
	return nil, ErrUnsupportedOperation
}

// CompoundFormat
// Encodes/decodes compound files
// lucene.experimental
type CompoundFormat interface {

	// GetCompoundReader Returns a Directory view (read-only) for the compound files in this segment
	GetCompoundReader(ctx context.Context, dir store.Directory, si *SegmentInfo, context *store.IOContext) (CompoundDirectory, error)

	// Write
	// Packs the provided segment's files into a compound format. All files referenced
	// by the provided SegmentInfo must have CodecUtil.writeIndexHeader and CodecUtil.writeFooter.
	Write(ctx context.Context, dir store.Directory, si *SegmentInfo, ioContext *store.IOContext) error
}

// DocValuesProducer
// Abstract API that produces numeric, binary, sorted, sortedset, and sortednumeric docvalues.
// lucene.experimental
type DocValuesProducer interface {
	io.Closer

	// GetNumeric
	// Returns NumericDocValues for this field. The returned instance need not be thread-safe:
	// it will only be used by a single thread.
	GetNumeric(ctx context.Context, field *document.FieldInfo) (NumericDocValues, error)

	// GetBinary
	// Returns BinaryDocValues for this field. The returned instance need not be thread-safe:
	// it will only be used by a single thread.
	GetBinary(ctx context.Context, field *document.FieldInfo) (BinaryDocValues, error)

	// GetSorted
	// Returns SortedDocValues for this field. The returned instance need not be
	// thread-safe: it will only be used by a single thread.
	GetSorted(ctx context.Context, fieldInfo *document.FieldInfo) (SortedDocValues, error)

	// GetSortedNumeric
	// Returns SortedNumericDocValues for this field. The returned instance
	// need not be thread-safe: it will only be used by a single thread.
	GetSortedNumeric(ctx context.Context, field *document.FieldInfo) (SortedNumericDocValues, error)

	// GetSortedSet
	// Returns SortedSetDocValues for this field. The returned instance need not
	// be thread-safe: it will only be used by a single thread.
	GetSortedSet(ctx context.Context, field *document.FieldInfo) (SortedSetDocValues, error)

	// CheckIntegrity
	// Checks consistency of this producer
	// Note that this may be costly in terms of I/O, e.g. may involve computing a checksum item
	// against large data files.
	// lucene.internal
	CheckIntegrity() error

	// GetMergeInstance
	// Returns an instance optimized for merging. This instance may only be consumed in the thread
	// that called GetMergeInstance().
	// The default implementation returns this
	GetMergeInstance() DocValuesProducer
}

// DocValuesConsumer
// Abstract API that consumes numeric, binary and sorted docvalues.
// Concrete implementations of this actually do "something" with the docvalues
// (write it into the index in a specific format).
// The lifecycle is:
//  1. DocValuesConsumer is created by NormsFormat.normsConsumer(SegmentWriteState).
//  2. addNumericField, addBinaryField, addSortedField, addSortedSetField, or addSortedNumericField
//     are called for each Numeric, Binary, Sorted, SortedSet, or SortedNumeric docvalues field.
//     The API is a "pull" rather than "push", and the implementation is free to iterate over the
//     values multiple times (Iterable.iterator()).
//  3. After all fields are added, the consumer is closed.
//
// lucene.experimental
type DocValuesConsumer interface {
	io.Closer

	// AddNumericField Writes numeric docvalues for a field.
	// @param field field information
	// @param valuesProducer Numeric values to write.
	// @throws IOException if an I/O error occurred.
	AddNumericField(ctx context.Context, field *document.FieldInfo, valuesProducer DocValuesProducer) error

	// AddBinaryField Writes binary docvalues for a field.
	// @param field field information
	// @param valuesProducer Binary values to write.
	// @throws IOException if an I/O error occurred.
	AddBinaryField(ctx context.Context, field *document.FieldInfo, valuesProducer DocValuesProducer) error

	// AddSortedField Writes pre-sorted binary docvalues for a field.
	// @param field field information
	// @param valuesProducer produces the values and ordinals to write
	// @throws IOException if an I/O error occurred.
	AddSortedField(ctx context.Context, field *document.FieldInfo, valuesProducer DocValuesProducer) error

	// AddSortedNumericField Writes pre-sorted numeric docvalues for a field
	// @param field field information
	// @param valuesProducer produces the values to write
	// @throws IOException if an I/O error occurred.
	AddSortedNumericField(ctx context.Context, field *document.FieldInfo, valuesProducer DocValuesProducer) error

	// AddSortedSetField Writes pre-sorted set docvalues for a field
	// @param field field information
	// @param valuesProducer produces the values to write
	// @throws IOException if an I/O error occurred.
	AddSortedSetField(ctx context.Context, field *document.FieldInfo, valuesProducer DocValuesProducer) error
}

// Merges in the fields from the readers in mergeState. The default implementation calls mergeNumericField,
// mergeBinaryField, mergeSortedField, mergeSortedSetField, or mergeSortedNumericField for each field,
// depending on its type. Implementations can override this method for more sophisticated merging
// (bulk-byte copying, etc).

// DocValuesFormat
// Encodes/decodes per-document values.
// Note, when extending this class, the name (getName) may written into the index in certain configurations.
// In order for the segment to be read, the name must resolve to your implementation via forName(String).
// This method uses Java's Service Provider Interface (SPI) to resolve format names.
// If you implement your own format, make sure that it has a no-arg constructor so SPI can load it.
// lucene.experimental
type DocValuesFormat interface {
	Named

	// FieldsConsumer Returns a DocValuesConsumer to write docvalues to the index.
	FieldsConsumer(ctx context.Context, state *SegmentWriteState) (DocValuesConsumer, error)

	// FieldsProducer Returns a DocValuesProducer to read docvalues from the index.
	// NOTE: by the time this call returns, it must hold open any files it will need to use; else,
	// those files may be deleted. Additionally, required files may be deleted during the execution
	// of this call before there is a chance to open them. Under these circumstances an IOException
	// should be thrown by the implementation. IOExceptions are expected and will automatically
	// cause a retry of the segment opening logic with the newly revised segments.
	FieldsProducer(ctx context.Context, state *SegmentReadState) (DocValuesProducer, error)
}

// FieldInfosFormat
// Encodes/decodes FieldInfos
// lucene.experimental
type FieldInfosFormat interface {

	// Read
	// Read the FieldInfos previously written with write.
	Read(ctx context.Context, directory store.Directory, segmentInfo *SegmentInfo, segmentSuffix string, ioContext *store.IOContext) (*FieldInfos, error)

	// Write
	// Writes the provided FieldInfos to the directory.
	Write(ctx context.Context, directory store.Directory, segmentInfo *SegmentInfo, segmentSuffix string, infos *FieldInfos, ioContext *store.IOContext) error
}

type LiveDocsFormat interface {

	// ReadLiveDocs
	// Read live docs bits.
	ReadLiveDocs(ctx context.Context, directory store.Directory, info *SegmentCommitInfo, context *store.IOContext) (util.Bits, error)

	// WriteLiveDocs
	// Persist live docs bits. Use SegmentCommitInfo.getNextDelGen to determine
	// the generation of the deletes file you should write to.
	WriteLiveDocs(ctx context.Context, bits util.Bits, directory store.Directory, info *SegmentCommitInfo, newDelCount int, ioContext *store.IOContext) error

	// Files
	// Records all files in use by this SegmentCommitInfo into the files argument.
	Files(ctx context.Context, info *SegmentCommitInfo, files map[string]struct{}) (map[string]struct{}, error)
}

// NormsFormat
// Encodes/decodes per-document score normalization values.
type NormsFormat interface {

	// NormsConsumer
	// Returns a NormsConsumer to write norms to the index.
	NormsConsumer(ctx context.Context, state *SegmentWriteState) (NormsConsumer, error)

	// NormsProducer
	// Returns a NormsProducer to read norms from the index.
	// NOTE: by the time this call returns, it must hold open any files it will need to use; else,
	// those files may be deleted. Additionally, required files may be deleted during the execution
	// of this call before there is a chance to open them. Under these circumstances an IOException
	// should be thrown by the implementation. IOExceptions are expected and will automatically
	// cause a retry of the segment opening logic with the newly revised segments.
	NormsProducer(ctx context.Context, state *SegmentReadState) (NormsProducer, error)
}

// PointsFormat
// Encodes/decodes indexed points.
// lucene.experimental
type PointsFormat interface {

	// FieldsWriter
	// Writes a new segment
	FieldsWriter(ctx context.Context, state *SegmentWriteState) (PointsWriter, error)

	// FieldsReader
	// Reads a segment. NOTE: by the time this call returns, it must hold open any files
	// it will need to use; else, those files may be deleted. Additionally, required files may be
	// deleted during the execution of this call before there is a chance to open them. Under these
	// circumstances an IOException should be thrown by the implementation. IOExceptions are expected
	// and will automatically cause a retry of the segment opening logic with the newly revised segments.
	FieldsReader(ctx context.Context, state *SegmentReadState) (PointsReader, error)
}

// PointsReader
// Abstract API to visit point values.
// lucene.experimental
type PointsReader interface {
	io.Closer

	// CheckIntegrity
	// Checks consistency of this reader.
	// Note that this may be costly in terms of I/O,
	// e.g. may involve computing a checksum item against large data files.
	// lucene.internal
	CheckIntegrity() error

	// GetValues
	// Return PointValues for the given field.
	GetValues(ctx context.Context, field string) (types.PointValues, error)

	// GetMergeInstance
	// Returns an instance optimized for merging.
	// This instance may only be used in the thread that acquires it.
	// The default implementation returns this
	GetMergeInstance() PointsReader
}

// GetMergeInstance Returns an instance optimized for merging.
// This instance may only be used in the thread that acquires it.
// The default implementation returns this
//GetMergeInstance() PointsReader

// PostingsFormat
// Encodes/decodes terms, postings, and proximity data.
// Note, when extending this class, the name (getName) may written into the index in certain
// configurations. In order for the segment to be read, the name must resolve to your
// implementation via forName(String). This method uses Java's Service SortFieldProvider Interface (SPI)
// to resolve format names.
//
// If you implement your own format, make sure that it has a no-arg constructor so SPI can load it.
// ServiceLoader
// lucene.experimental
type PostingsFormat interface {
	Named

	// FieldsConsumer
	// Writes a new segment
	FieldsConsumer(ctx context.Context, state *SegmentWriteState) (FieldsConsumer, error)

	// FieldsProducer
	// Reads a segment. NOTE: by the time this call returns, it must hold open any files it
	// will need to use; else, those files may be deleted. Additionally, required files may
	// be deleted during the execution of this call before there is a chance to open them.
	// Under these circumstances an IOException should be thrown by the implementation.
	// IOExceptions are expected and will automatically cause a retry of the segment opening
	// logic with the newly revised segments.
	FieldsProducer(ctx context.Context, state *SegmentReadState) (FieldsProducer, error)
}

// SegmentInfoFormat
// Expert: Controls the format of the
// SegmentInfo (segment metadata file).
// @see SegmentInfo
// @lucene.experimental
type SegmentInfoFormat interface {
	// Read SegmentInfo data from a directory.
	// dir: directory to read from
	// segmentName: name of the segment to read
	// segmentID: expected identifier for the segment
	Read(ctx context.Context, directory store.Directory, segmentName string,
		segmentID []byte, ioContext *store.IOContext) (*SegmentInfo, error)

	// Write SegmentInfo data.
	// The codec must add its SegmentInfo filename(s) to {@code info} before doing i/o.
	Write(ctx context.Context, directory store.Directory, info *SegmentInfo, ioContext *store.IOContext) error
}

// StoredFieldsFormat
// Controls the format of stored fields
type StoredFieldsFormat interface {

	// FieldsReader
	// Returns a StoredFieldsReader to load stored fields.
	FieldsReader(ctx context.Context, directory store.Directory, si *SegmentInfo, fn *FieldInfos, ioContext *store.IOContext) (StoredFieldsReader, error)

	// FieldsWriter
	// Returns a StoredFieldsWriter to write stored fields.
	FieldsWriter(ctx context.Context, directory store.Directory, si *SegmentInfo, ioContext *store.IOContext) (StoredFieldsWriter, error)
}

// StoredFieldsReader
// Codec API for reading stored fields.
// You need to implement VisitDocument(int, StoredFieldVisitor) to read the stored fields for a document,
// implement Clone() (creating clones of any IndexInputs used, etc), and Close()
// lucene.experimental
type StoredFieldsReader interface {
	io.Closer

	// VisitDocument
	// Visit the stored fields for document docID
	VisitDocument(ctx context.Context, docID int, visitor document.StoredFieldVisitor) error

	Clone(ctx context.Context) StoredFieldsReader

	// CheckIntegrity
	// Checks consistency of this reader.
	// Note that this may be costly in terms of I/O, e.g. may involve computing a checksum
	// item against large data files.
	// lucene.internal
	CheckIntegrity() error

	// GetMergeInstance
	// Returns an instance optimized for merging. This instance may not be cloned.
	// The default implementation returns this
	GetMergeInstance() StoredFieldsReader
}

// StoredFieldsWriter
// Codec API for writing stored fields:
// 1. For every document, startDocument() is called, informing the Codec that a new document has started.
// 2. writeField(FieldInfo, IndexableField) is called for each field in the document.
// 3. After all documents have been written, finish(FieldInfos, int) is called for verification/sanity-checks.
// 4. Finally the writer is closed (close())
// lucene.experimental
type StoredFieldsWriter interface {
	io.Closer

	// StartDocument
	// Called before writing the stored fields of the document.
	// writeField(FieldInfo, IndexableField) will be called for each stored field.
	// Note that this is called even if the document has no stored fields.
	StartDocument(ctx context.Context) error

	// FinishDocument
	// Called when a document and all its fields have been added.
	FinishDocument(ctx context.Context) error

	// WriteField
	// Writes a single stored field.
	WriteField(ctx context.Context, fieldInfo *document.FieldInfo, field document.IndexableField) error

	// Finish
	// Called before close(), passing in the number of documents that were written.
	// Note that this is intentionally redundant (equivalent to the number of calls to startDocument(),
	// but a Codec should check that this is the case to detect the JRE bug described in LUCENE-1282.
	Finish(ctx context.Context, fieldInfos *FieldInfos, numDocs int) error
}

// TermVectorsFormat
// Controls the format of term vectors
type TermVectorsFormat interface {

	// VectorsReader
	// Returns a TermVectorsReader to read term vectors.
	VectorsReader(ctx context.Context, directory store.Directory, segmentInfo *SegmentInfo, fieldInfos *FieldInfos, ioContext *store.IOContext) (TermVectorsReader, error)

	// VectorsWriter
	// Returns a TermVectorsWriter to write term vectors.
	VectorsWriter(ctx context.Context, directory store.Directory, segmentInfo *SegmentInfo, ioContext *store.IOContext) (TermVectorsWriter, error)
}

// TermVectorsReader
// Codec API for reading term vectors:
// lucene.experimental
type TermVectorsReader interface {
	io.Closer

	// Get
	// Returns term vectors for this document, or null if term vectors were not indexed.
	// If offsets are available they are in an OffsetAttribute available from the
	// org.apache.lucene.index.PostingsEnum.
	Get(ctx context.Context, doc int) (Fields, error)

	// CheckIntegrity
	// Checks consistency of this reader.
	// Note that this may be costly in terms of I/O, e.g. may involve computing a checksum
	// item against large data files.
	// lucene.internal
	CheckIntegrity() error

	// Clone Create a clone that one caller at a time may use to read term vectors.
	Clone(ctx context.Context) TermVectorsReader

	// GetMergeInstance Returns an instance optimized for merging. This instance may
	// only be consumed in the thread that called getMergeInstance().
	// The default implementation returns this
	GetMergeInstance() TermVectorsReader
}

// TermVectorsWriter
// Codec API for writing term vectors:
//  1. For every document, startDocument(int) is called, informing the Codec how many fields will be written.
//  2. startField(FieldInfo, int, boolean, boolean, boolean) is called for each field in the document,
//     informing the codec how many terms will be written for that field, and whether or not positions,
//     offsets, or payloads are enabled.
//  3. Within each field, startTerm(BytesRef, int) is called for each term.
//  4. If offsets and/or positions are enabled, then addPosition(int, int, int, BytesRef) will be called
//     for each term occurrence.
//  5. After all documents have been written, finish(FieldInfos, int) is called for verification/sanity-checks.
//  6. Finally the writer is closed (close())
//
// lucene.experimental
type TermVectorsWriter interface {
	io.Closer

	// StartDocument
	// Called before writing the term vectors of the document.
	// startField(FieldInfo, int, boolean, boolean, boolean) will be called numVectorFields times.
	// Note that if term vectors are enabled, this is called even if the document has no vector fields,
	// in this case numVectorFields will be zero.
	StartDocument(ctx context.Context, numVectorFields int) error

	// FinishDocument
	// Called after a doc and all its fields have been added.
	FinishDocument(ctx context.Context) error

	// StartField
	// Called before writing the terms of the field. startTerm(BytesRef, int) will be called numTerms times.
	StartField(ctx context.Context, fieldInfo *document.FieldInfo, numTerms int, positions, offsets, payloads bool) error

	// FinishField
	// Called after a field and all its terms have been added.
	FinishField(ctx context.Context) error

	// StartTerm
	// Adds a term and its term frequency freq. If this field has positions and/or offsets enabled,
	// then addPosition(int, int, int, BytesRef) will be called freq times respectively.
	StartTerm(ctx context.Context, term []byte, freq int) error

	// FinishTerm
	// Called after a term and all its positions have been added.
	FinishTerm(ctx context.Context) error

	// AddPosition
	// Adds a term position and offsets
	AddPosition(ctx context.Context, position, startOffset, endOffset int, payload []byte) error

	// Finish
	// Called before close(), passing in the number of documents that were written.
	// Note that this is intentionally redundant (equivalent to the number of calls to startDocument(int),
	// but a Codec should check that this is the case to detect the JRE bug described in LUCENE-1282.
	Finish(ctx context.Context, fieldInfos *FieldInfos, numDocs int) error
}
