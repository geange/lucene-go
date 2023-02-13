package index

import "github.com/geange/lucene-go/core/store"

var _ DocValuesWriter = &BinaryDocValuesWriter{}

type BinaryDocValuesWriter struct {
	bytes    *PagedBytes
	bytesOut store.DataOutput
}

func (b *BinaryDocValuesWriter) Flush(state *SegmentWriteState, sortMap DocMap, consumer DocValuesConsumer) error {
	//TODO implement me
	panic("implement me")
}

func (b *BinaryDocValuesWriter) GetDocValues() DocIdSetIterator {
	//TODO implement me
	panic("implement me")
}
