package index

import (
	"errors"
	"fmt"
	"github.com/geange/lucene-go/core/document"
	"go.uber.org/atomic"
	"io"
	"sort"
	"sync"
)

type IndexReader interface {
	io.Closer

	// GetTermVectors Retrieve term vectors for this document, or null if term vectors were not indexed.
	// The returned Fields instance acts like a single-document inverted index (the docID will be 0).
	GetTermVectors(docID int) (Fields, error)

	// GetTermVector Retrieve term vector for this document and field, or null if term vectors were not indexed.
	// The returned Fields instance acts like a single-document inverted index (the docID will be 0).
	GetTermVector(docID int, field string) (Terms, error)

	// NumDocs Returns the number of documents in this index.
	// NOTE: This operation may run in O(maxDoc). Implementations that can't return this number in
	// constant-time should cache it.
	NumDocs() int

	// MaxDoc Returns one greater than the largest possible document number. This may be used to,
	// e.g., determine how big to allocate an array which will have an element for every document
	// number in an index.
	MaxDoc() int

	// NumDeletedDocs Returns the number of deleted documents.
	// NOTE: This operation may run in O(maxDoc).
	NumDeletedDocs() int

	// Document Returns the stored fields of the nth Document in this index. This is just sugar for using
	// DocumentStoredFieldVisitor.
	// NOTE: for performance reasons, this method does not check if the requested document is deleted, and
	// therefore asking for a deleted document may yield unspecified results. Usually this is not required,
	// however you can test if the doc is deleted by checking the Bits returned from MultiBits.getLiveDocs.
	// NOTE: only the content of a field is returned, if that field was stored during indexing. Metadata like
	// boost, omitNorm, IndexOptions, tokenized, etc., are not preserved.
	// Throws: CorruptIndexException – if the index is corrupt
	// IOException – if there is a low-level IO error
	// TODO: we need a separate StoredField, so that the
	// Document returned here contains that class not
	// IndexableField
	Document(docID int) (*document.Document, error)

	// DocumentV1 Expert: visits the fields of a stored document, for custom processing/loading of each field.
	// If you simply want to load all fields, use document(int). If you want to load a subset,
	// use DocumentStoredFieldVisitor.
	DocumentV1(docID int, visitor document.StoredFieldVisitor) error

	// DocumentV2 Like document(int) but only loads the specified fields. Note that this is simply sugar for
	// DocumentStoredFieldVisitor.DocumentStoredFieldVisitor(Set).
	DocumentV2(docID int, fieldsToLoad map[string]struct{}) (*document.Document, error)

	// HasDeletions Returns true if any documents have been deleted. Implementers should consider overriding
	// this method if maxDoc() or numDocs() are not constant-time operations.
	HasDeletions() bool

	// DoClose Implements close.
	DoClose() error

	// GetContext Expert: Returns the root IndexReaderContext for this IndexReader's sub-reader tree.
	// If this reader is composed of sub readers, i.e. this reader being a composite reader, this method
	// returns a CompositeReaderContext holding the reader's direct children as well as a view of the
	// reader tree's atomic leaf contexts. All sub- IndexReaderContext instances referenced from this
	// readers top-level context are private to this reader and are not shared with another context tree.
	// For example, IndexSearcher uses this API to drive searching by one atomic leaf reader at a time.
	// If this reader is not composed of child readers, this method returns an LeafReaderContext.
	// Note: Any of the sub-CompositeReaderContext instances referenced from this top-level context do
	// not support CompositeReaderContext.leaves(). Only the top-level context maintains the convenience
	// leaf-view for performance reasons.
	GetContext() IndexReaderContext

	// Leaves Returns the reader's leaves, or itself if this reader is atomic. This is a convenience method
	// calling this.getContext().leaves().
	// See Also: IndexReaderContext.leaves()
	Leaves() ([]*LeafReaderContext, error)

	// GetReaderCacheHelper Optional method: Return a IndexReader.CacheHelper that can be used to cache based
	// on the content of this reader. Two readers that have different data or different sets of deleted
	// documents will be considered different.
	// A return value of null indicates that this reader is not suited for caching, which is typically the
	// case for short-lived wrappers that alter the content of the wrapped reader.
	GetReaderCacheHelper() CacheHelper

	// DocFreq Returns the number of documents containing the term. This method returns 0 if the term or field
	// does not exists. This method does not take into account deleted documents that have not yet been merged away.
	// See Also: TermsEnum.docFreq()
	DocFreq(term Term) (int, error)

	// TotalTermFreq Returns the total number of occurrences of term across all documents (the sum of the freq() for each doc that has this term). Note that, like other term measures, this measure does not take deleted documents into account.
	TotalTermFreq(term *Term) (int64, error)

	// GetSumDocFreq Returns the sum of TermsEnum.docFreq() for all terms in this field. Note that, just like
	// other term measures, this measure does not take deleted documents into account.
	// See Also: Terms.getSumDocFreq()
	GetSumDocFreq(field string) (int64, error)

	// GetDocCount Returns the number of documents that have at least one term for this field. Note that,
	// just like other term measures, this measure does not take deleted documents into account.
	// See Also: Terms.getDocCount()
	GetDocCount(field string) (int, error)

	// GetSumTotalTermFreq Returns the sum of TermsEnum.totalTermFreq for all terms in this field. Note that,
	// just like other term measures, this measure does not take deleted documents into account.
	// See Also: Terms.getSumTotalTermFreq()
	GetSumTotalTermFreq(field string) (int64, error)
	//RegisterParentReader(reader IndexReader)

	GetRefCount() int
	IncRef() error
	DecRef() error
	GetMetaData() *LeafMetaData
}

