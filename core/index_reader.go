package core

import "io"

type IndexReader interface {
	io.Closer

	// GetTermVectors Retrieve term vectors for this document, or null if term vectors were not indexed.
	// The returned Fields instance acts like a single-document inverted index (the docID will be 0).
	GetTermVectors(docID int) (Fields, error)

	// Returns the number of documents in this index.
	// NOTE: This operation may run in O(maxDoc). Implementations that can't return this number in constant-time should cache it.
	numDocs() int

	// Returns one greater than the largest possible document number. This may be used to, e.g., determine how big to allocate an array which will have an element for every document number in an index.
	maxDoc() int

	// DocumentV1 Expert: visits the fields of a stored document, for custom processing/loading of each field. If you simply want to load all fields, use document(int). If you want to load a subset, use DocumentStoredFieldVisitor.
	DocumentV1(docID int, visitor StoredFieldVisitor) error

	// DoClose Implements close.
	DoClose() error

	// GetContext Expert: Returns the root IndexReaderContext for this IndexReader's sub-reader tree.
	//Iff this reader is composed of sub readers, i.e. this reader being a composite reader, this method returns a CompositeReaderContext holding the reader's direct children as well as a view of the reader tree's atomic leaf contexts. All sub- IndexReaderContext instances referenced from this readers top-level context are private to this reader and are not shared with another context tree. For example, IndexSearcher uses this API to drive searching by one atomic leaf reader at a time. If this reader is not composed of child readers, this method returns an LeafReaderContext.
	//Note: Any of the sub-CompositeReaderContext instances referenced from this top-level context do not support CompositeReaderContext.leaves(). Only the top-level context maintains the convenience leaf-view for performance reasons.
	GetContext() IndexReaderContext

	// Optional method: Return a IndexReader.CacheHelper that can be used to cache based on the content of this reader. Two readers that have different data or different sets of deleted documents will be considered different.
	//A return value of null indicates that this reader is not suited for caching, which is typically the case for short-lived wrappers that alter the content of the wrapped reader.
	getReaderCacheHelper() CacheHelper

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
}

type CacheHelper interface {
	// TODO
}
