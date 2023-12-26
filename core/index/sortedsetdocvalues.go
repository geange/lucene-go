package index

import "github.com/geange/lucene-go/core/types"

// SortedSetDocValues A multi-valued version of SortedDocValues.
// Per-Document values in a SortedSetDocValues are deduplicated, dereferenced, and sorted into a
// dictionary of unique values. A pointer to the dictionary item (ordinal) can be retrieved for
// each document. Ordinals are dense and in increasing sorted order.
type SortedSetDocValues interface {
	types.DocValuesIterator

	// NextOrd Returns the next ordinal for the current document.
	// It is illegal to call this method after advanceExact(int) returned false.
	// 返回当前文档的下一个序数。在AdvanceExact(int)返回false之后调用此方法是非法的。
	NextOrd() (int64, error)

	// LookupOrd Retrieves the value for the specified ordinal.
	// The returned BytesRef may be re-used across calls to lookupOrd
	// so make sure to copy it if you want to keep it around.
	LookupOrd(ord int64) ([]byte, error)

	GetValueCount() int64
}

const (
	NO_MORE_ORDS = -1
)
