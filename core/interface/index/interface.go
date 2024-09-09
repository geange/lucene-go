package index

import (
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util"
	"github.com/geange/lucene-go/core/util/attribute"
	"github.com/geange/lucene-go/core/util/automaton"
)

type LeafMetaData interface {
	GetSort() Sort
}

type Term interface {
	Field() string
	Text() string
	Bytes() []byte
}

func TermCompare(a, b Term) int {
	cmp := strings.Compare(a.Field(), b.Field())
	if cmp != 0 {
		return cmp
	}
	return bytes.Compare(a.Bytes(), b.Bytes())
}

// Fields
// Provides a Terms index for fields that have it, and lists which fields do. This is primarily an
// internal/experimental API (see FieldsProducer), although it is also used to expose the set of term
// vectors per document.
type Fields interface {

	// DVFUIterator
	// Returns an iterator that will step through all fields names. This will not return null.
	// DVFUIterator() func() string

	Names() []string

	// Terms
	// Get the Terms for this field. This will return null if the field does not exist.
	Terms(field string) (Terms, error)

	// Size Returns the number of fields or -1 if the number of distinct field names is unknown. If >= 0,
	// iterator will return as many field names.
	Size() int

	// Zero-length Fields array.

}

type Terms interface {
	// Iterator
	// DVFUIterator Returns an iterator that will step through all terms. This method will not return null.
	Iterator() (TermsEnum, error)

	// Intersect
	// Returns a TermsEnum that iterates over all terms and documents that are accepted by the
	// provided CompiledAutomaton. If the startTerm is provided then the returned enum will only return
	// terms > startTerm, but you still must call next() first to get to the first term. Note that the provided
	// startTerm must be accepted by the automaton.
	// This is an expert low-level API and will only work for NORMAL compiled automata. To handle any compiled
	// automata you should instead use CompiledAutomaton.getTermsEnum instead.
	// NOTE: the returned TermsEnum cannot seek
	Intersect(compiled *automaton.CompiledAutomaton, startTerm []byte) (TermsEnum, error)

	// Size
	// Returns the number of terms for this field, or -1 if this measure isn't stored by the codec.
	// Note that, just like other term measures, this measure does not take deleted documents into account.
	Size() (int, error)

	// GetSumTotalTermFreq
	// Returns the sum of TermsEnum.totalTermFreq for all terms in this field. Note that,
	// just like other term measures, this measure does not take deleted documents into account.
	GetSumTotalTermFreq() (int64, error)

	// GetSumDocFreq
	// Returns the sum of TermsEnum.docFreq() for all terms in this field. Note that,
	// just like other term measures, this measure does not take deleted documents into account.
	GetSumDocFreq() (int64, error)

	// GetDocCount
	// Returns the number of documents that have at least one term for this field. Note that,
	// just like other term measures, this measure does not take deleted documents into account.
	GetDocCount() (int, error)

	// HasFreqs
	// Returns true if documents in this field store per-document term frequency (PostingsEnum.freq).
	HasFreqs() bool

	// HasOffsets
	// Returns true if documents in this field store offsets.
	HasOffsets() bool

	// HasPositions
	// Returns true if documents in this field store positions.
	HasPositions() bool

	// HasPayloads Returns true if documents in this field store payloads.
	HasPayloads() bool

	// GetMin
	// Returns the smallest term (in lexicographic order) in the field. Note that, just like other
	// term measures, this measure does not take deleted documents into account. This returns null when
	// there are no terms.
	GetMin() ([]byte, error)

	// GetMax
	// Returns the largest term (in lexicographic order) in the field. Note that, just like other term
	// measures, this measure does not take deleted documents into account. This returns null when there are no terms.
	GetMax() ([]byte, error)
}

