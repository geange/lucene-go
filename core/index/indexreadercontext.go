package index

import "github.com/google/uuid"

// IndexReaderContext A struct like class that represents a hierarchical relationship between IndexReader instances.
type IndexReaderContext interface {

	// Reader Returns the IndexReader, this context represents.
	Reader() IndexReader

	// Leaves
	// Returns the context's leaves if this context is a top-level context. For convenience, if this is
	// an LeafReaderContextImpl this returns itself as the only leaf.
	// Note: this is convenience method since leaves can always be obtained by walking the context tree
	// using children().
	// Throws: ErrUnsupportedOperation – if this is not a top-level context.
	// See Also: children()
	Leaves() ([]LeafReaderContext, error)

	// Children Returns the context's children iff this context is a composite context otherwise null.
	Children() []IndexReaderContext

	Identity() string
}

type BaseIndexReaderContext struct {
	// The reader context for this reader's immediate parent, or null if none
	parent *CompositeReaderContext

	// true if this context struct represents the top level reader within the hierarchical context
	isTopLevel bool

	// the doc base for this reader in the parent, 0 if parent is null
	docBaseInParent int

	// the ord for this reader in the parent, 0 if parent is null
	ordInParent int

	identity string
}

func NewBaseIndexReaderContext(parent *CompositeReaderContext, ordInParent, docBaseInParent int) *BaseIndexReaderContext {
	isTop := parent == nil
	return &BaseIndexReaderContext{
		parent:          parent,
		isTopLevel:      isTop,
		docBaseInParent: docBaseInParent,
		ordInParent:     ordInParent,
		identity:        uuid.New().String(),
	}
}

func (r *BaseIndexReaderContext) Identity() string {
	return r.identity
}
