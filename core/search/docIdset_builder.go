package search

import (
	"github.com/bits-and-blooms/bitset"
	"github.com/geange/lucene-go/core/index"
)

// DocIdSetBuilder
// A builder of DocIdSets. At first it uses a sparse structure to gather documents,
// and then upgrades to a non-sparse bit set once enough hits match. To add documents,
// you first need to call grow in order to reserve space, and then call
// DocIdSetBuilder.BulkAdder.add(int) on the returned DocIdSetBuilder.BulkAdder.
// lucene.internal
type DocIdSetBuilder struct {
	maxDoc    int
	threshold int
	// pkg-private for testing
	multivalued     bool
	numValuesPerDoc float64
	buffers         []*Buffer
	// accumulated size of the allocated buffers
	totalAllocated int
	bitSet         FixedBitSetAdder
	counter        int64
	adder          BulkAdder
}

// NewDocIdSetBuilder Create a builder that can contain doc IDs between 0 and maxDoc.
func NewDocIdSetBuilder(maxDoc int) *DocIdSetBuilder {
	return newDocIdSetBuilder(maxDoc, -1, -1)
}

// NewDocIdSetBuilderV1
// Create a DocIdSetBuilder instance that is optimized for accumulating docs that match the given Terms.
func NewDocIdSetBuilderV1(maxDoc int, terms index.Terms) (*DocIdSetBuilder, error) {
	docCount, err := terms.GetDocCount()
	if err != nil {
		return nil, err
	}

	sumDocFreq, err := terms.GetSumDocFreq()
	if err != nil {
		return nil, err
	}
	return newDocIdSetBuilder(maxDoc, docCount, sumDocFreq), nil
}

// NewDocIdSetBuilderV2
// Create a DocIdSetBuilder instance that is optimized for accumulating docs that match the given PointValues.
func NewDocIdSetBuilderV2(maxDoc int, values index.PointValues, field string) *DocIdSetBuilder {
	return newDocIdSetBuilder(maxDoc, values.GetDocCount(), values.Size())
}

func newDocIdSetBuilder(maxDoc, docCount int, valueCount int64) *DocIdSetBuilder {
	builder := &DocIdSetBuilder{
		maxDoc:      maxDoc,
		multivalued: docCount < 0 || int64(docCount) != valueCount,
	}

	if docCount <= 0 || valueCount < 0 {
		// assume one value per doc, this means the cost will be overestimated
		// if the docs are actually multi-valued
		builder.numValuesPerDoc = 1
	} else {
		// otherwise compute from index stats
		builder.numValuesPerDoc = float64(valueCount) / float64(docCount)
	}

	// For ridiculously small sets, we'll just use a sorted int[]
	// maxDoc >>> 7 is a good value if you want to save memory, lower values
	// such as maxDoc >>> 11 should provide faster building but at the expense
	// of using a full bitset even for quite sparse data
	builder.threshold = maxDoc >> 7

	return builder
}

// Add the content of the provided DocIdSetIterator to this builder.
// NOTE: if you need to build a DocIdSet out of a single DocIdSetIterator,
// you should rather use RoaringDocIdSet.Builder.
func (d *DocIdSetBuilder) Add(iter index.DocIdSetIterator) error {
	panic("")
}

// Grow
// Reserve space and return a DocIdSetBuilder.BulkAdder object
// that can be used to add up to numDocs documents.
func (d *DocIdSetBuilder) Grow(numDocs int) BulkAdder {
	panic("")
}

// Build a DocIdSet from the accumulated doc IDs.
func (d *DocIdSetBuilder) build() DocIdSet {
	panic("")
}

// BulkAdder Utility class to efficiently add many docs in one go.
// See Also: grow
type BulkAdder interface {
	Add(doc int)
}

type FixedBitSetAdder struct {
	bitSet *bitset.BitSet
}

func NewFixedBitSetAdder(bitSet *bitset.BitSet) *FixedBitSetAdder {
	return &FixedBitSetAdder{bitSet: bitSet}
}

func (f *FixedBitSetAdder) Add(doc int) {
	f.bitSet.Set(uint(doc))
}

type Buffer struct {
	array  []int
	length int
}

func NewBuffer(length int) *Buffer {
	return &Buffer{
		array:  make([]int, length),
		length: 0,
	}
}

func NewBufferV1(array []int, length int) *Buffer {
	return &Buffer{array: array, length: length}
}

var _ BulkAdder = &BufferAdder{}

type BufferAdder struct {
	buffer Buffer
}

func NewBufferAdder(buffer Buffer) *BufferAdder {
	return &BufferAdder{buffer: buffer}
}

func (b *BufferAdder) Add(doc int) {
	b.buffer.array[b.buffer.length] = doc
	b.buffer.length++
}
