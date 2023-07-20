package index

import (
	"errors"
	"github.com/geange/lucene-go/core/util/structure"
)

// CompositeReader
// Instances of this reader type can only be used to get stored fields from the underlying LeafReaders,
// but it is not possible to directly retrieve postings. To do that, get the LeafReaderContext for all
// sub-readers via leaves().
//
// Reader instances for indexes on disk are usually constructed with a call to one of the static
// DirectoryReader.open() methods, e.g. DirectoryReader.open(Directory). DirectoryReader implements
// the CompositeReader interface, it is not possible to directly get postings.
// Concrete subclasses of Reader are usually constructed with a call to one of the static open()
// methods, e.g. DirectoryReader.open(Directory).
//
// For efficiency, in this API documents are often referred to via document numbers, non-negative integers
// which each name a unique document in the index. These document numbers are ephemeral -- they may change
// as documents are added to and deleted from an index. Clients should thus not rely on a given document
// having the same number between sessions.
//
// NOTE: Reader instances are completely thread safe, meaning multiple threads can call any of its
// methods, concurrently. If your application requires external synchronization, you should not
// synchronize on the Reader instance; use your own (non-Lucene) objects instead.
type CompositeReader interface {
	Reader

	// GetSequentialSubReaders
	// Expert: returns the sequential sub readers that this reader is logically composed of.
	// This method may not return null.
	// NOTE: In contrast to previous Lucene versions this method is no longer public, code that
	// wants to get all LeafReaders this composite is composed of should use Reader.leaves().
	// See Also: Reader.leaves()
	GetSequentialSubReaders() []Reader
}

var _ ReaderContext = &CompositeReaderContext{}

// CompositeReaderContext ReaderContext for CompositeReader instance.
type CompositeReaderContext struct {
	*IndexReaderContextDefault

	children *structure.ArrayList[ReaderContext]
	leaves   *structure.ArrayList[ReaderContext]
	reader   CompositeReader
}

func NewCompositeReaderContext(reader CompositeReader) (*CompositeReaderContext, error) {
	return NewCompositeReaderBuilder(reader).Build()
}

func NewCompositeReaderContextV1(parent *CompositeReaderContext, reader CompositeReader,
	ordInParent, docbaseInParent int, children *structure.ArrayList[ReaderContext]) *CompositeReaderContext {
	return newCompositeReaderContext(parent, reader, ordInParent, docbaseInParent, children, nil)
}

func NewCompositeReaderContextV2(reader CompositeReader,
	children, leaves *structure.ArrayList[ReaderContext]) *CompositeReaderContext {
	return newCompositeReaderContext(nil, reader, 0, 0, children, leaves)
}

func newCompositeReaderContext(parent *CompositeReaderContext, reader CompositeReader,
	ordInParent, docbaseInParent int,
	children, leaves *structure.ArrayList[ReaderContext]) *CompositeReaderContext {

	return &CompositeReaderContext{
		IndexReaderContextDefault: NewIndexReaderContextDefault(parent, ordInParent, docbaseInParent),
		children:                  children,
		leaves:                    leaves,
		reader:                    reader,
	}
}

func (c *CompositeReaderContext) Reader() Reader {
	return c.reader
}

func (c *CompositeReaderContext) Leaves() ([]*LeafReaderContext, error) {
	if !c.IsTopLevel {
		return nil, errors.New("this is not a top-level context")
	}

	leaves := make([]*LeafReaderContext, 0, c.leaves.Size())
	values := c.leaves.ToArray()
	for i := range values {
		leaves = append(leaves, values[i].(*LeafReaderContext))
	}
	return leaves, nil
}

func (c *CompositeReaderContext) Children() []ReaderContext {
	return c.children.ToArray()
}

type CompositeReaderBuilder struct {
	reader CompositeReader
	//leaves      []ReaderContext
	leaves      *structure.ArrayList[ReaderContext]
	leafDocBase int
}

func NewCompositeReaderBuilder(reader CompositeReader) *CompositeReaderBuilder {
	return &CompositeReaderBuilder{
		reader: reader,
		leaves: structure.NewArrayList[ReaderContext](),
	}
}

func (c *CompositeReaderBuilder) Build() (*CompositeReaderContext, error) {
	v, err := c.build(nil, c.reader, 0, 0)
	if err != nil {
		return nil, err
	}
	return v.(*CompositeReaderContext), nil
}

func (c *CompositeReaderBuilder) build(parent *CompositeReaderContext, reader Reader,
	ord, docBase int) (ReaderContext, error) {
	if ar, ok := reader.(LeafReader); ok {
		ctx := NewLeafReaderContextV1(parent, ar, ord, docBase, c.leaves.Size(), c.leafDocBase)
		c.leaves.Add(ctx)
		c.leafDocBase += reader.MaxDoc()
		return ctx, nil
	}

	cr := reader.(CompositeReader)
	sequentialSubReaders := cr.GetSequentialSubReaders()
	children := structure.NewArrayListArray(make([]ReaderContext, len(sequentialSubReaders)))
	var newParent *CompositeReaderContext
	if parent == nil {
		newParent = NewCompositeReaderContextV2(cr, children, c.leaves)
	} else {
		newParent = NewCompositeReaderContextV1(parent, cr, ord, docBase, children)
	}

	newDocBase := 0

	for i := 0; i < len(sequentialSubReaders); i++ {
		r := sequentialSubReaders[i]

		readerContext, err := c.build(newParent, r, i, newDocBase)
		if err != nil {
			return nil, err
		}
		if err := children.Set(i, readerContext); err != nil {
			return nil, err
		}
		newDocBase += r.MaxDoc()
	}
	//assert newDocBase == cr.maxDoc();
	return newParent, nil
}
