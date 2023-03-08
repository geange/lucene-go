package index

import (
	"github.com/geange/lucene-go/core/store"
	"strings"
)

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
type DirectoryReader struct {
	directory store.Directory
}

// IndexExists Returns true if an index likely exists at the specified directory.
// Note that if a corrupt index exists, or if an index in the process of committing
// Params: directory – the directory to check for an index
// Returns: true if an index exists; false otherwise
func IndexExists(dir store.Directory) (bool, error) {
	// LUCENE-2812, LUCENE-2727, LUCENE-4738: this logic will
	// return true in cases that should arguably be false,
	// such as only IW.prepareCommit has been called, or a
	// corrupt first commit, but it's too deadly to make
	// this logic "smarter" and risk accidentally returning
	// false due to various cases like file description
	// exhaustion, access denied, etc., because in that
	// case IndexWriter may delete the entire index.  It's
	// safer to err towards "index exists" than try to be
	// smart about detecting not-yet-fully-committed or
	// corrupt indices.  This means that IndexWriter will
	// throw an exception on such indices and the app must
	// resolve the situation manually:
	files, err := dir.ListAll()
	if err != nil {
		return false, err
	}

	prefix := SEGMENTS + "_"
	for _, file := range files {
		if strings.HasPrefix(file, prefix) {
			return true, nil
		}
	}
	return false, nil
}
