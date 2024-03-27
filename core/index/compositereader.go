package index

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/geange/gods-generic/lists/arraylist"
	"github.com/geange/lucene-go/core/document"
)

// CompositeReader
// Instances of this reader type can only be used to get stored fields from the underlying LeafReaders,
// but it is not possible to directly retrieve postings. To do that, get the LeafReaderContext for all
// sub-readers via leaves().
//
// IndexReader instances for indexes on disk are usually constructed with a call to one of the static
// DirectoryReader.open() methods, e.g. DirectoryReader.open(Directory). DirectoryReader implements
// the CompositeReader interface, it is not possible to directly get postings.
// Concrete subclasses of IndexReader are usually constructed with a call to one of the static open()
// methods, e.g. DirectoryReader.open(Directory).
//
// For efficiency, in this API documents are often referred to via document numbers, non-negative integers
// which each name a unique document in the index. These document numbers are ephemeral -- they may change
// as documents are added to and deleted from an index. Clients should thus not rely on a given document
// having the same number between sessions.
//
// NOTE: IndexReader instances are completely thread safe, meaning multiple threads can call any of its
// methods, concurrently. If your application requires external synchronization, you should not
// synchronize on the IndexReader instance; use your own (non-Lucene) objects instead.
type CompositeReader interface {
	IndexReader

	// GetSequentialSubReaders
	// Expert: returns the sequential sub readers that this reader is logically composed of.
	// This method may not return null.
	// NOTE: In contrast to previous Lucene versions this method is no longer public, code that
	// wants to get all LeafReaders this composite is composed of should use Reader.leaves().
	// See Also: Reader.leaves()
	GetSequentialSubReaders() []IndexReader
}

var _ CompositeReader = &baseCompositeReader{}

type baseCompositeReader struct {
	*baseIndexReader

	subReaders       []IndexReader             //
	subReadersSorter func(a, b LeafReader) int //
	starts           []int                     // 1st docno for each reader
	maxDoc           int                       //
	numDocs          int                       // computed lazily
	subReadersList   []IndexReader             // List view solely for getSequentialSubReaders(), for effectiveness the array is used internally.
	readerContext    *CompositeReaderContext   //
}

func (b *baseCompositeReader) DoClose() error {
	return nil
}

func (b *baseCompositeReader) GetContext() (IndexReaderContext, error) {
	if b.readerContext == nil {
		readerContext, err := NewCompositeReaderContext(WithCompositeReaderContextV1(b))
		if err != nil {
			return nil, err
		}
		b.readerContext = readerContext
	}
	return b.readerContext, nil
}

func (b *baseCompositeReader) GetMetaData() *LeafMetaData {
	//TODO implement me
	panic("implement me")
}

func (b *baseCompositeReader) GetReaderCacheHelper() CacheHelper {
	//TODO implement me
	panic("implement me")
}

func (b *baseCompositeReader) GetSequentialSubReaders() []IndexReader {
	return b.subReadersList
}

func newBaseCompositeReader(subReaders []IndexReader,
	subReadersSorter func(a, b LeafReader) int) (*baseCompositeReader, error) {

	sort.Sort(&ReaderSorter{
		Readers:   subReaders,
		FnCompare: subReadersSorter,
	})

	reader := &baseCompositeReader{
		subReaders:       subReaders,
		subReadersSorter: subReadersSorter,
		starts:           make([]int, len(subReaders)+1),
		maxDoc:           0,
		numDocs:          -1,
		subReadersList:   subReaders,
	}

	reader.baseIndexReader = newBaseIndexReader(reader)

	maxDoc := 0
	for i := 0; i < len(subReaders); i++ {
		reader.starts[i] = maxDoc
		r := subReaders[i]
		maxDoc += r.MaxDoc() // compute maxDocs
		//r.RegisterParentReader(reader)
	}

	if maxDoc > GetActualMaxDocs() {
		return nil, errors.New("too many documents")
	}

	reader.maxDoc = maxDoc
	reader.starts[len(subReaders)] = maxDoc

	return reader, nil
}

func (b *baseCompositeReader) GetTermVectors(docID int) (Fields, error) {
	i, err := b.readerIndex(docID) // find subreader num
	if err != nil {
		return nil, err
	}
	return b.subReaders[i].GetTermVectors(docID - b.starts[i]) // dispatch to subreader
}

func (b *baseCompositeReader) NumDocs() int {
	// Don't call ensureOpen() here (it could affect performance)
	// We want to compute numDocs() lazily so that creating a wrapper that hides
	// some documents isn't slow at wrapping time, but on the first time that
	// numDocs() is called. This can help as there are lots of use-cases of a
	// reader that don't involve calling numDocs().
	// However it's not crucial to make sure that we don't call numDocs() more
	// than once on the sub readers, since they likely cache numDocs() anyway,
	// hence the lack of synchronization.
	numDocs := b.numDocs
	if numDocs == -1 {
		numDocs = 0
		for _, r := range b.subReaders {
			numDocs += r.NumDocs()
		}
		//assert numDocs >= 0;
		b.numDocs = numDocs
	}
	return numDocs
}