// IndexReader is an abstract class, providing an interface for accessing a point-in-time view of an index.
// Any changes made to the index via IndexWriter will not be visible until a new IndexReader is opened.
// It's best to use DirectoryReader.open(IndexWriter) to obtain an IndexReader,
// if your IndexWriter is in-process. When you need to re-open to see changes to the index,
// it's best to use DirectoryReader.openIfChanged(DirectoryReader) since the new reader will share resources
// with the previous one when possible. Search of an index is done entirely through this abstract interface,
// so that any subclass which implements it is searchable.
//
// There are two different types of IndexReaders:
//   - LeafReader: These indexes do not consist of several sub-readers, they are atomic. They support retrieval of
//     stored fields, doc values, terms, and postings.
//   - CompositeReader: Instances (like DirectoryReader) of this reader can only be used to get stored fields from
//     the underlying LeafReaders, but it is not possible to directly retrieve postings. To do that, get the
//     sub-readers via CompositeReader.getSequentialSubReaders.
//
// IndexReader instances for indexes on disk are usually constructed with a call to one of the static
// DirectoryReader.open() methods, e.g. DirectoryReader.open(org.apache.lucene.store.Directory).
//
// DirectoryReader implements the CompositeReader interface, it is not possible to directly get postings.
// For efficiency, in this API documents are often referred to via document numbers, non-negative integers
// which each name a unique document in the index. These document numbers are ephemeral -- they may change
// as documents are added to and deleted from an index. Clients should thus not rely on a given document
// having the same number between sessions.
//
// NOTE: IndexReader instances are completely thread safe, meaning multiple threads can call any of its
// methods, concurrently. If your application requires external synchronization, you should not synchronize
// on the IndexReader instance; use your own (non-Lucene) objects instead.
type IndexReader interface {
	io.Closer

	// GetTermVectors
	// Retrieve term vectors for this document, or null if term vectors were not indexed.
	// The returned Fields instance acts like a single-document inverted index (the docID will be 0).
	GetTermVectors(docID int) (Fields, error)

	// GetTermVector
	// Retrieve term vector for this document and field, or null if term vectors were not indexed.
	// The returned Fields instance acts like a single-document inverted index (the docID will be 0).
	GetTermVector(docID int, field string) (Terms, error)

	// NumDocs
	// Returns the number of documents in this index.
	// NOTE: This operation may run in O(maxDoc). Implementations that can't return this number in
	// constant-time should cache it.
	NumDocs() int

	// MaxDoc
	// Returns one greater than the largest possible document number. This may be used to,
	// e.g., determine how big to allocate an array which will have an element for every document
	// number in an index.
	MaxDoc() int

	// NumDeletedDocs
	// Returns the number of deleted documents.
	// NOTE: This operation may run in O(maxDoc).
	NumDeletedDocs() int

	// Document
	// Returns the stored fields of the nth Document in this index. This is just sugar for using
	// DocumentStoredFieldVisitor.
	// NOTE: for performance reasons, this method does not check if the requested document is deleted, and
	// therefore asking for a deleted document may yield unspecified results. Usually this is not required,
	// however you can test if the doc is deleted by checking the Bits returned from MultiBits.getLiveDocs.
	// NOTE: only the content of a field is returned, if that field was stored during indexing. Metadata like
	// boost, omitNorm, IndexOptions, tokenized, etc., are not preserved.
	// Throws: CorruptIndexException – if the index is corrupt
	// IOException – if there is a low-level IO error
	// TODO: we need a separate StoredField, so that the document returned here contains that class not IndexableField
	Document(ctx context.Context, docID int) (*document.Document, error)

	// DocumentWithVisitor
	// Expert: visits the fields of a stored document, for custom processing/loading of each field.
	// If you simply want to load all fields, use document(int). If you want to load a subset,
	// use DocumentStoredFieldVisitor.
	DocumentWithVisitor(ctx context.Context, docID int, visitor document.StoredFieldVisitor) error

	// DocumentWithFields
	// Like Document(docID int) but only loads the specified fields. Note that this is simply sugar for
	// DocumentStoredFieldVisitor.DocumentStoredFieldVisitor(Set).
	DocumentWithFields(ctx context.Context, docID int, fieldsToLoad []string) (*document.Document, error)

	// HasDeletions
	// Returns true if any documents have been deleted. Implementers should consider overriding
	// this method if maxDoc() or numDocs() are not constant-time operations.
	HasDeletions() bool

	// DoClose
	// Implements close.
	DoClose() error

	// GetContext
	// Expert: Returns the root IndexReaderContext for this IndexReader's sub-reader tree.
	// If this reader is composed of sub readers, i.e. this reader being a composite reader, this method
	// returns a CompositeReaderContext holding the reader's direct children as well as a view of the
	// reader tree's atomic leaf contexts. All sub- IndexReaderContext instances referenced from this
	// readers top-level context are private to this reader and are not shared with another context tree.
	// For example, IndexSearcher uses this API to drive searching by one atomic leaf reader at a time.
	// If this reader is not composed of child readers, this method returns an LeafReaderContextImpl.
	// Note: Any of the sub-CompositeReaderContext instances referenced from this top-level context do
	// not support CompositeReaderContext.leaves(). Only the top-level context maintains the convenience
	// leaf-view for performance reasons.
	GetContext() (ctx IndexReaderContext, err error)

	// Leaves
	// Returns the reader's leaves, or itself if this reader is atomic. This is a convenience method
	// calling this.getContext().leaves().
	// See Also: ReaderContext.leaves()
	Leaves() ([]LeafReaderContext, error)

	// GetReaderCacheHelper
	// Optional method: Return a Reader.CacheHelper that can be used to cache based
	// on the content of this reader. Two readers that have different data or different sets of deleted
	// documents will be considered different.
	// A return item of null indicates that this reader is not suited for caching, which is typically the
	// case for short-lived wrappers that alter the content of the wrapped reader.
	GetReaderCacheHelper() CacheHelper

	// DocFreq
	// Returns the number of documents containing the term. This method returns 0 if the term or field
	// does not exists. This method does not take into account deleted documents that have not yet been merged away.
	// See Also: TermsEnum.docFreq()
	DocFreq(ctx context.Context, term Term) (int, error)

	// TotalTermFreq
	// Returns the total number of occurrences of term across all documents (the sum of the freq() for each
	// doc that has this term). Note that, like other term measures, this measure does not take deleted
	// documents into account.
	TotalTermFreq(ctx context.Context, term Term) (int64, error)

	// GetSumDocFreq
	// Returns the sum of TermsEnum.docFreq() for all terms in this field. Note that, just like
	// other term measures, this measure does not take deleted documents into account.
	// See Also: Terms.getSumDocFreq()
	GetSumDocFreq(field string) (int64, error)

	// GetDocCount
	// Returns the number of documents that have at least one term for this field. Note that,
	// just like other term measures, this measure does not take deleted documents into account.
	// See Also: Terms.getDocCount()
	GetDocCount(field string) (int, error)

	// GetSumTotalTermFreq
	// Returns the sum of TermsEnum.totalTermFreq for all terms in this field. Note that,
	// just like other term measures, this measure does not take deleted documents into account.
	// See Also: Terms.getSumTotalTermFreq()
	GetSumTotalTermFreq(field string) (int64, error)
	//RegisterParentReader(reader Reader)

	GetRefCount() int
	IncRef() error
	DecRef() error
	GetMetaData() LeafMetaData
}

