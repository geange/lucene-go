package fst

import (
	"github.com/geange/lucene-go/core/store"
)

// Builder Builds a minimal FST (maps an IntsRef term to an arbitrary output) from pre-sorted terms with outputs.
// The FST becomes an FSA if you use NoOutputs. The FST is written on-the-fly into a compact serialized format
// byte array, which can be saved to / loaded from a Directory or used directly for traversal.
// The FST is always finite (no cycles).
//
// NOTE: The algorithm is described at http://citeseerx.ist.psu.edu/viewdoc/summary?doi=10.1.1.24.3698
//
// The parameterized type T is the output type. See the subclasses of Outputs.
//
// FSTs larger than 2.1GB are now possible (as of Lucene 4.2). FSTs containing more than 2.1B nodes are also
// now possible, however they cannot be packed.
//
// lucene.experimental
type Builder struct {
	dedupHash *NodeHash
	fst       *FST
	NO_OUTPUT any

	// private static final boolean DEBUG = true;

	// simplistic pruning: we prune node (and all following
	// nodes) if less than this number of terms go through it:
	minSuffixCount1 int

	// better pruning: we prune node (and all following
	// nodes) if the prior node has less than this number of
	// terms go through it:
	minSuffixCount2 int

	doShareNonSingletonNodes bool
	shareMaxTailLength       int

	// NOTE: cutting this over to ArrayList instead loses ~6%
	// in build performance on 9.8M Wikipedia terms; so we
	// left this as an array:
	// current "frontier"
	frontier []UnCompiledNode

	// Used for the BIT_TARGET_NEXT optimization (whereby
	// instead of storing the address of the target node for
	// a given arc, we mark a single bit noting that the next
	// node in the byte[] is the target node):
	lastFrozenNode int64

	// Reused temporarily while building the FST:
	numBytesPerArc      []int
	numLabelBytesPerArc []int64

	fixedLengthArcsBuffer *FixedLengthArcsBuffer

	arcCount                            int64
	nodeCount                           int64
	binarySearchNodeCount               int64
	directAddressingNodeCount           int64
	allowFixedLengthArcs                bool
	directAddressingMaxOversizingFactor float64
	directAddressingExpansionCredit     int64
	bytes                               *ByteStore
}

func (b *Builder) GetDirectAddressingMaxOversizingFactor() float64 {
	return b.directAddressingMaxOversizingFactor
}

// DIRECT_ADDRESSING_MAX_OVERSIZING_FACTOR Default oversizing factor used to decide whether to encode a node with direct addressing or binary search. Default is 1: ensure no oversizing on average.
// This factor does not determine whether to encode a node with a list of variable length arcs or with fixed length arcs. It only determines the effective encoding of a node that is already known to be encoded with fixed length arcs. See FST.shouldExpandNodeWithFixedLengthArcs() and FST.shouldExpandNodeWithDirectAddressing().
// For English words we measured 217K nodes, only 3.27% nodes are encoded with fixed length arcs, and 99.99% of them with direct addressing. Overall FST memory reduced by 1.67%.
// For worst case we measured 168K nodes, 50% of them are encoded with fixed length arcs, and 14% of them with direct encoding. Overall FST memory reduced by 0.8%.
// Use TestFstDirectAddressing.main() and TestFstDirectAddressing.testWorstCaseForDirectAddressing() to evaluate a change.
// see: setDirectAddressingMaxOversizingFactor
const DIRECT_ADDRESSING_MAX_OVERSIZING_FACTOR = 1.0

// FixedLengthArcsBuffer Reusable buffer for building nodes with fixed length arcs (binary search or direct addressing).
type FixedLengthArcsBuffer struct {
	bytes []byte

	bado *store.ByteArrayDataOutput
}

func NewFixedLengthArcsBuffer() *FixedLengthArcsBuffer {
	bytes := make([]byte, 11)
	return &FixedLengthArcsBuffer{
		bytes: bytes,
		bado:  store.NewByteArrayDataOutput(bytes),
	}
}

// Ensures the capacity of the internal byte array. Enlarges it if needed.
func (f *FixedLengthArcsBuffer) ensureCapacity(capacity int) error {
	if len(f.bytes) < capacity {
		f.bytes = make([]byte, capacity)
		return f.bado.Reset(f.bytes)
	}
	return nil
}

func (f *FixedLengthArcsBuffer) resetPosition() error {
	return f.bado.Reset(f.bytes)
}

func (f *FixedLengthArcsBuffer) writeByte(b byte) error {
	return f.bado.WriteByte(b)
}

func (f *FixedLengthArcsBuffer) writeVInt(i int64) error {
	return f.bado.WriteUvarint(uint64(i))
}

func (f *FixedLengthArcsBuffer) getPosition() int64 {
	return int64(f.bado.GetPosition())
}

func (f *FixedLengthArcsBuffer) GetBytes() []byte {
	return f.bytes
}