func (b *baseCompositeReader) MaxDoc() int {
	// Don't call ensureOpen() here (it could affect performance)
	return b.maxDoc
}

func (b *baseCompositeReader) DocumentWithVisitor(docID int, visitor document.StoredFieldVisitor) error {
	//ensureOpen();
	i, err := b.readerIndex(docID) // find subreader num
	if err != nil {
		return err
	}
	return b.subReaders[i].DocumentWithVisitor(docID-b.starts[i], visitor) // dispatch to subreader
}

func (b *baseCompositeReader) DocFreq(ctx context.Context, term Term) (int, error) {
	//ensureOpen();
	total := 0 // sum freqs in subreaders
	for i := 0; i < len(b.subReaders); i++ {
		sub, err := b.subReaders[i].DocFreq(ctx, term)
		if err != nil {
			return 0, err
		}
		//assert sub >= 0;
		//assert sub <= subReaders[i].getDocCount(term.field());
		total += sub
	}
	return total, nil
}

func (b *baseCompositeReader) TotalTermFreq(ctx context.Context, term *Term) (int64, error) {
	//ensureOpen();
	total := int64(0) // sum freqs in subreaders
	for i := 0; i < len(b.subReaders); i++ {
		sub, err := b.subReaders[i].TotalTermFreq(ctx, term)
		if err != nil {
			return 0, err
		}
		//assert sub >= 0;
		//assert sub <= subReaders[i].getSumTotalTermFreq(term.field());
		total += sub
	}
	return total, nil
}

func (b *baseCompositeReader) GetSumDocFreq(field string) (int64, error) {
	//ensureOpen();
	total := int64(0) // sum doc freqs in subreaders
	for _, reader := range b.subReaders {
		sub, err := reader.GetSumDocFreq(field)
		if err != nil {
			return 0, err
		}
		//assert sub >= 0;
		//assert sub <= reader.getSumTotalTermFreq(field);
		total += sub
	}
	return total, nil
}

func (b *baseCompositeReader) GetDocCount(field string) (int, error) {
	//ensureOpen();
	total := 0 // sum doc counts in subreaders
	for _, reader := range b.subReaders {
		sub, err := reader.GetDocCount(field)
		if err != nil {
			return 0, err
		}
		//assert sub >= 0;
		//assert sub <= reader.maxDoc();
		total += sub
	}
	return total, nil
}

func (b *baseCompositeReader) GetSumTotalTermFreq(field string) (int64, error) {
	//ensureOpen();
	total := int64(0) // sum doc total term freqs in subreaders
	for _, reader := range b.subReaders {
		sub, err := reader.GetSumTotalTermFreq(field)
		if err != nil {
			return 0, err
		}
		//assert sub >= 0;
		//assert sub >= reader.getSumDocFreq(field);
		total += sub
	}
	return total, nil
}

func (b *baseCompositeReader) readerIndex(docID int) (int, error) {
	if docID < 0 || docID >= b.maxDoc {
		return 0, fmt.Errorf("docID must be >= 0 and < maxDoc=%d (got docID=%d)", b.maxDoc, docID)
	}
	return SubIndex(docID, b.starts), nil
}

var _ IndexReaderContext = &CompositeReaderContext{}

// CompositeReaderContext IndexReaderContext for CompositeReader instance.
type CompositeReaderContext struct {
	*BaseIndexReaderContext

	children *arraylist.List[IndexReaderContext]
	leaves   *arraylist.List[IndexReaderContext]
	reader   CompositeReader
}

type compositeReaderContextOption struct {
	opt1 *compositeReaderContextOptionV1
	opt2 *compositeReaderContextOptionV2
	opt3 *compositeReaderContextOptionV3
}

type compositeReaderContextOptionV1 struct {
	reader CompositeReader
}

type compositeReaderContextOptionV2 struct {
	parent          *CompositeReaderContext
	reader          CompositeReader
	ordInParent     int
	docbaseInParent int
	children        *arraylist.List[IndexReaderContext]
}

type compositeReaderContextOptionV3 struct {
	reader           CompositeReader
	children, leaves *arraylist.List[IndexReaderContext]
}

type CompositeReaderContextOption func(*compositeReaderContextOption)

func WithCompositeReaderContextV1(reader CompositeReader) CompositeReaderContextOption {
	return func(o *compositeReaderContextOption) {
		o.opt1 = &compositeReaderContextOptionV1{reader: reader}
	}
}