// IndexReaderContext A struct like class that represents a hierarchical relationship between IndexReader instances.
type IndexReaderContext interface {

	// Reader Returns the IndexReader, this context represents.
	Reader() IndexReader

	// Leaves
	// Returns the context's leaves if this context is a top-level context. For convenience, if this is
	// an LeafReaderContextImpl this returns itself as the only leaf.
	// Note: this is convenience method since leaves can always be obtained by walking the context tree
	// using children().
	// Throws: ErrUnsupportedOperation – if this is not a top-level context.
	// See Also: children()
	Leaves() ([]LeafReaderContext, error)

	// Children Returns the context's children iff this context is a composite context otherwise null.
	Children() []IndexReaderContext

	Identity() string
}

type LeafReaderContext interface {
	Reader() IndexReader
	Leaves() ([]LeafReaderContext, error)
	Children() []IndexReaderContext
	Identity() string

	Ord() int
	DocBase() int
	LeafReader() LeafReader
}

type LeafReader interface {
	IndexReader

	// Terms
	// Returns the Terms index for this field, or null if it has none.
	Terms(field string) (Terms, error)

	// Postings
	// Returns PostingsEnum for the specified term. This will return null if either the field or term does not exist.
	// NOTE: The returned PostingsEnum may contain deleted docs.
	// See Also: TermsEnum.postings(PostingsEnum)
	Postings(ctx context.Context, term Term, flags int) (PostingsEnum, error)

	// GetNumericDocValues
	// Returns NumericDocValues for this field, or null if no numeric doc values were
	// indexed for this field. The returned instance should only be used by a single thread.
	GetNumericDocValues(field string) (NumericDocValues, error)

	// GetBinaryDocValues
	// Returns BinaryDocValues for this field, or null if no binary doc values were indexed
	// for this field. The returned instance should only be used by a single thread.
	GetBinaryDocValues(field string) (BinaryDocValues, error)

	// GetSortedDocValues
	// Returns SortedDocValues for this field, or null if no SortedDocValues were indexed
	// for this field. The returned instance should only be used by a single thread.
	GetSortedDocValues(field string) (SortedDocValues, error)

	// GetSortedNumericDocValues
	// Returns SortedNumericDocValues for this field, or null if no
	// SortedNumericDocValues were indexed for this field.
	// The returned instance should only be used by a single thread.
	GetSortedNumericDocValues(field string) (SortedNumericDocValues, error)

	// GetSortedSetDocValues
	// Returns SortedSetDocValues for this field, or null if no SortedSetDocValues
	// were indexed for this field. The returned instance should only be used by a single thread.
	GetSortedSetDocValues(field string) (SortedSetDocValues, error)

	// GetNormValues
	// Returns NumericDocValues representing norms for this field, or null if no NumericDocValues
	// were indexed. The returned instance should only be used by a single thread.
	GetNormValues(field string) (NumericDocValues, error)

	// GetFieldInfos
	// Get the FieldInfos describing all fields in this reader. Note: Implementations
	// should cache the FieldInfos instance returned by this method such that subsequent calls to
	// this method return the same instance.
	GetFieldInfos() FieldInfos

	// GetLiveDocs
	// Returns the Bits representing live (not deleted) docs. A set bit indicates the doc ID has
	// not been deleted. If this method returns null it means there are no deleted documents
	// (all documents are live). The returned instance has been safely published for use by multiple threads
	// without additional synchronization.
	GetLiveDocs() util.Bits

	// GetPointValues
	// Returns the PointValues used for numeric or spatial searches for the given field, or null
	// if there are no point fields.
	GetPointValues(field string) (types.PointValues, bool)

	// CheckIntegrity
	// Checks consistency of this reader.
	// Note that this may be costly in terms of I/O,
	// e.g. may involve computing a checksum item against large data files.
	CheckIntegrity() error

	// GetMetaData
	// Return metadata about this leaf.
	//GetMetaData() LeafMetaData
}

// PostingsEnum
// Iterates through the postings. NOTE: you must first call nextDoc before using any of the
// per-doc methods.
type PostingsEnum interface {
	types.DocIdSetIterator

	// Freq Returns term frequency in the current document, or 1 if the field was indexed with IndexOptions.DOCS.
	// Do not call this before nextDoc is first called, nor after nextDoc returns DocIdSetIterator.NO_MORE_DOCS.
	// NOTE: if the PostingsEnum was obtain with NONE, the result of this method is undefined.
	Freq() (int, error)

	// NextPosition Returns the next position, or -1 if positions were not indexed. Calling this more than freq() times is undefined.
	NextPosition() (int, error)

	// StartOffset Returns start offset for the current position, or -1 if offsets were not indexed.
	StartOffset() (int, error)

	// EndOffset Returns end offset for the current position, or -1 if offsets were not indexed.
	EndOffset() (int, error)

	// GetPayload Returns the payload at this position, or null if no payload was indexed. You should not
	// modify anything (neither members of the returned BytesRef nor bytes in the byte[]).
	GetPayload() ([]byte, error)
}

