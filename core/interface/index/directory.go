package index

import (
	"context"
	"github.com/geange/lucene-go/core/store"
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
	// wants to get all LeafReaders this composite is composed of should use Reader.leaves().
	// See Also: Reader.leaves()
	GetSequentialSubReaders() []IndexReader
}

// DirectoryReader is an implementation of CompositeReader that can read indexes in a Directory.
// DirectoryReader instances are usually constructed with a call to one of the static open() methods,
// e.g. open(Directory).
// For efficiency, in this API documents are often referred to via document numbers, non-negative
// integers which each name a unique document in the index. These document numbers are ephemeral
// -- they may change as documents are added to and deleted from an index. Clients should thus not
// rely on a given document having the same number between sessions.
//
// NOTE: IndexReader instances are completely thread safe, meaning multiple threads can call any of
// its methods, concurrently. If your application requires external synchronization, you should not
// synchronize on the IndexReader instance; use your own (non-Lucene) objects instead.
type DirectoryReader interface {
	CompositeReader

	Directory() store.Directory

	// GetVersion
	// Version number when this IndexReader was opened.
	// This method returns the version recorded in the commit that the reader opened.
	// This version is advanced every time a change is made with IndexWriter.
	GetVersion() int64

	// IsCurrent
	// Check whether any new changes have occurred to the index since this reader was opened.
	// If this reader was created by calling open, then this method checks if any further commits
	// (see IndexWriter.commit) have occurred in the directory.
	// If instead this reader is a near real-time reader (ie, obtained by a call to open(IndexWriter),
	// or by calling openIfChanged on a near real-time reader), then this method checks if either a
	// new commit has occurred, or any new uncommitted changes have taken place via the writer.
	// Note that even if the writer has only performed merging, this method will still return false.
	// In any event, if this returns false, you should call openIfChanged to get a new reader that sees the changes.
	// Throws: IOException – if there is a low-level IO error
	IsCurrent(ctx context.Context) (bool, error)

	// GetIndexCommit
	// Expert: return the IndexCommit that this reader has opened.
	// lucene.experimental
	GetIndexCommit() (IndexCommit, error)
}
