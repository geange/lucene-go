package bkd

import "fmt"

// BKDConfig Basic parameters for indexing points on the BKD tree.
type BKDConfig struct {
	// How many dimensions we are storing at the leaf (data) nodes
	NumDims int

	// How many dimensions we are indexing in the internal nodes
	NumIndexDims int

	// How many bytes each value in each dimension takes.
	BytesPerDim int

	// max points allowed on a Leaf block
	MaxPointsInLeafNode int

	// numDataDims * bytesPerDim
	PackedBytesLength int

	// numIndexDims * bytesPerDim
	PackedIndexBytesLength int

	//packedBytesLength plus docID size
	BytesPerDoc int
}

func NewBKDConfig(numDims, numIndexDims, bytesPerDim, maxPointsInLeafNode int) (*BKDConfig, error) {
	err := verifyParams(numDims, numIndexDims, bytesPerDim, maxPointsInLeafNode)
	if err != nil {
		return nil, err
	}

	config := &BKDConfig{}
	config.NumDims = numDims
	config.NumIndexDims = numIndexDims
	config.BytesPerDim = bytesPerDim
	config.MaxPointsInLeafNode = maxPointsInLeafNode
	config.PackedIndexBytesLength = numIndexDims * bytesPerDim
	config.PackedBytesLength = numDims * bytesPerDim
	// dimensional values (numDims * bytesPerDim) + docID (int)
	config.BytesPerDoc = config.PackedBytesLength + 4
	return config, nil
}

func verifyParams(numDims, numIndexDims, bytesPerDim, maxPointsInLeafNode int) error {
	if numDims < 1 || numDims > MAX_DIMS {
		return fmt.Errorf("numDims must be 1 .. %d (got: %d)", MAX_DIMS, numDims)
	}

	if numIndexDims < 1 || numIndexDims > MAX_INDEX_DIMS {
		return fmt.Errorf("numIndexDims must be 1 .. %d (got: %d)", MAX_INDEX_DIMS, numIndexDims)
	}

	if numIndexDims > numDims {
		return fmt.Errorf("numIndexDims cannot exceed numDims (%d) (got: %d)", numDims, numIndexDims)
	}

	if bytesPerDim <= 0 {
		return fmt.Errorf("bytesPerDim must be > 0; got %d", bytesPerDim)
	}

	if maxPointsInLeafNode <= 0 {
		return fmt.Errorf("maxPointsInLeafNode must be > 0; got %d", maxPointsInLeafNode)
	}

	//maxPointsInLeafNode > ArrayUtil.MAX_ARRAY_LENGTH
	return nil
}

const (
	// DEFAULT_MAX_POINTS_IN_LEAF_NODE Default maximum number of point in each leaf block
	DEFAULT_MAX_POINTS_IN_LEAF_NODE = 512

	// MAX_DIMS Maximum number of index dimensions (2 * max index dimensions)
	MAX_DIMS = 16

	// MAX_INDEX_DIMS Maximum number of index dimensions
	MAX_INDEX_DIMS = 8
)