// NumericDocValues
// A per-document numeric item.
type NumericDocValues interface {
	types.DocValuesIterator

	// LongValue
	// Returns the numeric item for the current document ID. It is illegal to call this method
	// after advanceExact(int) returned false.
	// Returns: numeric item
	LongValue() (int64, error)
}

// BinaryDocValues
// A per-document numeric item.
type BinaryDocValues interface {
	types.DocValuesIterator

	// BinaryValue
	// Returns the binary item for the current document ID. It is illegal to call this method after
	// advanceExact(int) returned false.
	// Returns: binary item
	BinaryValue() ([]byte, error)
}

// SortedDocValues
// A per-document byte[] with presorted values. This is fundamentally an iterator over the
// int ord values per document, with random access APIs to resolve an int ord to BytesRef.
// Per-Document values in a SortedDocValues are deduplicated, dereferenced, and sorted into a dictionary of
// unique values. A pointer to the dictionary item (ordinal) can be retrieved for each document. Ordinals
// are dense and in increasing sorted order.
type SortedDocValues interface {
	BinaryDocValues

	// OrdValue
	// Returns the ordinal for the current docID. It is illegal to call this method after
	// advanceExact(int) returned false.
	// Returns: ordinal for the document: this is dense, starts at 0, then increments by 1 for the
	// next item in sorted order.
	OrdValue() (int, error)

	// LookupOrd
	// Retrieves the item for the specified ordinal. The returned BytesRef may be re-used
	// across calls to lookupOrd(int) so make sure to copy it if you want to keep it around.
	// Params: ord – ordinal to lookup (must be >= 0 and < FnGetValueCount())
	// See Also: FnOrdValue()
	LookupOrd(ord int) ([]byte, error)

	// GetValueCount
	// Returns the number of unique values.
	// Returns: number of unique values in this SortedDocValues. This is also equivalent to one plus the maximum ordinal.
	GetValueCount() int

	// LookupTerm
	// If key exists, returns its ordinal, else returns -insertionPoint-1, like Arrays.binarySearch.
	// key: key to look up
	LookupTerm(key []byte) (int, error)

	// TermsEnum
	// Returns a TermsEnum over the values. The enum supports TermsEnum.ord() and TermsEnum.seekExact(long).
	TermsEnum() (TermsEnum, error)

	// Intersect
	// Returns a TermsEnum over the values, filtered by a CompiledAutomaton The enum supports TermsEnum.ord().
	Intersect(automaton *automaton.CompiledAutomaton) (TermsEnum, error)
}

// SortedNumericDocValues A list of per-document numeric values, sorted according to Long.CompareFn(long, long).
type SortedNumericDocValues interface {
	types.DocValuesIterator

	// NextValue Iterates to the next item in the current document. Do not call this more than
	// docValueCount times for the document.
	NextValue() (int64, error)

	// DocValueCount Retrieves the number of values for the current document. This must always be greater
	// than zero. It is illegal to call this method after advanceExact(int) returned false.
	DocValueCount() int
}

