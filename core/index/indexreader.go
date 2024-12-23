package index

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync/atomic"

	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
)

type IndexReaderSPI interface {
	GetTermVectors(docID int) (index.Fields, error)
	NumDocs() int
	MaxDoc() int
	DocumentWithVisitor(ctx context.Context, docID int, visitor document.StoredFieldVisitor) error
	GetContext() (index.IndexReaderContext, error)
	DoClose() error
}

type baseIndexReader struct {
	spi           IndexReaderSPI
	closedByChild *atomic.Bool
	refCount      *atomic.Int64
	parentReaders map[index.IndexReader]struct{}
	closed        *atomic.Bool
}

func newBaseIndexReader(spi IndexReaderSPI) *baseIndexReader {
	return &baseIndexReader{
		spi:           spi,
		refCount:      new(atomic.Int64),
		parentReaders: make(map[index.IndexReader]struct{}),
	}
}

func (r *baseIndexReader) Close() error {
	r.closed.Store(true)
	return r.spi.DoClose()
}

func (r *baseIndexReader) DocumentWithFields(ctx context.Context, docID int, fieldsToLoad []string) (*document.Document, error) {
	visitor := document.NewDocumentStoredFieldVisitor(fieldsToLoad...)
	if err := r.spi.DocumentWithVisitor(ctx, docID, visitor); err != nil {
		return nil, err
	}
	return visitor.GetDocument(), nil
}

// RegisterParentReader
// Expert: This method is called by IndexReaders which wrap other readers
// (e.g. CompositeReader or FilterLeafReader) to register the parent at the child (this reader) on
// construction of the parent. When this reader is closed, it will mark all registered parents as closed,
// too. The references to parent readers are weak only, so they can be GCed once they are no longer in use.
func (r *baseIndexReader) RegisterParentReader(reader index.IndexReader) {
	r.parentReaders[reader] = struct{}{}
}

// NotifyReaderClosedListeners overridden by StandardDirectoryReader and SegmentReader
func (r *baseIndexReader) NotifyReaderClosedListeners() error {
	return nil
}

func (r *baseIndexReader) reportCloseToParentReaders() error {
	//for parent, _ := range r.parentReaders {
	//	if p, ok := parent.(*ReaderBase); ok {
	//		p.closedByChild = true
	//		//p.refCount.Add(0)
	//		err := p.reportCloseToParentReaders()
	//		if err != nil {
	//			return err
	//		}
	//	}
	//}
	//return nil
	panic("")
}

func (r *baseIndexReader) GetRefCount() int {
	// NOTE: don't ensureOpen, so that callers can see
	// refCount is 0 (reader is closed)
	return int(r.refCount.Load())
}

func (r *baseIndexReader) IncRef() error {
	if !r.TryIncRef() {
		err := r.ensureOpen()
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *baseIndexReader) DecRef() error {
	// only check refcount here (don't call ensureOpen()), so we can
	// still close the reader if it was made invalid by a child:
	if r.refCount.Load() <= 0 {
		return errors.New("this Reader is closed")
	}

	rc := r.refCount.Add(-1)
	if rc == 0 {
		r.closed.Store(true)
		return r.spi.DoClose()
	}

	if rc < 0 {
		return fmt.Errorf("too many decRef calls: refCount is %d after decrement", rc)
	}
	return nil
}

func (r *baseIndexReader) ensureOpen() error {
	if r.refCount.Load() <= 0 {
		return errors.New("this Reader is closed")
	}

	// the happens before rule on reading the refCount, which must be after the fake write,
	// ensures that we see the item:
	if r.closedByChild.Load() {
		return errors.New("this Reader cannot be used anymore as one of its child readers was closed")
	}
	return nil
}

func (r *baseIndexReader) TryIncRef() bool {
	count := int64(0)
	for {
		count = r.refCount.Load()
		if count <= 0 {
			return false
		}
		return r.refCount.Swap(count+1) == count
	}
}

func (r *baseIndexReader) GetTermVector(docID int, field string) (index.Terms, error) {
	vectors, err := r.spi.GetTermVectors(docID)
	if err != nil {
		return nil, err
	}
	return vectors.Terms(field)
}

func (r *baseIndexReader) NumDeletedDocs() int {
	return r.spi.MaxDoc() - r.spi.NumDocs()
}

func (r *baseIndexReader) Document(ctx context.Context, docID int) (*document.Document, error) {
	visitor := document.NewDocumentStoredFieldVisitor()
	if err := r.spi.DocumentWithVisitor(ctx, docID, visitor); err != nil {
		return nil, err
	}
	return visitor.GetDocument(), nil
}

func (r *baseIndexReader) HasDeletions() bool {
	return r.NumDeletedDocs() > 0
}

func (r *baseIndexReader) Leaves() ([]index.LeafReaderContext, error) {
	ctx, err := r.spi.GetContext()
	if err != nil {
		return nil, err
	}
	return ctx.Leaves()
}

var _ sort.Interface = &ReaderSorter{}

type ReaderSorter struct {
	Readers   []index.IndexReader
	FnCompare func(a, b index.LeafReader) int
}

func (r *ReaderSorter) Len() int {
	return len(r.Readers)
}

func (r *ReaderSorter) Less(i, j int) bool {
	return r.FnCompare(r.Readers[i].(index.LeafReader), r.Readers[j].(index.LeafReader)) < 0
}

func (r *ReaderSorter) Swap(i, j int) {
	r.Readers[i], r.Readers[j] = r.Readers[j], r.Readers[i]
}

type ClosedListener interface {
	// Invoked when the resource (segment core, or index reader) that is being cached on is closed.
}
