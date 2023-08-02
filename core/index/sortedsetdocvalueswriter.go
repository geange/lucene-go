package index

import (
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/types"
)

var _ DocValuesWriter = &SortedSetDocValuesWriter{}

type SortedSetDocValuesWriter struct {
}

func (s *SortedSetDocValuesWriter) Flush(state *index.SegmentWriteState, sortMap DocMap, consumer index.DocValuesConsumer) error {
	//TODO implement me
	panic("implement me")
}

func (s *SortedSetDocValuesWriter) GetDocValues() types.DocIdSetIterator {
	//TODO implement me
	panic("implement me")
}
