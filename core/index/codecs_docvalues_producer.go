package index

import (
	"io"

	"github.com/geange/lucene-go/core/document"
)

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
	// Note that this may be costly in terms of I/O, e.g. may involve computing a checksum value
	// against large data files.
	// lucene.internal
	CheckIntegrity() error

	// Returns an instance optimized for merging. This instance may only be consumed in the thread
	// that called getMergeInstance().
	// The default implementation returns this
}
