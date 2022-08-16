package fst

import (
	"errors"
	"github.com/geange/lucene-go/core/store"
)

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
type FST[T any] struct {
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

func (f *FST[T]) finish(newStartNode int64) error {
	if f.startNode != -1 {
		return errors.New("already finished")
	}
	if newStartNode == int64(FINAL_END_NODE) && f.emptyOutput != nil {
		newStartNode = 0
	}
	f.startNode = newStartNode
	f.bytes.Finish()
	return nil
}

func (f *FST[T]) GetEmptyOutput() any {
	return f.emptyOutput
}

func (f *FST[T]) setEmptyOutput(v any) {
	if f.emptyOutput != nil {
		f.emptyOutput = f.outputs.Merge(f.emptyOutput, v)
	} else {
		f.emptyOutput = v
	}
}

func (f *FST[T]) Save(metaOut, out store.DataOutput) error {
	panic("")
}

// SaveToFile Writes an automaton to a file.
func (f *FST[T]) SaveToFile(path string) error {
	panic("")
}

// Read Reads an automaton from a file.
func Read(path string, outputs Outputs) error {
	panic("")
}

func (f *FST[T]) writeLabel(out store.DataOutput, v int) error {
	panic("")
}

// Reads one BYTE1/2/4 label from the provided DataInput.
func (f *FST[T]) readLabel(in store.DataInput) error {
	panic("")
}

// returns true if the node at this address has any outgoing arcs
func targetHasArcs[T any](arc *Arc[T]) bool {
	return arc.Target() > 0
}

// serializes new node by appending its bytes to the end
// of the current byte[]
func (f *FST[T]) addNode(builder *Builder[T], nodeIn *UnCompiledNode[T]) error {
	panic("")
}

// Returns whether the given node should be expanded with fixed length arcs. Nodes will be expanded depending on their depth (distance from the root node) and their number of arcs.
// Nodes with fixed length arcs use more space, because they encode all arcs with a fixed number of bytes, but they allow either binary search or direct addressing on the arcs (instead of linear scan) on lookup by arc label.
func (f *FST[T]) shouldExpandNodeWithFixedLengthArcs(builder *Builder[T], nodeIn *UnCompiledNode[T]) error {
	panic("")
}

func flag(flags, bits int) bool {
	return flags&bits != 0
}

// InputType Specifies allowed range of each int input label for this FST.
type InputType int

const (
	BYTE1 = InputType(iota)
	BYTE2
	BYTE4
)

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
