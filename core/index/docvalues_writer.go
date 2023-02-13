package index

type DocValuesWriter interface {
	Flush(state *SegmentWriteState, sortMap DocMap, consumer DocValuesConsumer) error
	GetDocValues() DocIdSetIterator
}
