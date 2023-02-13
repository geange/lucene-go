package index

var _ DocValuesWriter = &NumericDocValuesWriter{}

type NumericDocValuesWriter struct {
}

func (n *NumericDocValuesWriter) Flush(state *SegmentWriteState, sortMap DocMap, consumer DocValuesConsumer) error {
	//TODO implement me
	panic("implement me")
}

func (n *NumericDocValuesWriter) GetDocValues() DocIdSetIterator {
	//TODO implement me
	panic("implement me")
}