// TermsEnum DVFUIterator to seek (seekCeil(), seekExact()) or step through (next terms to obtain
// frequency information (docFreq), PostingsEnum or PostingsEnum for the current term (postings.
// Term enumerations are always ordered by .compareTo, which is Unicode sort order if the terms are
// UTF-8 bytes. Each term in the enumeration is greater than the one before it.
// The TermsEnum is unpositioned when you first obtain it and you must first successfully call next or one
// of the seek methods.
type TermsEnum interface {
	Next(context.Context) ([]byte, error)

	// Attributes Returns the related attributes.
	Attributes() *attribute.Source

	// SeekExact Attempts to seek to the exact term, returning true if the term is found. If this returns false,
	// the enum is unpositioned. For some codecs, seekExact may be substantially faster than seekCeil.
	// Returns: true if the term is found; return false if the enum is unpositioned.
	SeekExact(ctx context.Context, text []byte) (bool, error)

	// SeekCeil eeks to the specified term, if it exists, or to the next (ceiling) term. Returns SeekStatus to
	// indicate whether exact term was found, a different term was found, or isEof was hit. The target term may be
	// before or after the current term. If this returns SeekStatus.END, the enum is unpositioned.
	SeekCeil(ctx context.Context, text []byte) (SeekStatus, error)

	// SeekExactByOrd Seeks to the specified term by ordinal (position) as previously returned by ord. The
	// target ord may be before or after the current ord, and must be within bounds.
	SeekExactByOrd(ctx context.Context, ord int64) error

	// SeekExactExpert Expert: Seeks a specific position by TermState previously obtained from termState().
	// Callers should maintain the TermState to use this method. Low-level implementations may position the
	// TermsEnum without re-seeking the term dictionary.
	// Seeking by TermState should only be used iff the state was obtained from the same TermsEnum instance.
	// NOTE: Using this method with an incompatible TermState might leave this TermsEnum in undefined state.
	// On a segment level TermState instances are compatible only iff the source and the target TermsEnum operate
	// on the same field. If operating on segment level, TermState instances must not be used across segments.
	// NOTE: A seek by TermState might not restore the AttributeSourceV2's state. AttributeSourceV2 states must be
	// maintained separately if this method is used.
	// Params: 	term – the term the TermState corresponds to
	//			state – the TermState
	SeekExactExpert(ctx context.Context, term []byte, state TermState) error

	// Term Returns current term. Do not call this when the enum is unpositioned.
	Term() ([]byte, error)

	// Ord Returns ordinal position for current term. This is an optional method (the codec may throw
	// ErrUnsupportedOperation). Do not call this when the enum is unpositioned.
	Ord() (int64, error)

	// DocFreq Returns the number of documents containing the current term. Do not call this when the
	// enum is unpositioned. TermsEnum.SeekStatus.END.
	DocFreq() (int, error)

	// TotalTermFreq Returns the total number of occurrences of this term across all documents (the sum of the
	// freq() for each doc that has this term). Note that, like other term measures, this measure does not
	// take deleted documents into account.
	TotalTermFreq() (int64, error)

	// Postings Get PostingsEnum for the current term. Do not call this when the enum is unpositioned. This
	// method will not return null.
	// NOTE: the returned iterator may return deleted documents, so deleted documents have to be checked on top of the PostingsEnum.
	// Use this method if you only require documents and frequencies, and do not need any proximity data. This method is equivalent to postings(reuse, PostingsEnum.FREQS)
	// Params: reuse – pass a prior PostingsEnum for possible reuse
	// See Also: postings(PostingsEnum, int)
	//Postings(reuse PostingsEnum) (PostingsEnum, error)

	// Postings Get PostingsEnum for the current term, with control over whether freqs, positions, offsets or payloads are required. Do not call this when the enum is unpositioned. This method will not return null.
	// NOTE: the returned iterator may return deleted documents, so deleted documents have to be checked on top of the PostingsEnum.
	// Params: 	reuse – pass a prior PostingsEnum for possible reuse
	// 			flags – specifies which optional per-document values you require; see PostingsEnum.FREQS
	Postings(reuse PostingsEnum, flags int) (PostingsEnum, error)

	// Impacts Return a ImpactsEnum.
	// See Also: postings(PostingsEnum, int)
	Impacts(flags int) (ImpactsEnum, error)

	// TermState Expert: Returns the TermsEnums internal state to position the TermsEnum without re-seeking the
	// term dictionary.
	// NOTE: A seek by TermState might not capture the AttributeSourceV2's state. Callers must maintain the
	// AttributeSourceV2 states separately
	// See Also: TermState, seekExact(, TermState)
	TermState() (TermState, error)
}

type Impact interface {
	GetFreq() int
	GetNorm() int64
}

type Impacts interface {

	// NumLevels Return the number of levels on which we have impacts.
	// The returned item is always greater than 0 and may not always be the same,
	// even on a single postings list, depending on the current doc ID.
	NumLevels() int

	// GetDocIdUpTo Return the maximum inclusive doc ID until which the list of impacts
	// returned by getImpacts(int) is valid. This is a non-decreasing function of level.
	GetDocIdUpTo(level int) int

	// GetImpacts Return impacts on the given level. These impacts are sorted by increasing
	// frequency and increasing unsigned norm, and only valid until the doc ID returned by
	// getDocIdUpTo(int) for the same level, included. The returned list is never empty.
	// NOTE: There is no guarantee that these impacts actually appear in postings, only that
	// they trigger scores that are greater than or equal to the impacts that actually
	// appear in postings.
	GetImpacts(level int) []Impact
}

// ImpactsEnum Extension of PostingsEnum which also provides information about upcoming impacts.
type ImpactsEnum interface {
	PostingsEnum
	ImpactsSource
}

// ImpactsSource Source of Impacts.
type ImpactsSource interface {
	// AdvanceShallow Shallow-advance to target. This is cheaper than calling DocIdSetIterator.advance(int)
	// and allows further calls to getImpacts() to ignore doc IDs that are less than target in order to get
	// more precise information about impacts. This method may not be called on targets that are less than
	// the current DocIdSetIterator.docID(). After this method has been called, DocIdSetIterator.nextDoc()
	// may not be called if the current doc ID is less than target - 1 and DocIdSetIterator.advance(int)
	// may not be called on targets that are less than target.
	AdvanceShallow(target int) error

	// GetImpacts Get information about upcoming impacts for doc ids that are greater than or equal to the
	// maximum of DocIdSetIterator.docID() and the last target that was passed to advanceShallow(int).
	// This method may not be called on an unpositioned iterator on which advanceShallow(int) has never been
	// called. NOTE: advancing this iterator may invalidate the returned impacts, so they should not be used
	// after the iterator has been advanced.
	GetImpacts() (Impacts, error)
}

type FieldInfos interface {
	FieldInfo(fieldName string) *document.FieldInfo
	FieldInfoByNumber(fieldNumber int) *document.FieldInfo
	Size() int
	List() []*document.FieldInfo
	HasNorms() bool
	HasDocValues() bool
	HasVectors() bool
	HasPointValues() bool
}

