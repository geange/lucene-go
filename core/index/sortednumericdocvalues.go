package index

import (
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/types"
)

var _ DocValuesWriter = &SortedNumericDocValuesWriter{}

type SortedNumericDocValuesWriter struct {
}

func (s *SortedNumericDocValuesWriter) Flush(state *index.SegmentWriteState, sortMap DocMap, consumer index.DocValuesConsumer) error {
	//TODO implement me
	panic("implement me")
}

func (s *SortedNumericDocValuesWriter) GetDocValues() types.DocIdSetIterator {
	//TODO implement me
	panic("implement me")
}
