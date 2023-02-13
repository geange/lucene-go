package bkd

import (
	"bytes"
	"github.com/bits-and-blooms/bitset"
	"github.com/geange/lucene-go/core/store"
)

const (
	CODEC_NAME                     = "BKD"
	VERSION_START                  = 4
	VERSION_LEAF_STORES_BOUNDS     = 5
	VERSION_SELECTIVE_INDEXING     = 6
	VERSION_LOW_CARDINALITY_LEAVES = 7
	VERSION_META_FILE              = 9
	VERSION_CURRENT                = VERSION_META_FILE

	// SPLITS_BEFORE_EXACT_BOUNDS Number of splits before we compute the exact bounding box of an inner node.
	SPLITS_BEFORE_EXACT_BOUNDS = 4

	// DEFAULT_MAX_MB_SORT_IN_HEAP Default maximum heap to use, before spilling to (slower) disk
	DEFAULT_MAX_MB_SORT_IN_HEAP = 16.0
)

// TODO
//   - allow variable length byte[] (across docs and dims), but this is quite a bit more hairy
//   - we could also index "auto-prefix terms" here, and use better compression, and maybe only use for the "fully contained" case so we'd
//     only index docIDs
//   - the index could be efficiently encoded as an FST, so we don't have wasteful
//     (monotonic) long[] leafBlockFPs; or we could use MonotonicLongValues ... but then
//     the index is already plenty small: 60M OSM points --> 1.1 MB with 128 points
//     per leaf, and you can reduce that by putting more points per leaf
//   - we could use threads while building; the higher nodes are very parallelizable

// BKDWriter Recursively builds a block KD-tree to assign all incoming points in N-dim space to smaller
// and smaller N-dim rectangles (cells) until the number of points in a given rectangle
// is <= config.maxPointsInLeafNode. The tree is partially balanced, which means the leaf nodes will
// have the requested config.maxPointsInLeafNode except one that might have less. Leaf nodes may
// straddle the two bottom levels of the binary tree. Values that fall exactly on a cell boundary
// may be in either cell.
// The number of dimensions can be 1 to 8, but every byte[] value is fixed length.
// This consumes heap during writing: it allocates a Long[numLeaves], a byte[numLeaves*(1+config.bytesPerDim)]
// and then uses up to the specified maxMBSortInHeap heap space for writing.
// NOTE: This can write at most Integer.MAX_VALUE * config.maxPointsInLeafNode / config.bytesPerDim total points.
// lucene.experimental
type BKDWriter struct {
	config              *BKDConfig
	tempDir             *store.TrackingDirectoryWrapper
	tempFileNamePrefix  string
	maxMBSortInHeap     float64
	scratchDiff         []byte
	scratch1            []byte
	scratch2            []byte
	scratchBytesRef1    *bytes.Buffer
	scratchBytesRef2    *bytes.Buffer
	commonPrefixLengths []int
	docsSeen            *bitset.BitSet
	pointWriter         PointWriter
	finished            bool
	tempInput           store.IndexOutput
	maxPointsSortInHeap int
	minPackedValue      []byte
	maxPackedValue      []byte
	pointCount          int64
	totalPointCount     int64
	maxDoc              int
}
