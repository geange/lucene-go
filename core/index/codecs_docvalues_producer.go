package index

import (
	"github.com/geange/lucene-go/core/types"
	"io"
)

// DocValuesProducer Abstract API that produces numeric, binary, sorted, sortedset, and sortednumeric docvalues.
// lucene.experimental
type DocValuesProducer interface {
	io.Closer

	// GetNumeric Returns NumericDocValues for this field. The returned instance need not be thread-safe:
	// it will only be used by a single thread.
	GetNumeric(field *types.FieldInfo) (NumericDocValues, error)

	// GetBinary Returns BinaryDocValues for this field. The returned instance need not be thread-safe:
	// it will only be used by a single thread.
	GetBinary(field *types.FieldInfo) (BinaryDocValues, error)

	// GetSorted Returns SortedDocValues for this field. The returned instance need not be
	// thread-safe: it will only be used by a single thread.
	GetSorted(field *types.FieldInfo) (SortedDocValues, error)

	// GetSortedNumeric Returns SortedNumericDocValues for this field. The returned instance
	// need not be thread-safe: it will only be used by a single thread.
	GetSortedNumeric(field *types.FieldInfo) (SortedNumericDocValues, error)

	// GetSortedSet Returns SortedSetDocValues for this field. The returned instance need not
	// be thread-safe: it will only be used by a single thread.
	GetSortedSet(field *types.FieldInfo) (SortedSetDocValues, error)

	// CheckIntegrity Checks consistency of this producer
	// Note that this may be costly in terms of I/O, e.g. may involve computing a checksum value
	// against large data files.
	// lucene.internal
	CheckIntegrity() error

	// Returns an instance optimized for merging. This instance may only be consumed in the thread
	// that called getMergeInstance().
	// The default implementation returns this
}
