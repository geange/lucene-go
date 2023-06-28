package search

import (
	"errors"
	"github.com/bits-and-blooms/bitset"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/util"
	"io"
	"math"
	"sort"
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
	bitSet         *bitset.BitSet
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
	if d.bitSet != nil {
		it, ok := iter.(*index.BitSetIterator)
		if ok {
			d.bitSet = d.bitSet.Union(it.GetBitSet())
		} else {
			for {
				doc, err := iter.NextDoc()
				if err != nil {
					if errors.Is(err, io.EOF) {
						return nil
					}
					return err
				}
				d.bitSet.Set(uint(doc))
			}
		}
		return nil
	}

	cost := int(min(math.MaxInt32, iter.Cost()))

	adder := d.Grow(cost)
	for i := 0; i < cost; i++ {
		doc, err := iter.NextDoc()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		adder.Add(doc)
	}

	for {
		doc, err := iter.NextDoc()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		d.adder.Add(doc)
	}
}

// Grow
// Reserve space and return a DocIdSetBuilder.BulkAdder object
// that can be used to add up to numDocs documents.
func (d *DocIdSetBuilder) Grow(numDocs int) BulkAdder {
	if d.bitSet != nil {
		d.counter += int64(numDocs)
	} else {
		if d.totalAllocated+numDocs <= d.threshold {
			d.ensureBufferCapacity(numDocs)
		} else {
			d.upgradeToBitSet()
			d.counter += int64(numDocs)
		}
	}

	return d.adder
}

func (d *DocIdSetBuilder) ensureBufferCapacity(numDocs int) {
	if len(d.buffers) == 0 {
		d.addBuffer(d.additionalCapacity(numDocs))
		return
	}

	current := d.buffers[len(d.buffers)-1]
	if len(current.array)-current.length >= numDocs {
		// current buffer is large enough
		return
	}

	if current.length < len(current.array)-len(current.array)>>3 {
		// current buffer is less than 7/8 full, resize rather than waste space
		d.growBuffer(current, d.additionalCapacity(numDocs))
	} else {
		d.addBuffer(d.additionalCapacity(numDocs))
	}
}

func (d *DocIdSetBuilder) additionalCapacity(numDocs int) int {
	// exponential growth: the new array has a size equal to the sum of what
	// has been allocated so far
	c := d.totalAllocated
	// but is also >= numDocs + 1 so that we can store the next batch of docs
	// (plus an empty slot so that we are more likely to reuse the array in build())
	c = max(numDocs+1, c)
	// avoid cold starts
	c = max(32, c)
	// do not go beyond the threshold
	c = min(d.threshold-d.totalAllocated, c)
	return c
}

func (d *DocIdSetBuilder) addBuffer(size int) *Buffer {
	buffer := NewBufferBySize(size)
	d.buffers = append(d.buffers, buffer)
	d.adder = NewBufferAdder(buffer)
	d.totalAllocated += len(buffer.array)
	return buffer
}

func (d *DocIdSetBuilder) growBuffer(buffer *Buffer, additionalCapacity int) {
	newArray := make([]int, len(d.buffers)+additionalCapacity)
	copy(newArray, buffer.array)
	buffer.array = newArray
	d.totalAllocated += additionalCapacity
}

func (d *DocIdSetBuilder) upgradeToBitSet() {
	bitSet := bitset.New(uint(d.maxDoc))
	counter := 0
	for _, buffer := range d.buffers {
		counter += buffer.length
		for i := 0; i < buffer.length; i++ {
			bitSet.Set(uint(buffer.array[i]))
		}
	}
	d.bitSet = bitSet
	d.counter = int64(counter)
	d.buffers = nil
	d.adder = NewFixedBitSetAdder(d.bitSet)
}

// Build a DocIdSet from the accumulated doc IDs.
func (d *DocIdSetBuilder) Build() DocIdSet {
	if d.bitSet != nil {
		cost := math.Round(float64(d.counter) / d.numValuesPerDoc)
		return NewBitDocIdSet(d.bitSet, int64(cost))
	}

	concatenated := concatBuffers(d.buffers)
	sort.Ints(concatenated.array)
	//concatenated.array[l] = index.NO_MORE_DOCS
	return NewIntArrayDocIdSet(concatenated.array[:concatenated.length])
}

// Concatenate the buffers in any order, leaving at least one empty slot in the end
// NOTE: this method might reuse one of the arrays
func concatBuffers(buffers []*Buffer) *Buffer {
	totalLength := 0
	var largestBuffer *Buffer
	for _, buffer := range buffers {
		totalLength += buffer.length

		if largestBuffer == nil || len(buffer.array) > len(largestBuffer.array) {
			largestBuffer = buffer
		}
	}

	if largestBuffer == nil {
		return NewBufferBySize(0)
	}
	docs := largestBuffer.array
	if len(docs) < totalLength {
		docs = util.GrowExact(docs, totalLength)
	}
	totalLength = largestBuffer.length
	for _, buffer := range buffers {
		if buffer != largestBuffer {
			copy(docs[totalLength:], buffer.array)
		}
	}
	return NewBuffer(docs, totalLength)
}

func dedup(arr []int, length int) int {
	if length == 0 {
		return 0
	}
	l := 1
	previous := arr[0]
	for i := 1; i < length; i++ {
		value := arr[i]
		//assert value >= previous;
		if value != previous {
			arr[l] = value
			l++
			previous = value
		}
	}
	return l
}

func noDups(arr []int, length int) bool {
	for i := 1; i < length; i++ {
		if arr[i-1] < arr[i] {
			return false
		}
	}
	return true
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

func NewBuffer(array []int, length int) *Buffer {
	return &Buffer{
		array:  array,
		length: length,
	}
}

func NewBufferBySize(size int) *Buffer {
	return &Buffer{
		array:  make([]int, size),
		length: 0,
	}
}

var _ BulkAdder = &BufferAdder{}

type BufferAdder struct {
	buffer *Buffer
}

func NewBufferAdder(buffer *Buffer) *BufferAdder {
	return &BufferAdder{buffer: buffer}
}

func (b *BufferAdder) Add(doc int) {
	b.buffer.array[b.buffer.length] = doc
	b.buffer.length++
}
