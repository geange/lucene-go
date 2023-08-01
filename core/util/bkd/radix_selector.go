package bkd

import (
	"github.com/geange/lucene-go/core/store"
	"sort"
)

const (
	HISTOGRAM_SIZE          = 256      // size of the histogram
	MAX_SIZE_OFFLINE_BUFFER = 1024 * 8 // size of the online buffer: 8 KB

)

// RadixSelector Offline Radix selector for BKD tree.
// lucene.internal
type RadixSelector struct {
	histogram           []int64         // histogram array
	bytesSorted         int             // number of bytes to be sorted: config.bytesPerDim + Integer.BYTES
	maxPointsSortInHeap int             // flag to when we are moving to sort on heap
	offlineBuffer       []byte          // reusable buffer
	partitionBucket     []int           // holder for sortPartition points
	scratch             []byte          // scratch array to hold temporary data
	tempDir             store.Directory // Directory to create new Offline writer
	tempFileNamePrefix  string          // prefix for temp files
	config              *Config         // BKD tree configuration
}

func NewRadixSelector(config *Config, maxPointsSortInHeap int,
	tempDir store.Directory, tempFileNamePrefix string) *RadixSelector {

	INTEGER_BYTES := 4

	numberOfPointsOffline := MAX_SIZE_OFFLINE_BUFFER / config.BytesPerDoc()

	bytesSorted := config.BytesPerDim() + (config.NumDims()-config.NumIndexDims())*config.BytesPerDim() + INTEGER_BYTES

	return &RadixSelector{
		histogram:           make([]int64, HISTOGRAM_SIZE),
		bytesSorted:         bytesSorted,
		maxPointsSortInHeap: maxPointsSortInHeap,
		offlineBuffer:       make([]byte, numberOfPointsOffline*config.BytesPerDoc()),
		partitionBucket:     make([]int, bytesSorted),
		scratch:             make([]byte, bytesSorted),
		tempDir:             tempDir,
		tempFileNamePrefix:  tempFileNamePrefix,
		config:              config,
	}
}

// Select It uses the provided points from the given from to the given to to populate the
// partitionSlices array holder (length > 1) with two path slices so the path slice at
// position 0 contains sortPartition - from points where the value of the dim is lower or equal
// to the to -from points on the slice at position 1. The dimCommonPrefix provides a hint
// for the length of the common prefix length for the dim where are partitioning the points.
// It return the value of the dim at the sortPartition point. If the provided points is wrapping
// an OfflinePointWriter, the writer is destroyed in the process to save disk space.
func (b *RadixSelector) Select(points *PathSlice, partitionSlices []*PathSlice,
	from, to, partitionPoint int64, dim, dimCommonPrefix int) ([]byte, error) {

	b.checkArgs(from, to, partitionPoint)

	//assert partitionSlices.length > 1 : "[sortPartition alices] must be > 1, got " + partitionSlices.length;

	if writer, ok := points.writer.(*HeapPointWriter); ok {
		partition := b.heapRadixSelect(writer, dim, int(from), int(to), int(partitionPoint), dimCommonPrefix)
		partitionSlices[0] = NewPathSlice(points.writer, from, partitionPoint-from)
		partitionSlices[1] = NewPathSlice(points.writer, partitionPoint, to-partitionPoint)
		return partition, nil
	}

	panic("unsupported pointWriter")

	//offlinePointWriter := points.writer.(*OfflinePointWriter)
	//
	//left := b.getPointWriter(partitionPoint-from, "left"+strconv.Itoa(dim))
	//right := b.getPointWriter(to-partitionPoint, "right"+strconv.Itoa(dim))
	//partitionSlices[0] = NewPathSlice(left, 0, partitionPoint-from)
	//partitionSlices[1] = NewPathSlice(right, 0, to-partitionPoint)
	//sortPartition := b.buildHistogramAndPartition(offlinePointWriter, left, right, from, to, partitionPoint, 0, dimCommonPrefix, dim)
	//return sortPartition, nil
}

func (b *RadixSelector) getPointWriter(count int64, desc string) PointWriter {
	// As we recurse, we hold two on-heap point writers at any point. Therefore the
	// max size for these objects is half of the total points we can have on-heap.
	if int(count) <= b.maxPointsSortInHeap/2 {
		return NewHeapPointWriter(b.config, int(count))
	}
	return NewOfflinePointWriter(b.config, b.tempDir, b.tempFileNamePrefix, desc, count)
}

func (b *RadixSelector) checkArgs(from, to, partitionPoint int64) {
	if partitionPoint < from {
		panic("partitionPoint must be >= from")
	}
	if partitionPoint >= to {
		panic("partitionPoint must be < to")
	}
}

func (b *RadixSelector) findCommonPrefixAndHistogram(points *OfflinePointWriter, from, to int64, dim, dimCommonPrefix int) int {
	panic("")
}

func (b *RadixSelector) buildHistogramAndPartition(points *OfflinePointWriter, left, right PointWriter,
	from, to, partitionPoint int64, iteration, baseCommonPrefix, dim int) []byte {
	panic("")
}

func (b *RadixSelector) heapRadixSelect(points *HeapPointWriter,
	dim, from, to, partitionPoint, commonPrefixLength int) []byte {

	dimOffset := dim*b.config.BytesPerDim() + commonPrefixLength
	dimCmpBytes := b.config.BytesPerDim() - commonPrefixLength
	dataOffset := b.config.packedIndexBytesLength - dimCmpBytes

	sorter := &heapRadixSort{
		from:        from,
		to:          to,
		dimOffset:   dimOffset,
		dimCmpBytes: dimCmpBytes,
		dataOffset:  dataOffset,
		selector:    b,
		points:      points,
	}
	SortK(sorter, partitionPoint)

	partition := make([]byte, b.config.BytesPerDim())
	pointValue := points.GetPackedValueSlice(partitionPoint)
	packedValue := pointValue.PackedValue()
	copy(partition, b.getDimValues(packedValue, dim))
	return partition
}

func (b *RadixSelector) getDimValues(bs []byte, dim int) []byte {
	from := dim * b.config.BytesPerDim()
	to := from + b.config.BytesPerDim()
	return bs[from:to]
}

func (b *RadixSelector) HeapRadixSort(points *HeapPointWriter, from, to, dim, commonPrefixLength int) {
	dimOffset := dim*b.config.BytesPerDim() + commonPrefixLength
	dimCmpBytes := b.config.BytesPerDim() - commonPrefixLength
	dataOffset := b.config.packedIndexBytesLength - dimCmpBytes

	sorter := &heapRadixSort{
		from:        from,
		to:          to,
		dimOffset:   dimOffset,
		dimCmpBytes: dimCmpBytes,
		dataOffset:  dataOffset,
		selector:    b,
		points:      points,
	}
	sort.Sort(sorter)
}

// PathSlice Sliced reference to points in an PointWriter.
type PathSlice struct {
	writer PointWriter
	start  int64
	count  int64
}

func NewPathSlice(writer PointWriter, start, count int64) *PathSlice {
	return &PathSlice{
		writer: writer,
		start:  start,
		count:  count,
	}
}

func (p *PathSlice) PointWriter() PointWriter {
	return p.writer
}

func (p *PathSlice) Start() int64 {
	return p.start
}

func (p *PathSlice) Count() int64 {
	return p.count
}
