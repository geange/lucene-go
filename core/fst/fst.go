package fst

import "github.com/geange/lucene-go/core/store"

var (
	BIT_FINAL_ARC   = 1 << 0
	BIT_LAST_ARC    = 1 << 1
	BIT_TARGET_NEXT = 1 << 2

	// BIT_STOP_NODE TODO: we can free up a bit if we can nuke this:
	BIT_STOP_NODE = 1 << 3

	// BIT_ARC_HAS_OUTPUT This flag is set if the arc has an output.
	BIT_ARC_HAS_OUTPUT = 1 << 4

	BIT_ARC_HAS_FINAL_OUTPUT = 1 << 5

	// ARCS_FOR_BINARY_SEARCH Value of the arc flags to declare a node with fixed length arcs designed for binary search.
	ARCS_FOR_BINARY_SEARCH = byte(BIT_ARC_HAS_FINAL_OUTPUT)

	// ARCS_FOR_DIRECT_ADDRESSING Value of the arc flags to declare a node with fixed length arcs and
	// bit table designed for direct addressing.
	ARCS_FOR_DIRECT_ADDRESSING = 1 << 6

	// FIXED_LENGTH_ARC_SHALLOW_DEPTH See Also: shouldExpandNodeWithFixedLengthArcs
	FIXED_LENGTH_ARC_SHALLOW_DEPTH = 3

	// FIXED_LENGTH_ARC_SHALLOW_NUM_ARCS See Also: shouldExpandNodeWithFixedLengthArcs
	FIXED_LENGTH_ARC_SHALLOW_NUM_ARCS = 5

	// FIXED_LENGTH_ARC_DEEP_NUM_ARCS See Also: shouldExpandNodeWithFixedLengthArcs
	FIXED_LENGTH_ARC_DEEP_NUM_ARCS = 10

	// DIRECT_ADDRESSING_MAX_OVERSIZE_WITH_CREDIT_FACTOR Maximum oversizing factor allowed for direct
	// addressing compared to binary search when expansion credits allow the oversizing. This factor
	// prevents expansions that are obviously too costly even if there are sufficient credits.
	// See Also: shouldExpandNodeWithDirectAddressing
	DIRECT_ADDRESSING_MAX_OVERSIZE_WITH_CREDIT_FACTOR = 1.66

	// FILE_FORMAT_NAME Increment version to change it
	FILE_FORMAT_NAME = "FST"

	VERSION_START = 6

	VERSION_CURRENT = 7

	// FINAL_END_NODE Never serialized; just used to represent the virtual
	// final node w/ no arcs:
	FINAL_END_NODE = -1

	// NON_FINAL_END_NODE Never serialized; just used to represent the virtual
	// non-final node w/ no arcs:
	NON_FINAL_END_NODE = 0

	// END_LABEL If arc has this label then that arc is final/accepted
	END_LABEL = -1

	DEFAULT_MAX_BLOCK_BITS = 30
)

// FST Represents an finite state machine (FST), using a compact byte[] format.
// The format is similar to what's used by Morfologik (https://github.com/morfologik/morfologik-stemming).
// See the package documentation for some simple examples.
type FST struct {
	inputType InputType

	// if non-null, this FST accepts the empty string and
	// produces this output
	emptyOutput any

	// A BytesStore, used during building, or during reading when the FST is very large (more than 1 GB). If the FST is less than 1 GB then bytesArray is set instead.
	bytes BytesStore

	fstStore  FSTStore
	startNode int64
	outputs   Outputs
}

// InputType Specifies allowed range of each int input label for this FST.
type InputType int

const (
	BYTE1 = InputType(iota)
	BYTE2
	BYTE4
)

// Arc Represents a single arc.
type Arc struct {
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

	//*** Fields for a direct addressing node. nodeFlags == ARCS_FOR_DIRECT_ADDRESSING.

	//Start position in the FST.BytesReader of the presence bits for a direct addressing node, aka the bit-table
	bitTableStart int64

	firstLabel int

	// Index of the current label of a direct addressing node. While arcIdx is the current index in the label range, presenceIndex is its corresponding index in the list of actually present labels. It is equal to the number of bits set before the bit at arcIdx in the bit-table. This field is a cache to avoid to count bits set repeatedly when iterating the next arcs.
	presenceIndex int
}

type BitTable struct {
}

// BytesReader Reads bytes stored in an FST.
type BytesReader interface {
	store.DataOutput

	// GetPosition Get current read position.
	GetPosition() int64

	// SetPosition Set current read position.
	SetPosition(pos int64) error

	// Reversed Returns true if this reader uses reversed bytes under-the-hood.
	Reversed() bool
}
