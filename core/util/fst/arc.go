package fst

// Arc Represents a single arc.
type Arc struct {
	flags           byte
	nodeFlags       byte
	label           int
	output          Output
	target          int64
	nextFinalOutput Output
	nextArc         int64

	// Fields for arcs belonging to a node with fixed length arcs.
	// So only valid when bytesPerArc != 0.
	// nodeFlags == ARCS_FOR_BINARY_SEARCH || nodeFlags == ARCS_FOR_DIRECT_ADDRESSING.

	bytesPerArc  int
	posArcsStart int64
	arcIdx       int
	numArcs      int

	// Fields for a direct addressing node. nodeFlags == ARCS_FOR_DIRECT_ADDRESSING.

	// Start position in the Fst.BytesReader of the presence bits for a direct addressing node, aka the bit-table
	bitTableStart int64

	// First label of a direct addressing node.
	// 第一个label的值
	firstLabel int

	// Index of the current label of a direct addressing node. While arcIdx is the current index in the label range,
	// presenceIndex is its corresponding index in the list of actually present labels. It is equal to the number
	// of bits set before the bit at arcIdx in the bit-table. This field is a cache to avoid to count bits set
	// repeatedly when iterating the next arcs.
	presenceIndex int
}

func (r *Arc) matchFlag(value int) bool {
	return flag(int(r.flags), value)
}

func (r *Arc) IsLast() bool {
	return r.matchFlag(BitLastArc)
}

func (r *Arc) IsFinal() bool {
	return r.matchFlag(BitFinalArc)
}

func (r *Arc) Label() int {
	return r.label
}

func (r *Arc) Output() Output {
	return r.output
}

// Target Ord/address to target node.
func (r *Arc) Target() int64 {
	return r.target
}

func (r *Arc) Flags() byte {
	return r.flags
}

func (r *Arc) NextFinalOutput() Output {
	return r.nextFinalOutput
}

// NextArc Address (into the byte[]) of the next arc - only for list of variable length arc.
// Or ord/address to the next node if label == END_LABEL.
func (r *Arc) NextArc() int64 {
	return r.nextArc
}

// ArcIdx Where we are in the array; only valid if bytesPerArc != 0.
func (r *Arc) ArcIdx() int {
	return r.arcIdx
}

// NodeFlags Node header flags. Only meaningful to check if the value is either ArcsForBinarySearch
// or ArcsForDirectAddressing (other value when bytesPerArc == 0).
func (r *Arc) NodeFlags() byte {
	return r.nodeFlags
}

// PosArcsStart Where the first arc in the array starts; only valid if bytesPerArc != 0
func (r *Arc) PosArcsStart() int64 {
	return r.posArcsStart
}

// BytesPerArc Non-zero if this arc is part of a node with fixed length arcs,
// which means all arcs for the node are encoded with a fixed number of bytes
// so that we binary search or direct address. We do when there are enough arcs leaving one node.
// It wastes some bytes but gives faster lookups.
func (r *Arc) BytesPerArc() int {
	return r.bytesPerArc
}

// NumArcs How many arcs; only valid if bytesPerArc != 0 (fixed length arcs).
// For a node designed for binary search this is the array size.
// For a node designed for direct addressing, this is the label range.
func (r *Arc) NumArcs() int {
	return r.numArcs
}

// FirstLabel First label of a direct addressing node. Only valid if nodeFlags == ArcsForDirectAddressing.
func (r *Arc) FirstLabel() int {
	return r.firstLabel
}
