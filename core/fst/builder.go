package fst

// Builder Builds a minimal FST (maps an IntsRef term to an arbitrary output) from pre-sorted terms with outputs. The FST becomes an FSA if you use NoOutputs. The FST is written on-the-fly into a compact serialized format byte array, which can be saved to / loaded from a Directory or used directly for traversal. The FST is always finite (no cycles).
// NOTE: The algorithm is described at http://citeseerx.ist.psu.edu/viewdoc/summary?doi=10.1.1.24.3698
// The parameterized type T is the output type. See the subclasses of Outputs.
// FSTs larger than 2.1GB are now possible (as of Lucene 4.2). FSTs containing more than 2.1B nodes are also now possible, however they cannot be packed.
// lucene.experimental
type Builder[T any] struct {
	NO_OUTPUT any
}

// Node NOTE: not many instances of Node or CompiledNode are in
// memory while the FST is being built; it's only the
// current "frontier":
type Node interface {
	IsCompiled() bool
}

var _ Node = &CompiledNode{}

type CompiledNode struct {
	node int64
}

func (c *CompiledNode) IsCompiled() bool {
	return true
}

//var _ Node = &UnCompiledNode[any]{}

type UnCompiledNode[T any] struct {
	owner      *Builder[T]
	numArcs    int
	arcs       []Arc[T]
	output     any
	isFinal    bool
	inputCount int64
	depth      int
}

func (u *UnCompiledNode[T]) IsCompiled() bool {
	return false
}

func (u *UnCompiledNode[T]) Clear() {
	u.numArcs = 0
	u.isFinal = false
	u.output = u.owner.NO_OUTPUT
	u.inputCount = 0

	// We don't clear the depth here because it never changes
	// for nodes on the frontier (even when reused).
}
