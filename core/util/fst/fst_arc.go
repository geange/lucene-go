package fst

type FSTArc struct {
	label           int
	output          any
	target          int64
	flags           byte
	nextFinalOutput any
	nextArc         int64
	nodeFlags       byte

	//*** Fields for arcs belonging to a node with fixed length arcs.
	// So only valid when bytesPerArc != 0.
	// nodeFlags == ARCS_FOR_BINARY_SEARCH || nodeFlags == ARCS_FOR_DIRECT_ADDRESSING.

	bytesPerArc  int
	posArcsStart int64
	arcIdx       int
	numArcs      int

	// *** Fields for a direct addressing node. nodeFlags == ARCS_FOR_DIRECT_ADDRESSING.

	// Start position in the FST.BytesReader of the presence bits for a direct addressing node, aka the bit-table
	bitTableStart int64

	// First label of a direct addressing node.
	firstLabel int

	// Index of the current label of a direct addressing node. While arcIdx is the current index in the label range, presenceIndex is its corresponding index in the list of actually present labels. It is equal to the number of bits set before the bit at arcIdx in the bit-table. This field is a cache to avoid to count bits set repeatedly when iterating the next arcs.
	presenceIndex int
}

func (r *FSTArc) flag(value int) bool {
	return flag(int(r.flags), value)
}

func (r *FSTArc) IsLast() bool {
	return r.flag(BIT_LAST_ARC)
}

func (r *FSTArc) IsFinal() bool {
	return r.flag(BIT_FINAL_ARC)
}

func (r *FSTArc) Label() int {
	return r.label
}

func (r *FSTArc) Output() any {
	return r.output
}

// Target Ord/address to target node.
func (r *FSTArc) Target() any {
	return r.target
}

func (r *FSTArc) Flags() byte {
	return r.flags
}

func (r *FSTArc) NextFinalOutput() any {
	return r.nextFinalOutput
}

// NextArc Address (into the byte[]) of the next arc - only for list of variable length arc.
// Or ord/address to the next node if label == END_LABEL.
func (r *FSTArc) NextArc() int64 {
	return r.nextArc
}

// ArcIdx Where we are in the array; only valid if bytesPerArc != 0.
func (r *FSTArc) ArcIdx() int {
	return r.arcIdx
}

// NodeFlags Node header flags. Only meaningful to check if the value is either ARCS_FOR_BINARY_SEARCH
// or ARCS_FOR_DIRECT_ADDRESSING (other value when bytesPerArc == 0).
func (r *FSTArc) NodeFlags() byte {
	return r.nodeFlags
}

// PosArcsStart Where the first arc in the array starts; only valid if bytesPerArc != 0
func (r *FSTArc) PosArcsStart() int64 {
	return r.posArcsStart
}

// BytesPerArc Non-zero if this arc is part of a node with fixed length arcs,
// which means all arcs for the node are encoded with a fixed number of bytes
// so that we binary search or direct address. We do when there are enough arcs leaving one node.
// It wastes some bytes but gives faster lookups.
func (r *FSTArc) BytesPerArc() int {
	return r.bytesPerArc
}

// NumArcs How many arcs; only valid if bytesPerArc != 0 (fixed length arcs).
// For a node designed for binary search this is the array size.
// For a node designed for direct addressing, this is the label range.
func (r *FSTArc) NumArcs() int {
	return r.numArcs
}

// FirstLabel First label of a direct addressing node. Only valid if nodeFlags == ARCS_FOR_DIRECT_ADDRESSING.
func (r *FSTArc) FirstLabel() int {
	return r.firstLabel
}
