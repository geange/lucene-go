package index

import (
	"errors"
)

// CompositeReader
// Instances of this reader type can only be used to get stored fields from the underlying LeafReaders,
// but it is not possible to directly retrieve postings. To do that, get the LeafReaderContext for all
// sub-readers via leaves().
//
// IndexReader instances for indexes on disk are usually constructed with a call to one of the static
// DirectoryReader.open() methods, e.g. DirectoryReader.open(Directory). DirectoryReader implements
// the CompositeReader interface, it is not possible to directly get postings.
// Concrete subclasses of IndexReader are usually constructed with a call to one of the static open()
// methods, e.g. DirectoryReader.open(Directory).
//
// For efficiency, in this API documents are often referred to via document numbers, non-negative integers
// which each name a unique document in the index. These document numbers are ephemeral -- they may change
// as documents are added to and deleted from an index. Clients should thus not rely on a given document
// having the same number between sessions.
//
// NOTE: IndexReader instances are completely thread safe, meaning multiple threads can call any of its
// methods, concurrently. If your application requires external synchronization, you should not
// synchronize on the IndexReader instance; use your own (non-Lucene) objects instead.
type CompositeReader interface {
	IndexReader

	// GetSequentialSubReaders
	// Expert: returns the sequential sub readers that this reader is logically composed of.
	// This method may not return null.
	// NOTE: In contrast to previous Lucene versions this method is no longer public, code that
	// wants to get all LeafReaders this composite is composed of should use IndexReader.leaves().
	// See Also: IndexReader.leaves()
	GetSequentialSubReaders() []IndexReader
}

var _ IndexReaderContext = &CompositeReaderContext{}

// CompositeReaderContext IndexReaderContext for CompositeReader instance.
type CompositeReaderContext struct {
	*IndexReaderContextDefault

	children []IndexReaderContext
	leaves   []*LeafReaderContext
	reader   CompositeReader
}

func (c *CompositeReaderContext) Reader() IndexReader {
	return c.reader
}

func (c *CompositeReaderContext) Leaves() ([]*LeafReaderContext, error) {
	if c.IsTopLevel {
		return nil, errors.New("this is not a top-level context")
	}
	return c.leaves, nil
}

func (c *CompositeReaderContext) Children() []IndexReaderContext {
	return c.children
}
