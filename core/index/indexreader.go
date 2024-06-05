package index

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"sync/atomic"

	"github.com/geange/lucene-go/core/document"
)

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
	Document(docID int) (*document.Document, error)

	// DocumentWithVisitor
	// Expert: visits the fields of a stored document, for custom processing/loading of each field.
	// If you simply want to load all fields, use document(int). If you want to load a subset,
	// use DocumentStoredFieldVisitor.
	DocumentWithVisitor(docID int, visitor document.StoredFieldVisitor) error

	// DocumentWithFields
	// Like Document(docID int) but only loads the specified fields. Note that this is simply sugar for
	// DocumentStoredFieldVisitor.DocumentStoredFieldVisitor(Set).
	DocumentWithFields(docID int, fieldsToLoad map[string]struct{}) (*document.Document, error)

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
	// Returns the total number of occurrences of term across all documents (the sum of the freq() for each doc that has this term). Note that, like other term measures, this measure does not take deleted documents into account.
	TotalTermFreq(ctx context.Context, term *Term) (int64, error)

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
	GetMetaData() *LeafMetaData
}

type IndexReaderSPI interface {
	GetTermVectors(docID int) (Fields, error)
	NumDocs() int
	MaxDoc() int
	DocumentWithVisitor(docID int, visitor document.StoredFieldVisitor) error
	GetContext() (IndexReaderContext, error)
	DoClose() error
}

type baseIndexReader struct {
	spi           IndexReaderSPI
	closedByChild *atomic.Bool
	refCount      *atomic.Int64
	parentReaders map[IndexReader]struct{}
	closed        *atomic.Bool
}

func newBaseIndexReader(spi IndexReaderSPI) *baseIndexReader {
	return &baseIndexReader{
		spi:           spi,
		refCount:      new(atomic.Int64),
		parentReaders: make(map[IndexReader]struct{}),
	}
}

func (r *baseIndexReader) Close() error {
	r.closed.Store(true)
	return r.spi.DoClose()
}

func (r *baseIndexReader) DocumentWithFields(docID int, fieldsToLoad map[string]struct{}) (*document.Document, error) {
	visitor := document.NewDocumentStoredFieldVisitorV1(fieldsToLoad)
	if err := r.spi.DocumentWithVisitor(docID, visitor); err != nil {
		return nil, err
	}
	return visitor.GetDocument(), nil
}

// RegisterParentReader
// Expert: This method is called by IndexReaders which wrap other readers
// (e.g. CompositeReader or FilterLeafReader) to register the parent at the child (this reader) on
// construction of the parent. When this reader is closed, it will mark all registered parents as closed,
// too. The references to parent readers are weak only, so they can be GCed once they are no longer in use.
func (r *baseIndexReader) RegisterParentReader(reader IndexReader) {
	r.parentReaders[reader] = struct{}{}
}

// NotifyReaderClosedListeners overridden by StandardDirectoryReader and SegmentReader
func (r *baseIndexReader) NotifyReaderClosedListeners() error {
	return nil
}

func (r *baseIndexReader) reportCloseToParentReaders() error {
	//for parent, _ := range r.parentReaders {
	//	if p, ok := parent.(*ReaderBase); ok {
	//		p.closedByChild = true
	//		//p.refCount.Add(0)
	//		err := p.reportCloseToParentReaders()
	//		if err != nil {
	//			return err
	//		}
	//	}
	//}
	//return nil
	panic("")
}

func (r *baseIndexReader) GetRefCount() int {
	// NOTE: don't ensureOpen, so that callers can see
	// refCount is 0 (reader is closed)
	return int(r.refCount.Load())
}

func (r *baseIndexReader) IncRef() error {
	if !r.TryIncRef() {
		err := r.ensureOpen()
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *baseIndexReader) DecRef() error {
	// only check refcount here (don't call ensureOpen()), so we can
	// still close the reader if it was made invalid by a child:
	if r.refCount.Load() <= 0 {
		return errors.New("this Reader is closed")
	}

	rc := r.refCount.Add(-1)
	if rc == 0 {
		r.closed.Store(true)
		return r.spi.DoClose()
	}

	if rc < 0 {
		return fmt.Errorf("too many decRef calls: refCount is %d after decrement", rc)
	}
	return nil
}

func (r *baseIndexReader) ensureOpen() error {
	if r.refCount.Load() <= 0 {
		return errors.New("this Reader is closed")
	}

	// the happens before rule on reading the refCount, which must be after the fake write,
	// ensures that we see the item:
	if r.closedByChild.Load() {
		return errors.New("this Reader cannot be used anymore as one of its child readers was closed")
	}
	return nil
}

func (r *baseIndexReader) TryIncRef() bool {
	count := int64(0)
	for {
		count = r.refCount.Load()
		if count <= 0 {
			return false
		}
		return r.refCount.Swap(count+1) == count
	}
}

func (r *baseIndexReader) GetTermVector(docID int, field string) (Terms, error) {
	vectors, err := r.spi.GetTermVectors(docID)
	if err != nil {
		return nil, err
	}
	return vectors.Terms(field)
}

func (r *baseIndexReader) NumDeletedDocs() int {
	return r.spi.MaxDoc() - r.spi.NumDocs()
}

func (r *baseIndexReader) Document(docID int) (*document.Document, error) {
	visitor := document.NewDocumentStoredFieldVisitor()
	if err := r.spi.DocumentWithVisitor(docID, visitor); err != nil {
		return nil, err
	}
	return visitor.GetDocument(), nil
}

func (r *baseIndexReader) HasDeletions() bool {
	return r.NumDeletedDocs() > 0
}

func (r *baseIndexReader) Leaves() ([]LeafReaderContext, error) {
	ctx, err := r.spi.GetContext()
	if err != nil {
		return nil, err
	}
	return ctx.Leaves()
}

// CacheHelper
// A utility class that gives hooks in order to help build a cache based on the data that is contained in this index.
// lucene.experimental
type CacheHelper interface {
	// GetKey
	// Get a key that the resource can be cached on. The given entry can be compared using identity,
	// ie. Object.equals is implemented as == and Object.hashCode is implemented as System.identityHashCode.
	GetKey()
}

var _ sort.Interface = &ReaderSorter{}

type ReaderSorter struct {
	Readers   []IndexReader
	FnCompare func(a, b LeafReader) int
}

func (r *ReaderSorter) Len() int {
	return len(r.Readers)
}

func (r *ReaderSorter) Less(i, j int) bool {
	return r.FnCompare(r.Readers[i].(LeafReader), r.Readers[j].(LeafReader)) < 0
}

func (r *ReaderSorter) Swap(i, j int) {
	r.Readers[i], r.Readers[j] = r.Readers[j], r.Readers[i]
}

type ClosedListener interface {
	// Invoked when the resource (segment core, or index reader) that is being cached on is closed.
}
