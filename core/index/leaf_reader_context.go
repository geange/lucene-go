package index

// LeafReaderContext IndexReaderContext for LeafReader instances.
type LeafReaderContext struct {
	IndexReaderContextImp

	// The reader's ord in the top-level's leaves array
	Ord int

	// The reader's absolute doc base
	DocBase int

	reader LeafReader
	leaves []LeafReaderContext
}

func (l *LeafReaderContext) LeafReader() LeafReader {
	return l.reader
}

func (l *LeafReaderContext) Reader() IndexReader {
	return l.reader
}

func (l *LeafReaderContext) Leaves() ([]LeafReaderContext, error) {
	//TODO implement me
	panic("implement me")
}

func (l *LeafReaderContext) Children() []IndexReaderContext {
	//TODO implement me
	panic("implement me")
}

func (l *LeafReaderContext) Identity() string {
	//TODO implement me
	panic("implement me")
}
