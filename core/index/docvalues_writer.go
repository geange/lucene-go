package index

import "github.com/geange/lucene-go/core/types"

type DocValuesWriter interface {
	Flush(state *SegmentWriteState, sortMap DocMap, consumer DocValuesConsumer) error
	GetDocValues() types.DocIdSetIterator
}
