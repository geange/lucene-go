package fst

type InputType int

const (
	BitFinalArc          = 1 << 0
	BitLastArc           = 1 << 1
	BitTargetNext        = 1 << 2
	BitStopNode          = 1 << 3
	BitArcHasOutput      = 1 << 4 // This flag is set if the arc has an output.
	BitArcHasFinalOutput = 1 << 5

	// ArcsForBinarySearch
	// value of the arc flags to declare a node with fixed length arcs designed for binary search.
	// We use this as a marker because this one flag is illegal by itself.
	ArcsForBinarySearch = BitArcHasFinalOutput

	// ArcsForDirectAddressing
	// value of the arc flags to declare a node with fixed length arcs and bit table designed for direct addressing.
	ArcsForDirectAddressing = 1 << 6
)

const (
	DEFAULT_MAX_BLOCK_BITS = 30
	INTEGER_BYTES          = 4

	BYTE1 = InputType(iota)
	BYTE2
	BYTE4

	// FIXED_LENGTH_ARC_SHALLOW_DEPTH
	// See Also: shouldExpandNodeWithFixedLengthArcs
	// 0 => only root node.
	FIXED_LENGTH_ARC_SHALLOW_DEPTH = 3

	// FIXED_LENGTH_ARC_SHALLOW_NUM_ARCS
	// See Also: shouldExpandNodeWithFixedLengthArcs
	FIXED_LENGTH_ARC_SHALLOW_NUM_ARCS = 5

	// FIXED_LENGTH_ARC_DEEP_NUM_ARCS
	// See Also: shouldExpandNodeWithFixedLengthArcs
	FIXED_LENGTH_ARC_DEEP_NUM_ARCS = 10

	// DIRECT_ADDRESSING_MAX_OVERSIZE_WITH_CREDIT_FACTOR
	// Maximum oversizing factor allowed for direct addressing compared to binary search
	// when expansion credits allow the oversizing. This factor prevents expansions
	// that are obviously too costly even if there are sufficient credits.
	// See Also: shouldExpandNodeWithDirectAddressing
	DIRECT_ADDRESSING_MAX_OVERSIZE_WITH_CREDIT_FACTOR = 1.66

	FILE_FORMAT_NAME = "FST"
	VERSION_START    = 6
	VERSION_CURRENT  = 7

	// FINAL_END_NODE
	// Never serialized; just used to represent the virtual
	// final node w/ no arcs:
	FINAL_END_NODE = -1

	// NON_FINAL_END_NODE
	// Never serialized; just used to represent the virtual
	// non-final node w/ no arcs:
	NON_FINAL_END_NODE = 0

	// END_LABEL
	// If arc has this label then that arc is final/accepted
	END_LABEL = -1
)
