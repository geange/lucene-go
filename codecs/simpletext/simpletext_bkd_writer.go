package simpletext

import (
	"bytes"
	"github.com/bits-and-blooms/bitset"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/bkd"
)

const (
	CODEC_NAME                    = "BKD"
	VERSION_START                 = 0
	VERSION_COMPRESSED_DOC_IDS    = 1
	VERSION_COMPRESSED_VALUES     = 2
	VERSION_IMPLICIT_SPLIT_DIM_1D = 3
	VERSION_CURRENT               = VERSION_IMPLICIT_SPLIT_DIM_1D
	DEFAULT_MAX_MB_SORT_IN_HEAP   = 16.0
)

type SimpleTextBKDWriter struct {
	// How many dimensions we are storing at the leaf (data) nodes
	config *bkd.BKDConfig

	scratch             *bytes.Buffer
	tempFileNamePrefix  string
	maxMBSortInHeap     float64
	scratchDiff         []byte
	scratch1            []byte
	scratch2            []byte
	scratchBytesRef1    []byte
	scratchBytesRef2    []byte
	commonPrefixLengths []int
	docsSeen            *bitset.BitSet
	pointWriter         index.PointsWriter
	finished            bool
	tempInput           store.IndexOutput
	maxPointsSortInHeap int

	// Minimum per-dim values, packed
	minPackedValue []byte

	// Maximum per-dim values, packed
	maxPackedValue []byte

	pointCount int64

	// An upper bound on how many points the caller will add (includes deletions)
	totalPointCount int64

	maxDoc int
}

func NewSimpleTextBKDWriter(maxDoc int, tempDir store.Directory, tempFileNamePrefix string,
	config *bkd.BKDConfig, maxMBSortInHeap float64, totalPointCount int64) *SimpleTextBKDWriter {
	panic("")
}

func (s *SimpleTextBKDWriter) Add(packedValue []byte, docID int) error {
	panic("")
}

// How many points have been added so far
func (s *SimpleTextBKDWriter) getPointCount() int64 {
	return s.pointCount
}

// Write a field from a MutablePointValues. This way of writing points is faster than regular writes with add since there is opportunity for reordering points before writing them to disk. This method does not use transient disk in order to reorder points.
func (s *SimpleTextBKDWriter) writeField(out store.IndexOutput, fieldName string, reader index.MutablePointValues) {

}
