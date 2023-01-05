package index

import "github.com/geange/lucene-go/core/store"

// CompoundDirectory A read-only Directory that consists of a view over a compound file.
// See Also: CompoundFormat
// lucene.experimental
type CompoundDirectory interface {
	store.Directory

	// CheckIntegrity Checks consistency of this directory.
	// Note that this may be costly in terms of I/O, e.g. may involve computing a checksum
	// value against large data files.
	CheckIntegrity() error
}
