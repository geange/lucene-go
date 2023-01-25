package index

import (
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
)

// PointValuesWriter Buffers up pending byte[][] value(s) per doc, then flushes when segment flushes.
type PointValuesWriter struct {
	fieldInfo         *types.FieldInfo
	bytes             *PagedBytes
	bytesOut          store.DataOutput
	docIDs            []int
	numPoints         int
	numDocs           int
	lastDocID         int
	packedBytesLength int
}
