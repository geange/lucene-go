package index

var _ DocValuesWriter = &SortedNumericDocValuesWriter{}

type SortedNumericDocValuesWriter struct {
}

func (s *SortedNumericDocValuesWriter) Flush(state *SegmentWriteState, sortMap DocMap, consumer DocValuesConsumer) error {
	//TODO implement me
	panic("implement me")
}

func (s *SortedNumericDocValuesWriter) GetDocValues() DocIdSetIterator {
	//TODO implement me
	panic("implement me")
}
