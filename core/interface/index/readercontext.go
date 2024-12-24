package index

// IndexReaderContext
// represents a hierarchical relationship between IndexReader instances.
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

// LeafReaderContext
// IndexReaderContext for LeafReader instances.
type LeafReaderContext interface {
	Reader() IndexReader
	Leaves() ([]LeafReaderContext, error)
	Children() []IndexReaderContext
	Identity() string

	Ord() int
	DocBase() int
	LeafReader() LeafReader
}
