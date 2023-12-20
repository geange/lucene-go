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

// CompoundDirectory A read-only Directory that consists of a view over a compound file.
// See Also: CompoundFormat
// lucene.experimental
type CompoundDirectory interface {
	store.Directory

	// CheckIntegrity Checks consistency of this directory.
	// Note that this may be costly in terms of I/O, e.g. may involve computing a checksum
	// item against large data files.
	CheckIntegrity() error
}

type CompoundDirectoryDefault struct {
}

var (
	ErrUnsupportedOperation = errors.New("unsupported operation exception")
)

func (*CompoundDirectoryDefault) DeleteFile(name string) error {
	return ErrUnsupportedOperation
}

func (*CompoundDirectoryDefault) Rename(source, dest string) error {
	return ErrUnsupportedOperation
}

func (*CompoundDirectoryDefault) SyncMetaData() error {
	return nil
}

func (*CompoundDirectoryDefault) CreateOutput(name string, context *store.IOContext) (store.IndexOutput, error) {
	return nil, ErrUnsupportedOperation
}

func (*CompoundDirectoryDefault) CreateTempOutput(prefix, suffix string,
	context *store.IOContext) (store.IndexOutput, error) {
	return nil, ErrUnsupportedOperation
}

func (*CompoundDirectoryDefault) Sync(names []string) error {
	return ErrUnsupportedOperation
}

func (*CompoundDirectoryDefault) ObtainLock(name string) (store.Lock, error) {
	return nil, ErrUnsupportedOperation
}

// CompoundFormat Encodes/decodes compound files
// lucene.experimental
type CompoundFormat interface {

	// GetCompoundReader Returns a Directory view (read-only) for the compound files in this segment
	GetCompoundReader(dir store.Directory, si *SegmentInfo, context *store.IOContext) (CompoundDirectory, error)

	// Write Packs the provided segment's files into a compound format. All files referenced
	// by the provided SegmentInfo must have CodecUtil.writeIndexHeader and CodecUtil.writeFooter.
	Write(dir store.Directory, si *SegmentInfo, context *store.IOContext) error
}

// DocValuesProducer Abstract API that produces numeric, binary, sorted, sortedset, and sortednumeric docvalues.
// lucene.experimental
type DocValuesProducer interface {
	io.Closer

	// GetNumeric Returns NumericDocValues for this field. The returned instance need not be thread-safe:
	// it will only be used by a single thread.
	GetNumeric(field *document.FieldInfo) (NumericDocValues, error)

	// GetBinary Returns BinaryDocValues for this field. The returned instance need not be thread-safe:
	// it will only be used by a single thread.
	GetBinary(field *document.FieldInfo) (BinaryDocValues, error)

	// GetSorted Returns SortedDocValues for this field. The returned instance need not be
	// thread-safe: it will only be used by a single thread.
	GetSorted(field *document.FieldInfo) (SortedDocValues, error)

	// GetSortedNumeric Returns SortedNumericDocValues for this field. The returned instance
	// need not be thread-safe: it will only be used by a single thread.
	GetSortedNumeric(field *document.FieldInfo) (SortedNumericDocValues, error)

	// GetSortedSet Returns SortedSetDocValues for this field. The returned instance need not
	// be thread-safe: it will only be used by a single thread.
	GetSortedSet(field *document.FieldInfo) (SortedSetDocValues, error)

	// CheckIntegrity Checks consistency of this producer
	// Note that this may be costly in terms of I/O, e.g. may involve computing a checksum item
	// against large data files.
	// lucene.internal
	CheckIntegrity() error

	// Returns an instance optimized for merging. This instance may only be consumed in the thread
	// that called getMergeInstance().
	// The default implementation returns this
}

// DocValuesConsumer Abstract API that consumes numeric, binary and sorted docvalues.
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

type DocValuesFormat interface {
	NamedSPI

	// FieldsConsumer Returns a DocValuesConsumer to write docvalues to the index.
	FieldsConsumer(state *SegmentWriteState) (DocValuesConsumer, error)

	// FieldsProducer Returns a DocValuesProducer to read docvalues from the index.
	// NOTE: by the time this call returns, it must hold open any files it will need to use; else,
	// those files may be deleted. Additionally, required files may be deleted during the execution
	// of this call before there is a chance to open them. Under these circumstances an IOException
	// should be thrown by the implementation. IOExceptions are expected and will automatically
	// cause a retry of the segment opening logic with the newly revised segments.
	FieldsProducer(state *SegmentReadState) (DocValuesProducer, error)
}

type FieldInfosFormat interface {

	// Read the FieldInfos previously written with write.
	Read(directory store.Directory, segmentInfo *SegmentInfo,
		segmentSuffix string, ctx *store.IOContext) (*FieldInfos, error)

	// Write Writes the provided FieldInfos to the directory.
	Write(directory store.Directory, segmentInfo *SegmentInfo,
		segmentSuffix string, infos *FieldInfos, context *store.IOContext) error
}

