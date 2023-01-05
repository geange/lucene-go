package index

import (
	"io"
)

// PointsReader Abstract API to visit point values.
// lucene.experimental
type PointsReader interface {
	io.Closer

	// CheckIntegrity Checks consistency of this reader.
	// Note that this may be costly in terms of I/O,
	// e.g. may involve computing a checksum value against large data files.
	// lucene.internal
	CheckIntegrity() error

	// GetValues Return PointValues for the given field.
	GetValues(field string) (PointValues, error)
}

// GetMergeInstance Returns an instance optimized for merging.
// This instance may only be used in the thread that acquires it.
// The default implementation returns this
//GetMergeInstance() PointsReader
