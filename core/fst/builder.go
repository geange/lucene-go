package fst

import (
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util"
)

const (
	// DIRECT_ADDRESSING_MAX_OVERSIZING_FACTOR Default oversizing factor used to decide whether to encode a node with direct addressing or binary search. Default is 1: ensure no oversizing on average.
	// This factor does not determine whether to encode a node with a list of variable length arcs or with fixed length arcs. It only determines the effective encoding of a node that is already known to be encoded with fixed length arcs. See FST.shouldExpandNodeWithFixedLengthArcs() and FST.shouldExpandNodeWithDirectAddressing().
	// For English words we measured 217K nodes, only 3.27% nodes are encoded with fixed length arcs, and 99.99% of them with direct addressing. Overall FST memory reduced by 1.67%.
	// For worst case we measured 168K nodes, 50% of them are encoded with fixed length arcs, and 14% of them with direct encoding. Overall FST memory reduced by 0.8%.
	// Use TestFstDirectAddressing.main() and TestFstDirectAddressing.testWorstCaseForDirectAddressing() to evaluate a change.
	// See Also: setDirectAddressingMaxOversizingFactor
	DIRECT_ADDRESSING_MAX_OVERSIZING_FACTOR = 1.0
)

// Builder Builds a minimal FST (maps an IntsRef term to an arbitrary output) from pre-sorted terms with outputs. The FST becomes an FSA if you use NoOutputs. The FST is written on-the-fly into a compact serialized format byte array, which can be saved to / loaded from a Directory or used directly for traversal. The FST is always finite (no cycles).
// NOTE: The algorithm is described at http://citeseerx.ist.psu.edu/viewdoc/summary?doi=10.1.1.24.3698
// The parameterized type T is the output type. See the subclasses of Outputs.
// FSTs larger than 2.1GB are now possible (as of Lucene 4.2). FSTs containing more than 2.1B nodes are also now possible, however they cannot be packed.
// lucene.experimental
type Builder[T any] struct {
	dedupHash *NodeHash[T]
	fst       *FST[T]
	NO_OUTPUT T

	// private static final boolean DEBUG = true;
	// simplistic pruning: we prune node (and all following
	// nodes) if less than this number of terms go through it:
	minSuffixCount1 int

	// better pruning: we prune node (and all following
	// nodes) if the prior node has less than this number of
	// terms go through it:
	minSuffixCount2 int

	doShareNonSingletonNodes bool

	shareMaxTailLength int

	// NOTE: cutting this over to ArrayList instead loses ~6%
	// in build performance on 9.8M Wikipedia terms; so we
	// left this as an array:
	// current "frontier"
	frontier []UnCompiledNode[T]

	// Used for the BIT_TARGET_NEXT optimization (whereby
	// instead of storing the address of the target node for
	// a given arc, we mark a single bit noting that the next
	// node in the byte[] is the target node):
	lastFrozenNode int64

	numBytesPerArc []int

	numLabelBytesPerArc []int

	fixedLengthArcsBuffer *FixedLengthArcsBuffer

	arcCount                  int64
	nodeCount                 int64
	binarySearchNodeCount     int64
	directAddressingNodeCount int64

	allowFixedLengthArcs                bool
	directAddressingMaxOversizingFactor float64
	directAddressingExpansionCredit     int64

	bytes *BytesStore
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
	arcs       []BuilderArc[T]
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

// FixedLengthArcsBuffer Reusable buffer for building nodes with fixed length arcs (binary search or direct addressing).
type FixedLengthArcsBuffer struct {
	// Initial capacity is the max length required for the header of a node with fixed length arcs:
	// header(byte) + numArcs(vint) + numBytes(vint)
	bytes []byte
	bado  store.ByteArrayDataOutput
}

func (f *FixedLengthArcsBuffer) resetPosition() *FixedLengthArcsBuffer {
	f.bado.Reset(f.bytes)
	return f
}

func (f *FixedLengthArcsBuffer) writeByte(b byte) *FixedLengthArcsBuffer {
	f.bado.WriteByte(b)
	return f
}

func (f *FixedLengthArcsBuffer) writeVInt(i int) *FixedLengthArcsBuffer {
	f.bado.WriteUvarint(uint64(i))
	return f
}

func (f *FixedLengthArcsBuffer) getPosition() int {
	return f.bado.GetPosition()
}

func (f *FixedLengthArcsBuffer) getBytes() []byte {
	return f.bytes
}

func (f *FixedLengthArcsBuffer) ensureCapacity(capacity int) *FixedLengthArcsBuffer {
	if len(f.bytes) < capacity {
		f.bytes = make([]byte, util.Oversize(capacity, 8))
		f.bado.Reset(f.bytes)
	}
	return f
}