type LiveDocsFormat interface {

	// ReadLiveDocs Read live docs bits.
	ReadLiveDocs(dir store.Directory, info *SegmentCommitInfo, context *store.IOContext) (util.Bits, error)

	// WriteLiveDocs Persist live docs bits. Use SegmentCommitInfo.getNextDelGen to determine
	// the generation of the deletes file you should write to.
	WriteLiveDocs(bits util.Bits, dir store.Directory,
		info *SegmentCommitInfo, newDelCount int, context *store.IOContext) error

	// Files Records all files in use by this SegmentCommitInfo into the files argument.
	Files(info *SegmentCommitInfo, files map[string]struct{}) (map[string]struct{}, error)
}

// NormsFormat Encodes/decodes per-document score normalization values.
type NormsFormat interface {

	// NormsConsumer Returns a NormsConsumer to write norms to the index.
	NormsConsumer(state *SegmentWriteState) (NormsConsumer, error)

	// NormsProducer Returns a NormsProducer to read norms from the index.
	// NOTE: by the time this call returns, it must hold open any files it will need to use; else,
	// those files may be deleted. Additionally, required files may be deleted during the execution
	// of this call before there is a chance to open them. Under these circumstances an IOException
	// should be thrown by the implementation. IOExceptions are expected and will automatically
	// cause a retry of the segment opening logic with the newly revised segments.
	NormsProducer(state *SegmentReadState) (NormsProducer, error)
}

// PointsFormat Encodes/decodes indexed points.
// lucene.experimental
type PointsFormat interface {

	// FieldsWriter Writes a new segment
	FieldsWriter(state *SegmentWriteState) (PointsWriter, error)

	// FieldsReader Reads a segment. NOTE: by the time this call returns, it must hold open any files
	// it will need to use; else, those files may be deleted. Additionally, required files may be
	// deleted during the execution of this call before there is a chance to open them. Under these
	// circumstances an IOException should be thrown by the implementation. IOExceptions are expected
	// and will automatically cause a retry of the segment opening logic with the newly revised segments.
	FieldsReader(state *SegmentReadState) (PointsReader, error)
}

// PointsReader Abstract API to visit point values.
// lucene.experimental
type PointsReader interface {
	io.Closer

	// CheckIntegrity Checks consistency of this reader.
	// Note that this may be costly in terms of I/O,
	// e.g. may involve computing a checksum item against large data files.
	// lucene.internal
	CheckIntegrity() error

	// GetValues Return PointValues for the given field.
	GetValues(field string) (types.PointValues, error)
}

// GetMergeInstance Returns an instance optimized for merging.
// This instance may only be used in the thread that acquires it.
// The default implementation returns this
//GetMergeInstance() PointsReader

// PostingsFormat Encodes/decodes terms, postings, and proximity data.
// Note, when extending this class, the name (getName) may written into the index in certain
// configurations. In order for the segment to be read, the name must resolve to your
// implementation via forName(String). This method uses Java's Service SortFieldProvider Interface (SPI)
// to resolve format names.
//
// If you implement your own format, make sure that it has a no-arg constructor so SPI can load it.
// ServiceLoader
// lucene.experimental
type PostingsFormat interface {
	NamedSPI

	// FieldsConsumer Writes a new segment
	FieldsConsumer(state *SegmentWriteState) (FieldsConsumer, error)

	// FieldsProducer Reads a segment. NOTE: by the time this call returns, it must hold open any files it
	// will need to use; else, those files may be deleted. Additionally, required files may
	// be deleted during the execution of this call before there is a chance to open them.
	// Under these circumstances an IOException should be thrown by the implementation.
	// IOExceptions are expected and will automatically cause a retry of the segment opening
	// logic with the newly revised segments.
	FieldsProducer(state *SegmentReadState) (FieldsProducer, error)
}

// SegmentInfoFormat Expert: Controls the format of the
// SegmentInfo (segment metadata file).
// @see SegmentInfo
// @lucene.experimental
type SegmentInfoFormat interface {
	// Read {@link SegmentInfo} data from a directory.
	// @param directory directory to read from
	// @param segmentName name of the segment to read
	// @param segmentID expected identifier for the segment
	// @return infos instance to be populated with data
	// @throws IOException If an I/O error occurs
	Read(ctx context.Context, dir store.Directory, segmentName string,
		segmentID []byte, context *store.IOContext) (*SegmentInfo, error)

	// Write {@link SegmentInfo} data.
	// The codec must add its SegmentInfo filename(s) to {@code info} before doing i/o.
	// @throws IOException If an I/O error occurs
	Write(ctx context.Context, dir store.Directory, info *SegmentInfo, ioContext *store.IOContext) error
}

type StoredFieldsFormat interface {

	// FieldsReader Returns a StoredFieldsReader to load stored fields.
	FieldsReader(directory store.Directory, si *SegmentInfo,
		fn *FieldInfos, context *store.IOContext) (StoredFieldsReader, error)

	// FieldsWriter Returns a StoredFieldsWriter to write stored fields.
	FieldsWriter(directory store.Directory,
		si *SegmentInfo, context *store.IOContext) (StoredFieldsWriter, error)
}

