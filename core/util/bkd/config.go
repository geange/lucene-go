package bkd

import "fmt"

// Config Basic parameters for indexing points on the BKD tree.
type Config struct {
	// How many dimensions we are storing at the leaf (data) nodes
	// 我们在叶（数据）节点上存储了多少维度
	numDims int

	// How many dimensions we are indexing in the internal nodes
	// 当前在节点上存储的纬度数量
	numIndexDims int

	// How many bytes each value in each dimension takes.
	// 每个纬度占用的字节数量
	bytesPerDim int

	// max points allowed on a Leaf block
	// 叶子块支持的最大的点的数量
	maxPointsInLeafNode int

	// numDataDims * bytesPerDim
	packedBytesLength int

	// numIndexDims * bytesPerDim
	packedIndexBytesLength int

	//packedBytesLength plus docID size
	bytesPerDoc int
}

func NewConfig(numDims, numIndexDims, bytesPerDim, maxPointsInLeafNode int) (*Config, error) {
	err := verifyParams(numDims, numIndexDims, bytesPerDim, maxPointsInLeafNode)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	config.numDims = numDims
	config.numIndexDims = numIndexDims
	config.bytesPerDim = bytesPerDim
	config.maxPointsInLeafNode = maxPointsInLeafNode
	config.packedIndexBytesLength = numIndexDims * bytesPerDim

	packedBytesLength := numDims * bytesPerDim
	config.packedBytesLength = packedBytesLength
	// dimensional values (numDims * bytesPerDim) + docID (int)
	config.bytesPerDoc = packedBytesLength + 4
	return config, nil
}

func (c *Config) NumDims() int {
	return c.numDims
}

func (c *Config) NumIndexDims() int {
	return c.numIndexDims
}

func (c *Config) BytesPerDim() int {
	return c.bytesPerDim
}

func (c *Config) MaxPointsInLeafNode() int {
	return c.maxPointsInLeafNode
}

func (c *Config) PackedIndexBytesLength() int {
	return c.packedIndexBytesLength
}

func (c *Config) PackedBytesLength() int {
	return c.packedBytesLength
}

func (c *Config) BytesPerDoc() int {
	return c.bytesPerDoc
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