// SortedSetDocValues A multi-valued version of SortedDocValues.
// Per-Document values in a SortedSetDocValues are deduplicated, dereferenced, and sorted into a
// dictionary of unique values. A pointer to the dictionary item (ordinal) can be retrieved for
// each document. Ordinals are dense and in increasing sorted order.
type SortedSetDocValues interface {
	types.DocValuesIterator

	// NextOrd Returns the next ordinal for the current document.
	// It is illegal to call this method after advanceExact(int) returned false.
	// 返回当前文档的下一个序数。在AdvanceExact(int)返回false之后调用此方法是非法的。
	NextOrd() (int64, error)

	// LookupOrd Retrieves the value for the specified ordinal.
	// The returned BytesRef may be re-used across calls to lookupOrd
	// so make sure to copy it if you want to keep it around.
	LookupOrd(ord int64) ([]byte, error)

	GetValueCount() int64
}

// TermState Encapsulates all required internal state to position the associated TermsEnum without re-seeking.
// See Also: TermsEnum.seekExact(org.apache.lucene.util.BytesRef, TermState), TermsEnum.termState()
type TermState interface {

	// CopyFrom
	// Copies the content of the given TermState to this instance
	// other – the TermState to copy
	CopyFrom(other TermState)
}

type Sort interface {
	SetSort(fields []SortField)
	GetSort() []SortField
}

// SortField
// Stores information about how to sort documents by terms in an individual field.
// Fields must be indexed in order to sort by them.
// Created: Feb 11, 2004 1:25:29 PM
// Since: lucene 1.4
// See Also: Sort
type SortField interface {
	// GetMissingValue
	// Return the item to use for documents that don't have a item.
	// A item of null indicates that default should be used.
	GetMissingValue() any

	// SetMissingValue
	// Set the item to use for documents that don't have a item.
	SetMissingValue(missingValue any) error

	// GetField
	// Returns the name of the field. Could return null if the sort is by SCORE or DOC.
	// Returns: Name of field, possibly null.
	GetField() string

	// GetType
	// Returns the type of contents in the field.
	// Returns: One of the constants SCORE, DOC, STRING, INT or FLOAT.
	GetType() SortFieldType

	// GetReverse
	// Returns whether the sort should be reversed.
	// Returns: True if natural order should be reversed.
	GetReverse() bool

	GetComparatorSource() FieldComparatorSource

	// SetCanUsePoints
	// For numeric sort fields, setting this field, indicates that the same numeric data
	// has been indexed with two fields: doc values and points and that these fields have the same name.
	// This allows to use sort optimization and skip non-competitive documents.
	SetCanUsePoints()

	GetCanUsePoints() bool

	SetBytesComparator(fn BytesComparator)

	GetBytesComparator() BytesComparator

	// GetComparator
	// Returns the FieldComparator to use for sorting.
	// - numHits: number of top hits the queue will store
	// - sortPos: position of this SortField within Sort. The comparator is primary if sortPos==0, secondary
	//		if sortPos==1, etc. Some comparators can optimize themselves when they are the primary sort.
	// Returns: FieldComparator to use when sorting
	// lucene.experimental
	GetComparator(numHits, sortPos int) FieldComparator

	//rewrite(searcher search.IndexSearcher)

	GetIndexSorter() IndexSorter

	Serialize(ctx context.Context, out store.DataOutput) error
	Equals(other SortField) bool
	String() string
}

type BytesComparator func(a, b []byte) int

// FieldComparator
// Expert: a FieldComparator compares hits so as to determine their sort order when collecting the top results
// with TopFieldCollector. The concrete public FieldComparator classes here correspond to the SortField types.
//
// The document IDs passed to these methods must only move forwards, since they are using doc values iterators
// to retrieve sort values.
//
// This API is designed to achieve high performance sorting, by exposing a tight interaction with
// FieldValueHitQueue as it visits hits. Whenever a hit is competitive, it's enrolled into a virtual slot,
// which is an int ranging from 0 to numHits-1. Segment transitions are handled by creating a dedicated
// per-segment LeafFieldComparator which also needs to interact with the FieldValueHitQueue but can optimize
// based on the segment to collect.
//
// # The following functions need to be implemented
//
// compare Compare a hit at 'slot a' with hit 'slot b'.
//
// setTopValue This method is called by TopFieldCollector to notify the FieldComparator of the top most item,
// which is used by future calls to LeafFieldComparator.compareTop.
// getLeafComparator(LeafReaderContextImpl) Invoked when the search is switching to the next segment. You may
// need to update internal state of the comparator, for example retrieving new values from DocValues.
//
// item Return the sort item stored in the specified slot. This is only called at the end of the search, in order
// to populate FieldDoc.fields when returning the top results.
//
// See Also: LeafFieldComparator
// lucene.experimental
type FieldComparator interface {
	// Compare
	// hit at slot1 with hit at slot2.
	// slot1: first slot to compare
	// slot2: second slot to compare
	// Returns: any N < 0 if slot2's item is sorted after slot1, any N > 0 if the slot2's item is sorted
	// before slot1 and 0 if they are equal
	Compare(slot1, slot2 int) int

	// SetTopValue
	// Record the top item, for future calls to LeafFieldComparator.compareTop. This is only
	// called for searches that use searchAfter (deep paging), and is called before any calls to
	// getLeafComparator(LeafReaderContextImpl).
	SetTopValue(value any)

	// Value
	// Return the actual item in the slot.
	// slot: the item
	// Returns: item in this slot
	Value(slot int) any

	// GetLeafComparator
	// Get a per-segment LeafFieldComparator to collect the given LeafReaderContextImpl.
	// All docIDs supplied to this LeafFieldComparator are relative to the current reader (you must
	// add docBase if you need to map it to a top-level docID).
	// context: current reader context
	// Returns: the comparator to use for this segment
	// Throws: IOException – if there is a low-level IO error
	GetLeafComparator(context LeafReaderContext) (LeafFieldComparator, error)

	// CompareValues
	// Returns a negative integer if first is less than second,
	// 0 if they are equal and a positive integer otherwise.
	// Default impl to assume the type implements Comparable and invoke .compareTo;
	// be sure to override this method if your FieldComparator's type isn't a
	// Comparable or if your values may sometimes be null
	CompareValues(first, second any) int

	// SetSingleSort
	// Informs the comparator that sort is done on this single field.
	// This is useful to enable some optimizations for skipping non-competitive documents.
	SetSingleSort()

	// DisableSkipping
	// Informs the comparator that the skipping of documents should be disabled.
	// This function is called in cases when the skipping functionality should not be applied
	// or not necessary. One example for numeric comparators is when we don't know if the same
	// numeric data has been indexed with docValues and points if these two fields have the
	// same name. As the skipping functionality relies on these fields to have the same data
	// and as we don't know if it is true, we have to disable it. Another example could be
	// when search sort is a part of the index sort, and can be already efficiently handled by
	// TopFieldCollector, and doing extra work for skipping in the comparator is redundant.
	DisableSkipping()
}

