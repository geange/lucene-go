package index

import (
	"context"
	"errors"
	"github.com/geange/lucene-go/core/interface/index"
)

type BaseLeafReader struct {
	*baseIndexReader

	LeafReaderBaseInner

	readerContext index.LeafReaderContext
}

type LeafReaderBaseInner interface {
	Terms(field string) (index.Terms, error)
}

func NewBaseLeafReader(reader index.LeafReader) *BaseLeafReader {
	return &BaseLeafReader{
		LeafReaderBaseInner: reader,
		readerContext:       NewLeafReaderContext(reader),
		baseIndexReader:     newBaseIndexReader(reader),
	}
}

func (r *BaseLeafReader) Postings(ctx context.Context, term index.Term, flags int) (index.PostingsEnum, error) {
	terms, err := r.Terms(term.Field())
	if err != nil {
		return nil, err
	}
	if terms == nil {
		return nil, nil
	}
	termsEnum, err := terms.Iterator()
	if err != nil {
		return nil, err
	}

	if ok, err := termsEnum.SeekExact(ctx, term.Bytes()); err == nil && ok {
		return termsEnum.Postings(nil, flags)
	}

	return nil, nil
}

func (r *BaseLeafReader) GetContext() (index.IndexReaderContext, error) {
	return r.readerContext, nil
}

func (r *BaseLeafReader) DocFreq(ctx context.Context, term index.Term) (int, error) {
	terms, err := r.Terms(term.Field())
	if err != nil {
		return 0, err
	}
	if terms == nil {
		return 0, nil
	}

	termsEnum, err := terms.Iterator()
	if err != nil {
		return 0, err
	}
	if ok, err := termsEnum.SeekExact(ctx, term.Bytes()); err == nil && ok {
		return termsEnum.DocFreq()
	} else {
		return 0, err
	}
}

func (r *BaseLeafReader) TotalTermFreq(ctx context.Context, term index.Term) (int64, error) {
	terms, err := r.Terms(term.Field())
	if err != nil {
		return 0, err
	}
	if terms == nil {
		return 0, nil
	}

	termsEnum, err := terms.Iterator()
	if err != nil {
		return 0, err
	}
	if ok, err := termsEnum.SeekExact(ctx, term.Bytes()); err == nil && ok {
		return termsEnum.TotalTermFreq()
	} else {
		return 0, err
	}
}

func (r *BaseLeafReader) GetSumDocFreq(field string) (int64, error) {
	terms, err := r.Terms(field)
	if err != nil {
		return 0, err
	}
	if terms == nil {
		return 0, nil
	}

	return terms.GetSumDocFreq()
}

func (r *BaseLeafReader) GetDocCount(field string) (int, error) {
	terms, err := r.Terms(field)
	if err != nil {
		return 0, err
	}
	if terms == nil {
		return 0, nil
	}

	return terms.GetDocCount()
}

func (r *BaseLeafReader) GetSumTotalTermFreq(field string) (int64, error) {
	terms, err := r.Terms(field)
	if err != nil {
		return 0, err
	}
	if terms == nil {
		return 0, nil
	}

	return terms.GetSumTotalTermFreq()
}

// LeafReaderContextImpl IndexReaderContext for LeafReader instances.
type LeafReaderContextImpl struct {
	*BaseIndexReaderContext

	// The reader's ord in the top-level's leaves array
	ord int

	// The reader's absolute doc base
	docBase int

	reader index.LeafReader
	leaves []index.LeafReaderContext
}

func NewLeafReaderContext(leafReader index.LeafReader) index.LeafReaderContext {
	return NewLeafReaderContextV1(nil, leafReader, 0, 0, 0, 0)
}

func NewLeafReaderContextV1(parent *CompositeReaderContext, reader index.LeafReader,
	ord, docBase, leafOrd, leafDocBase int) index.LeafReaderContext {

	ctx := &LeafReaderContextImpl{
		BaseIndexReaderContext: NewBaseIndexReaderContext(parent, ord, docBase),
		ord:                    leafOrd,
		docBase:                leafDocBase,
		reader:                 reader,
		leaves:                 nil,
	}

	if ctx.isTopLevel {
		ctx.leaves = []index.LeafReaderContext{ctx}
	}

	return ctx
}

func (l *LeafReaderContextImpl) Ord() int {
	return l.ord
}

func (l *LeafReaderContextImpl) DocBase() int {
	return l.docBase
}

func (l *LeafReaderContextImpl) LeafReader() index.LeafReader {
	return l.reader
}

func (l *LeafReaderContextImpl) Reader() index.IndexReader {
	return l.reader
}

func (l *LeafReaderContextImpl) Leaves() ([]index.LeafReaderContext, error) {
	if !l.isTopLevel {
		return nil, errors.New("this is not a top-level context")
	}
	return l.leaves, nil
}

func (l *LeafReaderContextImpl) Children() []index.IndexReaderContext {
	return nil
}

func (l *LeafReaderContextImpl) Identity() string {
	return l.identity
}
