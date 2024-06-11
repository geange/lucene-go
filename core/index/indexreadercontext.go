package index

import "github.com/google/uuid"

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
