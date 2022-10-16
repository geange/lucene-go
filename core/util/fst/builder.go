package fst

const (
	DIRECT_ADDRESSING_MAX_OVERSIZING_FACTOR = float32(1.0)
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
// lucene.experimental
type Builder struct {
	dedupHash *NodeHash
	fst       *FST
	NO_OUTPUT any

	// private static final boolean DEBUG = true;

	// simplistic pruning: we prune node (and all following nodes) if less than this number of terms go through it:
	// 简化剪枝：如果通过节点的项少于此数量，则剪枝节点（以及所有后续节点）
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
	frontier []*UnCompiledNode

	// Used for the BIT_TARGET_NEXT optimization (whereby
	// instead of storing the address of the target node for
	// a given arc, we mark a single bit noting that the next
	// node in the byte[] is the target node):
	lastFrozenNode int64

	// Reused temporarily while building the FST:
	numBytesPerArc        []int
	numLabelBytesPerArc   []int
	fixedLengthArcsBuffer *FixedLengthArcsBuffer

	arcCount                            int64
	nodeCount                           int64
	binarySearchNodeCount               int64
	directAddressingNodeCount           int64
	allowFixedLengthArcs                bool
	directAddressingMaxOversizingFactor float32
	directAddressingExpansionCredit     int64

	bytes *BytesStore
}