// FieldComparatorSource
// Provides a FieldComparator for custom field sorting.
// lucene.experimental
type FieldComparatorSource interface {
	NewComparator(fieldName string, numHits, sortPos int, reversed bool) FieldComparator
}

// SortFieldType Specifies the type of the terms to be sorted, or special types such as CUSTOM
type SortFieldType int

func (s SortFieldType) String() string {
	switch s {
	case SCORE:
		return "SCORE"
	case DOC:
		return "DOC"
	case STRING:
		return "STRING"
	case INT:
		return "INT"
	case FLOAT:
		return "FLOAT"
	case LONG:
		return "LONG"
	case DOUBLE:
		return "DOUBLE"
	case CUSTOM:
		return "CUSTOM"
	case STRING_VAL:
		return "STRING_VAL"
	case REWRITEABLE:
		return "REWRITEABLE"
	default:
		return ""
	}
}

const (
	// SCORE // Sort by document score (relevance).
	// Sort values are Float and higher values are at the front.
	SCORE = SortFieldType(iota)

	// DOC Sort by document number (index order).
	// Sort values are Integer and lower values are at the front.
	DOC

	// STRING Sort using term values as Strings.
	// Sort values are String and lower values are at the front.
	STRING

	// INT Sort using term values as encoded Integers.
	// Sort values are Integer and lower values are at the front.
	INT

	// FLOAT Sort using term values as encoded Floats.
	// Sort values are Float and lower values are at the front.
	FLOAT

	// LONG Sort using term values as encoded Longs.
	// Sort values are Long and lower values are at the front.
	LONG

	// DOUBLE Sort using term values as encoded Doubles.
	// Sort values are Double and lower values are at the front.
	DOUBLE

	// CUSTOM Sort using a custom cmp.
	// Sort values are any Comparable and sorting is done according to natural order.
	CUSTOM

	// STRING_VAL Sort using term values as Strings,
	// but comparing by item (using String.compareTo) for all comparisons.
	// This is typically slower than STRING, which uses ordinals to do the sorting.
	STRING_VAL

	// REWRITEABLE Force rewriting of SortField using rewrite(IndexSearcher) before it can be used for sorting
	REWRITEABLE
)

// CacheHelper
// A utility class that gives hooks in order to help build a cache based on the data that is contained in this index.
// lucene.experimental
type CacheHelper interface {
	// GetKey
	// Get a key that the resource can be cached on. The given entry can be compared using identity,
	// ie. Object.equals is implemented as == and Object.hashCode is implemented as System.identityHashCode.
	GetKey() string
}

// SeekStatus Represents returned result from seekCeil.
type SeekStatus int

const (
	// SEEK_STATUS_END The term was not found, and the end of iteration was hit.
	SEEK_STATUS_END = iota

	// SEEK_STATUS_FOUND The precise term was found.
	SEEK_STATUS_FOUND

	// SEEK_STATUS_NOT_FOUND A different term was found after the requested term
	SEEK_STATUS_NOT_FOUND
)

