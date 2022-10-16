package fst

const (
	// BIT_FINAL_ARC arc 对应的字符是不是term最后一个字符
	BIT_FINAL_ARC = 1 << 0

	// BIT_LAST_ARC arc 是不是当前节点的最后一个出度
	BIT_LAST_ARC = 1 << 1

	// BIT_TARGET_NEXT 存储FST的二进制数组中紧邻的下一个字符区间数据是不是当前字符的下一个字符
	BIT_TARGET_NEXT = 1 << 2

	// TODO: we can free up a bit if we can nuke this:

	// BIT_STOP_NODE arc 的 target 是不是一个终止节点
	BIT_STOP_NODE = 1 << 3 //

	// BIT_ARC_HAS_OUTPUT arc 是否有 output 值
	BIT_ARC_HAS_OUTPUT = 1 << 4 // This flag is set if the arc has an output.

	// BIT_ARC_HAS_FINAL_OUTPUT arc 是否有 final output 值
	BIT_ARC_HAS_FINAL_OUTPUT          = 1 << 5                   //
	ARCS_FOR_BINARY_SEARCH            = BIT_ARC_HAS_FINAL_OUTPUT // Value of the arc flags to declare a node with fixed length arcs designed for binary search. We use this as a marker because this one flag is illegal by itself.
	ARCS_FOR_DIRECT_ADDRESSING        = 1 << 6                   // Value of the arc flags to declare a node with fixed length arcs and bit table designed for direct addressing.
	FIXED_LENGTH_ARC_SHALLOW_DEPTH    = 3                        // See Also: shouldExpandNodeWithFixedLengthArcs
	FIXED_LENGTH_ARC_SHALLOW_NUM_ARCS = 5                        // See Also: shouldExpandNodeWithFixedLengthArcs
	FIXED_LENGTH_ARC_DEEP_NUM_ARCS    = 10                       // See Also: shouldExpandNodeWithFixedLengthArcs
)

// Arc Represents a single arc.
type Arc[T any] struct {
	label           int64
	output          *Box[T]
	target          int
	flags           byte
	nextFinalOutput *Box[T]
	nextArc         int
	nodeFlags       byte

	// Fields for arcs belonging to a node with fixed length arcs.
	// So only valid when bytesPerArc != 0.
	// nodeFlags == ARCS_FOR_BINARY_SEARCH || nodeFlags == ARCS_FOR_DIRECT_ADDRESSING.

	bytesPerArc  int
	posArcsStart int
	arcIdx       int
	numArcs      int

	//*** Fields for a direct addressing node. nodeFlags == ARCS_FOR_DIRECT_ADDRESSING.

	// Start position in the FST.BytesReader of the presence bits for a direct addressing node, aka the bit-table
	bitTableStart int

	// First label of a direct addressing node.
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

func (a *Arc[T]) Flag(v int) bool {
	return flag(int(a.flags), v)
}

func (a *Arc[T]) IsLast() bool {
	return a.Flag(BIT_LAST_ARC)
}

func (a *Arc[T]) IsFinal() bool {
	return a.Flag(BIT_FINAL_ARC)
}

func flag(flags, bit int) bool {
	return flags&bit != 0
}

func (a *Arc[T]) Label() int64 {
	return a.label
}

func (a *Arc[T]) Output() *Box[T] {
	return a.output
}

func (a *Arc[T]) Target() int {
	return a.target
}

func (a *Arc[T]) Flags() byte {
	return a.flags
}

func (a *Arc[T]) NextFinalOutput() *Box[T] {
	return a.nextFinalOutput
}

// NextArc Address (into the byte[]) of the next arc - only for list of variable length arc.
// Or ord/address to the next node if label == END_LABEL.
func (a *Arc[T]) NextArc() int {
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
func (a *Arc[T]) PosArcsStart() int {
	return a.posArcsStart
}

// BytesPerArc Non-zero if this arc is part of a node with fixed length arcs, which means all arcs for
// the node are encoded with a fixed number of bytes so that we binary search or direct address.
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
