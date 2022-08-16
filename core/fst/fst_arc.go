package fst

// Arc Represents a single arc.
type Arc[T any] struct {
	label           int
	output          T
	target          int64
	flags           byte
	nextFinalOutput T
	nextArc         int64
	nodeFlags       byte

	//*** Fields for arcs belonging to a node with fixed length arcs.
	// So only valid when bytesPerArc != 0.
	// nodeFlags == ARCS_FOR_BINARY_SEARCH || nodeFlags == ARCS_FOR_DIRECT_ADDRESSING.
	bytesPerArc  int
	posArcsStart int64
	arcIdx       int
	numArcs      int

	//*** Fields for a direct addressing node. nodeFlags == ARCS_FOR_DIRECT_ADDRESSING.

	//Start position in the FST.BytesReader of the presence bits for a direct addressing node, aka the bit-table
	bitTableStart int64

	firstLabel int

	// Index of the current label of a direct addressing node. While arcIdx is the current index in the label range, presenceIndex is its corresponding index in the list of actually present labels. It is equal to the number of bits set before the bit at arcIdx in the bit-table. This field is a cache to avoid to count bits set repeatedly when iterating the next arcs.
	presenceIndex int
}

func (a *Arc[T]) CopyFrom(other *Arc[T]) *Arc[T] {
	a.label = other.Label()
	a.target = other.Target()
	a.flags = other.Flags()
	a.output = other.Output()
	a.nextFinalOutput = other.NextFinalOutput()
	a.nextArc = other.NextArc()
	a.nodeFlags = other.NodeFlags()
	a.bytesPerArc = other.BytesPerArc()

	// Fields for arcs belonging to a node with fixed length arcs.
	// We could avoid copying them if bytesPerArc() == 0 (this was the case with previous code, and the current code
	// still supports that), but it may actually help external uses of FST to have consistent arc state, and debugging
	// is easier.
	a.posArcsStart = other.PosArcsStart()
	a.arcIdx = other.ArcIdx()
	a.numArcs = other.NumArcs()
	a.bitTableStart = other.bitTableStart
	a.firstLabel = other.FirstLabel()
	a.presenceIndex = other.presenceIndex
	return a
}

func (a *Arc[T]) IsLast() bool {
	return flag(int(a.flags), BIT_FINAL_ARC)
}

func (a *Arc[T]) IsFinal() bool {
	return flag(int(a.flags), BIT_FINAL_ARC)
}

func (a *Arc[T]) Label() int {
	return a.label
}

func (a *Arc[T]) Output() T {
	return a.output
}

func (a *Arc[T]) Target() int64 {
	return a.target
}

func (a *Arc[T]) Flags() byte {
	return a.flags
}

func (a *Arc[T]) NextFinalOutput() T {
	return a.nextFinalOutput
}

// NextArc Address (into the byte[]) of the next arc - only for list of variable length arc.
// Or ord/address to the next node if label == END_LABEL.
func (a *Arc[T]) NextArc() int64 {
	return a.nextArc
}

// ArcIdx Where we are in the array; only valid if bytesPerArc != 0.
func (a *Arc[T]) ArcIdx() int {
	return a.arcIdx
}

// NodeFlags Node header flags. Only meaningful to check if the value is either ARCS_FOR_BINARY_SEARCH
// or ARCS_FOR_DIRECT_ADDRESSING (other value when bytesPerArc == 0).
func (a *Arc[T]) NodeFlags() byte {
	return a.nodeFlags
}

// PosArcsStart Where the first arc in the array starts; only valid if bytesPerArc != 0
func (a *Arc[T]) PosArcsStart() int64 {
	return a.posArcsStart
}

// BytesPerArc Non-zero if this arc is part of a node with fixed length arcs, which means all arcs
// for the node are encoded with a fixed number of bytes so that we binary search or direct address.
// We do when there are enough arcs leaving one node. It wastes some bytes but gives faster lookups.
func (a *Arc[T]) BytesPerArc() int {
	return a.bytesPerArc
}

// NumArcs How many arcs; only valid if bytesPerArc != 0 (fixed length arcs). For a node designed for
// binary search this is the array size. For a node designed for direct addressing, this is the label range.
func (a *Arc[T]) NumArcs() int {
	return a.numArcs
}

// FirstLabel First label of a direct addressing node. Only valid if nodeFlags == ARCS_FOR_DIRECT_ADDRESSING.
func (a *Arc[T]) FirstLabel() int {
	return a.firstLabel
}
