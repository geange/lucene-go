package index

import "errors"

// LeafReaderContext IndexReaderContext for LeafReader instances.
type LeafReaderContext struct {
	*IndexReaderContextImp

	// The reader's ord in the top-level's leaves array
	Ord int

	// The reader's absolute doc base
	DocBase int

	reader LeafReader
	leaves []*LeafReaderContext
}

func NewLeafReaderContext(leafReader LeafReader) *LeafReaderContext {
	return NewLeafReaderContext6(nil, leafReader, 0, 0, 0, 0)
}

func NewLeafReaderContext6(parent *CompositeReaderContext, reader LeafReader,
	ord, docBase, leafOrd, leafDocBase int) *LeafReaderContext {

	ctx := &LeafReaderContext{
		IndexReaderContextImp: NewIndexReaderContextImp(parent, ord, docBase),
		Ord:                   leafOrd,
		DocBase:               leafDocBase,
		reader:                reader,
		leaves:                nil,
	}

	if ctx.IsTopLevel {
		ctx.leaves = []*LeafReaderContext{ctx}
	}

	return ctx
}

func (l *LeafReaderContext) LeafReader() LeafReader {
	return l.reader
}

func (l *LeafReaderContext) Reader() IndexReader {
	return l.reader
}

func (l *LeafReaderContext) Leaves() ([]*LeafReaderContext, error) {
	if !l.IsTopLevel {
		return nil, errors.New("this is not a top-level context")
	}
	return l.leaves, nil
}

func (l *LeafReaderContext) Children() []IndexReaderContext {
	return nil
}

func (l *LeafReaderContext) Identity() string {
	return l.identity
}