func WithCompositeReaderContextV2(parent *CompositeReaderContext, reader CompositeReader,
	ordInParent, docbaseInParent int, children *arraylist.List[IndexReaderContext]) CompositeReaderContextOption {
	return func(o *compositeReaderContextOption) {
		o.opt2 = &compositeReaderContextOptionV2{
			parent:          parent,
			reader:          reader,
			ordInParent:     ordInParent,
			docbaseInParent: docbaseInParent,
			children:        children,
		}
	}
}

func WithCompositeReaderContextV3(reader CompositeReader,
	children, leaves *arraylist.List[IndexReaderContext]) CompositeReaderContextOption {
	return func(o *compositeReaderContextOption) {
		o.opt3 = &compositeReaderContextOptionV3{
			reader:   reader,
			children: children,
			leaves:   leaves,
		}
	}
}

func NewCompositeReaderContext(fn CompositeReaderContextOption) (*CompositeReaderContext, error) {
	opt := &compositeReaderContextOption{}
	fn(opt)

	if opt.opt1 != nil {
		return NewCompositeReaderBuilder(opt.opt1.reader).Build()
	}

	if opt.opt2 != nil {
		option := opt.opt2
		return newCompositeReaderContext(option.parent, option.reader, option.ordInParent, option.docbaseInParent, option.children, nil), nil
	}

	if opt.opt3 != nil {
		option := opt.opt3
		return newCompositeReaderContext(nil, option.reader, 0, 0, option.children, option.leaves), nil
	}

	return nil, errors.New("todo")
}

func newCompositeReaderContext(parent *CompositeReaderContext, reader CompositeReader,
	ordInParent, docbaseInParent int,
	children, leaves *arraylist.List[IndexReaderContext]) *CompositeReaderContext {

	return &CompositeReaderContext{
		BaseIndexReaderContext: NewBaseIndexReaderContext(parent, ordInParent, docbaseInParent),
		children:               children,
		leaves:                 leaves,
		reader:                 reader,
	}
}

func (c *CompositeReaderContext) Reader() IndexReader {
	return c.reader
}

func (c *CompositeReaderContext) Leaves() ([]LeafReaderContext, error) {
	if !c.isTopLevel {
		return nil, errors.New("this is not a top-level context")
	}

	leaves := make([]LeafReaderContext, 0, c.leaves.Size())
	values := c.leaves.Values()
	for i := range values {
		leaves = append(leaves, values[i].(LeafReaderContext))
	}
	return leaves, nil
}

func (c *CompositeReaderContext) Children() []IndexReaderContext {
	return c.children.Values()
}

type CompositeReaderBuilder struct {
	reader CompositeReader
	//leaves      []ReaderContext
	leaves      *arraylist.List[IndexReaderContext]
	leafDocBase int
}

func NewCompositeReaderBuilder(reader CompositeReader) *CompositeReaderBuilder {
	return &CompositeReaderBuilder{
		reader: reader,
		leaves: arraylist.New[IndexReaderContext](),
	}
}

func (c *CompositeReaderBuilder) Build() (*CompositeReaderContext, error) {
	v, err := c.build(nil, c.reader, 0, 0)
	if err != nil {
		return nil, err
	}
	return v.(*CompositeReaderContext), nil
}

func (c *CompositeReaderBuilder) build(parent *CompositeReaderContext, reader IndexReader,
	ord, docBase int) (IndexReaderContext, error) {
	if ar, ok := reader.(LeafReader); ok {
		ctx := NewLeafReaderContextV1(parent, ar, ord, docBase, c.leaves.Size(), c.leafDocBase)
		c.leaves.Add(ctx)
		c.leafDocBase += reader.MaxDoc()
		return ctx, nil
	}

	cr := reader.(CompositeReader)
	sequentialSubReaders := cr.GetSequentialSubReaders()
	children := arraylist.New[IndexReaderContext](make([]IndexReaderContext, len(sequentialSubReaders))...)
	var newParent *CompositeReaderContext
	if parent == nil {
		newParent, _ = NewCompositeReaderContext(WithCompositeReaderContextV3(cr, children, c.leaves))
	} else {
		newParent, _ = NewCompositeReaderContext(WithCompositeReaderContextV2(parent, cr, ord, docBase, children))
	}

	newDocBase := 0

	for i := 0; i < len(sequentialSubReaders); i++ {
		r := sequentialSubReaders[i]

		readerContext, err := c.build(newParent, r, i, newDocBase)
		if err != nil {
			return nil, err
		}
		children.Set(i, readerContext)
		newDocBase += r.MaxDoc()
	}
	//assert newDocBase == cr.maxDoc();
	return newParent, nil
}
