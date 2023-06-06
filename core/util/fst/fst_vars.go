package fst

type INPUT_TYPE int

const (
	BYTE1 = INPUT_TYPE(iota)
	BYTE2
	BYTE4

	BIT_FINAL_ARC   = 1 << 0
	BIT_LAST_ARC    = 1 << 1
	BIT_TARGET_NEXT = 1 << 2

	// BIT_STOP_NODE TODO: we can free up a bit if we can nuke this:
	BIT_STOP_NODE = 1 << 3

	// BIT_ARC_HAS_OUTPUT This flag is set if the arc has an output.
	BIT_ARC_HAS_OUTPUT = 1 << 4

	BIT_ARC_HAS_FINAL_OUTPUT = 1 << 5

	// ARCS_FOR_BINARY_SEARCH Value of the arc flags to declare a node with fixed length arcs designed for binary search.
	// We use this as a marker because this one flag is illegal by itself.
	ARCS_FOR_BINARY_SEARCH = BIT_ARC_HAS_FINAL_OUTPUT

	// ARCS_FOR_DIRECT_ADDRESSING Value of the arc flags to declare a node with fixed length arcs and bit table designed for direct addressing.
	ARCS_FOR_DIRECT_ADDRESSING = 1 << 6

	// FIXED_LENGTH_ARC_SHALLOW_DEPTH See Also: shouldExpandNodeWithFixedLengthArcs
	// 0 => only root node.
	FIXED_LENGTH_ARC_SHALLOW_DEPTH = 3

	// FIXED_LENGTH_ARC_SHALLOW_NUM_ARCS See Also: shouldExpandNodeWithFixedLengthArcs
	FIXED_LENGTH_ARC_SHALLOW_NUM_ARCS = 5

	// FIXED_LENGTH_ARC_DEEP_NUM_ARCS See Also: shouldExpandNodeWithFixedLengthArcs
	FIXED_LENGTH_ARC_DEEP_NUM_ARCS = 10

	// DIRECT_ADDRESSING_MAX_OVERSIZE_WITH_CREDIT_FACTOR Maximum oversizing factor allowed for direct addressing compared to binary search when expansion credits allow the oversizing. This factor prevents expansions that are obviously too costly even if there are sufficient credits.
	// See Also: shouldExpandNodeWithDirectAddressing
	DIRECT_ADDRESSING_MAX_OVERSIZE_WITH_CREDIT_FACTOR = 1.66

	FILE_FORMAT_NAME = "Fst"
	VERSION_START    = 6
	VERSION_CURRENT  = 7

	// FINAL_END_NODE Never serialized; just used to represent the virtual
	// final node w/ no arcs:
	FINAL_END_NODE = -1

	// NON_FINAL_END_NODE Never serialized; just used to represent the virtual
	// non-final node w/ no arcs:
	NON_FINAL_END_NODE = 0

	// END_LABEL If arc has this label then that arc is final/accepted
	END_LABEL = -1
)