// LeafFieldComparator Expert: comparator that gets instantiated on each leaf from a top-level FieldComparator instance.
// A leaf comparator must define these functions:
// setBottom This method is called by FieldValueHitQueue to notify the FieldComparator of the current weakest ("bottom") slot. Note that this slot may not hold the weakest item according to your comparator, in cases where your comparator is not the primary one (ie, is only used to break ties from the comparators before it).
// compareBottom Compare a new hit (docID) against the "weakest" (bottom) entry in the queue.
// compareTop Compare a new hit (docID) against the top item previously set by a call to FieldComparator.setTopValue.
// copy Installs a new hit into the priority queue. The FieldValueHitQueue calls this method when a new hit is competitive.
// See Also: FieldComparator
// lucene.experimental
type LeafFieldComparator interface {
	// SetBottom Set the bottom slot, ie the "weakest" (sorted last) entry in the queue. When compareBottom is called, you should compare against this slot. This will always be called before compareBottom.
	// Params: slot – the currently weakest (sorted last) slot in the queue
	SetBottom(slot int) error

	// CompareBottom compare the bottom of the queue with this doc. This will only invoked after setBottom has
	// been called. This should return the same result as FieldComparator.compare(int, int)} as if bottom were
	// slot1 and the new document were slot 2.
	// For a search that hits many results, this method will be the hotspot (invoked by far the most frequently).
	// Params: doc – that was hit
	// Returns: any N < 0 if the doc's item is sorted after the bottom entry (not competitive), any N > 0 if
	// the doc's item is sorted before the bottom entry and 0 if they are equal.
	CompareBottom(doc int) (int, error)

	// CompareTop compare the top item with this doc. This will only invoked after setTopValue has been called.
	// This should return the same result as FieldComparator.compare(int, int)} as if topValue were slot1 and
	// the new document were slot 2. This is only called for searches that use searchAfter (deep paging).
	// Params: doc – that was hit
	// Returns: any N < 0 if the doc's item is sorted after the top entry (not competitive), any N > 0 if the
	// doc's item is sorted before the top entry and 0 if they are equal.
	CompareTop(doc int) (int, error)

	// Copy This method is called when a new hit is competitive.
	// You should copy any state associated with this document that will be required for future comparisons,
	// into the specified slot.
	// Params:  slot – which slot to copy the hit to
	//			doc – docID relative to current reader
	Copy(slot, doc int) error

	// SetScorer Sets the Scorer to use in case a document's score is needed.
	// Params: scorer – Scorer instance that you should use to obtain the current hit's score, if necessary.
	SetScorer(scorer Scorable) error

	// CompetitiveIterator Returns a competitive iterator
	// Returns: an iterator over competitive docs that are stronger than already collected docs or
	// null if such an iterator is not available for the current comparator or segment.
	CompetitiveIterator() (types.DocIdSetIterator, error)

	// SetHitsThresholdReached Informs this leaf comparator that hits threshold is reached.
	// This method is called from a collector when hits threshold is reached.
	SetHitsThresholdReached() error
}

// Scorable Allows access to the score of a Query
//type Scorable interface {
//	// Score Returns the score of the current document matching the query.
//	Score() (float64, error)
//
//	// SmoothingScore Returns the smoothing score of the current document matching the query. This score is used when the query/term does not appear in the document, and behaves like an idf. The smoothing score is particularly important when the Scorer returns a product of probabilities so that the document score does not go to zero when one probability is zero. This can return 0 or a smoothing score.
//	// Smoothing scores are described in many papers, including: Metzler, D. and Croft, W. B. , "Combining the Language Model and Inference Network Approaches to Retrieval," Information Processing and Management Special Issue on Bayesian Networks and Information Retrieval, 40(5), pp.735-750.
//	SmoothingScore(docId int) (float64, error)
//}

// IndexSorter
// Handles how documents should be sorted in an index, both within a segment and
// between segments. Implementers must provide the following methods:
// getDocComparator(LeafReader, int) - an object that determines how documents within a segment
// are to be sorted getComparableProviders(List) - an array of objects that return a sortable
// long item per document and segment getProviderName() - the SPI-registered name of a
// SortFieldProvider to serialize the sort The companion SortFieldProvider should be
// registered with SPI via META-INF/services
type IndexSorter interface {

	// GetComparableProviders
	// Get an array of IndexSorter.ComparableProvider, one per segment,
	// for merge sorting documents in different segments
	// Params: readers – the readers to be merged
	GetComparableProviders(readers []LeafReader) ([]ComparableProvider, error)

	// GetDocComparator
	// Get a comparator that determines the sort order of docs within a single IndexReader.
	// NB We cannot simply use the FieldComparator API because it requires docIDs to be sent in-order.
	// The default implementations allocate array[maxDoc] to hold native values for comparison, but 1)
	// they are transient (only alive while sorting this one segment) and 2) in the typical index
	// sorting case, they are only used to sort newly flushed segments, which will be smaller than
	// merged segments
	//
	// reader: the IndexReader to sort
	// maxDoc: the number of documents in the Reader
	GetDocComparator(reader LeafReader, maxDoc int) (DocComparator, error)

	// GetProviderName
	// The SPI-registered name of a SortFieldProvider that will deserialize the parent SortField
	GetProviderName() string
}

// ComparableProvider
// Used for sorting documents across segments
// 用于跨多个段（segment）进行文档排序
type ComparableProvider interface {
	// GetAsComparableLong Returns a long so that the natural ordering of long values
	// matches the ordering of doc IDs for the given comparator
	GetAsComparableLong(docID int) (int64, error)
}

// DocComparator
// A comparator of doc IDs, used for sorting documents within a segment
// 用于段内文档的排序
type DocComparator interface {
	// Compare
	// Compare docID1 against docID2. The contract for the return item is the same as Compare(any, any).
	Compare(docID1, docID2 int) int
}

// FieldsProducer Sole constructor. (For invocation by subclass constructors, typically implicit.)
type FieldsProducer interface {
	io.Closer

	Fields

	// CheckIntegrity
	// Checks consistency of this reader.
	// Note that this may be costly in terms of I/O, e.g. may involve computing a checksum item against large
	// data files.
	CheckIntegrity() error

	// GetMergeInstance Returns an instance optimized for merging. This instance may only be consumed in the
	// thread that called getMergeInstance().
	// The default implementation returns this
	GetMergeInstance() FieldsProducer
}