type StoredFieldsReader interface {
	io.Closer

	// VisitDocument Visit the stored fields for document docID
	VisitDocument(docID int, visitor document.StoredFieldVisitor) error

	Clone() StoredFieldsReader

	// CheckIntegrity Checks consistency of this reader.
	// Note that this may be costly in terms of I/O, e.g. may involve computing a checksum
	// item against large data files.
	// lucene.internal
	CheckIntegrity() error

	// GetMergeInstance Returns an instance optimized for merging. This instance may not be cloned.
	//The default implementation returns this
	GetMergeInstance() StoredFieldsReader
}

// StoredFieldsWriter Codec API for writing stored fields:
// 1. For every document, startDocument() is called, informing the Codec that a new document has started.
// 2. writeField(FieldInfo, IndexableField) is called for each field in the document.
// 3. After all documents have been written, finish(FieldInfos, int) is called for verification/sanity-checks.
// 4. Finally the writer is closed (close())
// lucene.experimental
type StoredFieldsWriter interface {
	io.Closer

	// StartDocument Called before writing the stored fields of the document. writeField(FieldInfo, IndexableField) will be called for each stored field. Note that this is called even if the document has no stored fields.
	StartDocument() error

	// FinishDocument Called when a document and all its fields have been added.
	FinishDocument() error

	// WriteField Writes a single stored field.
	WriteField(info *document.FieldInfo, field document.IndexableField) error

	// Finish Called before close(), passing in the number of documents that were written. Note that this is intentionally redundant (equivalent to the number of calls to startDocument(), but a Codec should check that this is the case to detect the JRE bug described in LUCENE-1282.
	Finish(fis *FieldInfos, numDocs int) error
}

// TermVectorsFormat Controls the format of term vectors
type TermVectorsFormat interface {

	// VectorsReader Returns a TermVectorsReader to read term vectors.
	VectorsReader(dir store.Directory, segmentInfo *SegmentInfo,
		fieldInfos *FieldInfos, context *store.IOContext) (TermVectorsReader, error)

	// VectorsWriter Returns a TermVectorsWriter to write term vectors.
	VectorsWriter(dir store.Directory,
		segmentInfo *SegmentInfo, context *store.IOContext) (TermVectorsWriter, error)
}

// TermVectorsReader Codec API for reading term vectors:
// lucene.experimental
type TermVectorsReader interface {
	io.Closer

	// Get Returns term vectors for this document, or null if term vectors were not indexed.
	// If offsets are available they are in an OffsetAttribute available from the
	// org.apache.lucene.index.PostingsEnum.
	Get(doc int) (Fields, error)

	// CheckIntegrity Checks consistency of this reader.
	// Note that this may be costly in terms of I/O, e.g. may involve computing a checksum
	// item against large data files.
	// lucene.internal
	CheckIntegrity() error

	// Clone Create a clone that one caller at a time may use to read term vectors.
	Clone() TermVectorsReader

	// GetMergeInstance Returns an instance optimized for merging. This instance may
	// only be consumed in the thread that called getMergeInstance().
	// The default implementation returns this
	GetMergeInstance() TermVectorsReader
}

// TermVectorsWriter Codec API for writing term vectors:
// 1. For every document, startDocument(int) is called, informing the Codec how many fields will be written.
// 2. startField(FieldInfo, int, boolean, boolean, boolean) is called for each field in the document, informing the codec how many terms will be written for that field, and whether or not positions, offsets, or payloads are enabled.
// 3. Within each field, startTerm(BytesRef, int) is called for each term.
// 4. If offsets and/or positions are enabled, then addPosition(int, int, int, BytesRef) will be called for each term occurrence.
// 5. After all documents have been written, finish(FieldInfos, int) is called for verification/sanity-checks.
// 6. Finally the writer is closed (close())
// lucene.experimental
type TermVectorsWriter interface {
	io.Closer

	// StartDocument Called before writing the term vectors of the document. startField(FieldInfo, int, boolean, boolean, boolean) will be called numVectorFields times. Note that if term vectors are enabled, this is called even if the document has no vector fields, in this case numVectorFields will be zero.
	StartDocument(numVectorFields int) error

	// FinishDocument Called after a doc and all its fields have been added.
	FinishDocument() error

	// StartField Called before writing the terms of the field. startTerm(BytesRef, int) will be called numTerms times.
	StartField(info *document.FieldInfo, numTerms int, positions, offsets, payloads bool) error

	// FinishField Called after a field and all its terms have been added.
	FinishField() error

	// StartTerm Adds a term and its term frequency freq. If this field has positions and/or offsets enabled, then addPosition(int, int, int, BytesRef) will be called freq times respectively.
	StartTerm(term []byte, freq int) error

	// FinishTerm Called after a term and all its positions have been added.
	FinishTerm() error

	// AddPosition Adds a term position and offsets
	AddPosition(position, startOffset, endOffset int, payload []byte) error

	// Finish Called before close(), passing in the number of documents that were written. Note that this is intentionally redundant (equivalent to the number of calls to startDocument(int), but a Codec should check that this is the case to detect the JRE bug described in LUCENE-1282.
	Finish(fis *FieldInfos, numDocs int) error
}