type IndexReaderDefaultSPI interface {
	GetTermVectors(docID int) (Fields, error)
	NumDocs() int
	MaxDoc() int
	DocumentV1(docID int, visitor document.StoredFieldVisitor) error
	GetContext() IndexReaderContext
	DoClose() error
}

type IndexReaderDefault struct {
	IndexReaderDefaultSPI

	closed        bool
	closedByChild bool
	refCount      *atomic.Int64
	parentReaders map[IndexReader]struct{}
	sync.Mutex
}

func NewIndexReaderDefault(spi IndexReaderDefaultSPI) *IndexReaderDefault {
	return &IndexReaderDefault{
		IndexReaderDefaultSPI: spi,
		refCount:              atomic.NewInt64(0),
		parentReaders:         make(map[IndexReader]struct{}),
	}
}

func (r *IndexReaderDefault) Close() error {
	r.closed = true
	return r.DoClose()
}

func (r *IndexReaderDefault) DocumentV2(docID int, fieldsToLoad map[string]struct{}) (*document.Document, error) {
	visitor := document.NewDocumentStoredFieldVisitorV1(fieldsToLoad)
	if err := r.DocumentV1(docID, visitor); err != nil {
		return nil, err
	}
	return visitor.GetDocument(), nil
}

// RegisterParentReader Expert: This method is called by IndexReaders which wrap other readers
// (e.g. CompositeReader or FilterLeafReader) to register the parent at the child (this reader) on
// construction of the parent. When this reader is closed, it will mark all registered parents as closed,
// too. The references to parent readers are weak only, so they can be GCed once they are no longer in use.
func (r *IndexReaderDefault) RegisterParentReader(reader IndexReader) {
	r.parentReaders[reader] = struct{}{}
}

// NotifyReaderClosedListeners overridden by StandardDirectoryReader and SegmentReader
func (r *IndexReaderDefault) NotifyReaderClosedListeners() error {
	return nil
}

func (r *IndexReaderDefault) reportCloseToParentReaders() error {
	//for parent, _ := range r.parentReaders {
	//	if p, ok := parent.(*IndexReaderDefault); ok {
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

func (r *IndexReaderDefault) GetRefCount() int {
	// NOTE: don't ensureOpen, so that callers can see
	// refCount is 0 (reader is closed)
	return int(r.refCount.Load())
}

func (r *IndexReaderDefault) IncRef() error {
	if !r.TryIncRef() {
		err := r.ensureOpen()
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *IndexReaderDefault) DecRef() error {
	// only check refcount here (don't call ensureOpen()), so we can
	// still close the reader if it was made invalid by a child:
	if r.refCount.Load() <= 0 {
		return errors.New("this IndexReader is closed")
	}

	rc := r.refCount.Dec()
	if rc == 0 {
		r.closed = true
		return r.DoClose()
	}

	if rc < 0 {
		return fmt.Errorf("too many decRef calls: refCount is %d after decrement", rc)
	}
	return nil
}

func (r *IndexReaderDefault) ensureOpen() error {
	if r.refCount.Load() <= 0 {
		return errors.New("this IndexReader is closed")
	}

	// the happens before rule on reading the refCount, which must be after the fake write,
	// ensures that we see the value:
	if r.closedByChild {
		return errors.New("this IndexReader cannot be used anymore as one of its child readers was closed")
	}
	return nil
}

func (r *IndexReaderDefault) TryIncRef() bool {
	count := int64(0)
	for {
		count = r.refCount.Load()
		if count > 0 {
			if r.refCount.CAS(count, count+1) {
				return true
			}
		} else {
			break
		}
	}
	return false
}

func (r *IndexReaderDefault) GetTermVector(docID int, field string) (Terms, error) {
	vectors, err := r.GetTermVectors(docID)
	if err != nil {
		return nil, err
	}
	return vectors.Terms(field)
}

func (r *IndexReaderDefault) NumDeletedDocs() int {
	return r.MaxDoc() - r.NumDocs()
}

func (r *IndexReaderDefault) Document(docID int) (*document.Document, error) {
	visitor := document.NewDocumentStoredFieldVisitor()
	if err := r.DocumentV1(docID, visitor); err != nil {
		return nil, err
	}
	return visitor.GetDocument(), nil
}

func (r *IndexReaderDefault) HasDeletions() bool {
	return r.NumDeletedDocs() > 0
}

func (r *IndexReaderDefault) Leaves() ([]*LeafReaderContext, error) {
	return r.GetContext().Leaves()
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

var _ sort.Interface = &IndexReaderSorter{}

type IndexReaderSorter struct {
	Readers   []IndexReader
	FnCompare func(a, b LeafReader) int
}

func (r *IndexReaderSorter) Len() int {
	return len(r.Readers)
}

func (r *IndexReaderSorter) Less(i, j int) bool {
	return r.FnCompare(r.Readers[i].(LeafReader), r.Readers[j].(LeafReader)) < 0
}

func (r *IndexReaderSorter) Swap(i, j int) {
	r.Readers[i], r.Readers[j] = r.Readers[j], r.Readers[i]
}

type ClosedListener interface {
	// Invoked when the resource (segment core, or index reader) that is being cached on is closed.
}
