package bkd

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/selector"
	"github.com/geange/lucene-go/core/util/sorter"
	"io"
	"slices"
	"strconv"
)

const (
	HISTOGRAM_SIZE          = 256      // size of the histogram
	MAX_SIZE_OFFLINE_BUFFER = 1024 * 8 // size of the online buffer: 8 KB
	INTEGER_BYTES           = 4
)

// RadixSelector Offline Radix selector for BKD tree.
// lucene.internal
type RadixSelector struct {
	histogram           []int           // histogram array
	bytesSorted         int             // number of bytes to be sorted: config.bytesPerDim + Integer.BYTES
	maxPointsSortInHeap int             // flag to when we are moving to sort on heap
	offlineBuffer       []byte          // reusable buffer
	partitionBucket     []int           // holder for sortPartition points
	scratch             []byte          // scratch array to hold temporary data
	tempDir             store.Directory // Directory to create new Offline writer
	tempFileNamePrefix  string          // prefix for temp files
	config              *Config         // BKD tree configuration
}

// NewRadixSelector Sole constructor.
func NewRadixSelector(config *Config, maxPointsSortInHeap int,
	tempDir store.Directory, tempFileNamePrefix string) *RadixSelector {

	numberOfPointsOffline := MAX_SIZE_OFFLINE_BUFFER / config.BytesPerDoc()

	bytesSorted := config.BytesPerDim() +
		(config.NumDims()-config.NumIndexDims())*config.BytesPerDim() + INTEGER_BYTES

	return &RadixSelector{
		histogram:           make([]int, HISTOGRAM_SIZE),
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

// Select
// It uses the provided points from the given from to the given to to populate the
// partitionSlices array holder (length > 1) with two path slices so the path slice at
// position 0 contains sortPartition - from points where the value of the dim is lower or equal
// to the to -from points on the slice at position 1. The dimCommonPrefix provides a hint
// for the length of the common prefix length for the dim where are partitioning the points.
// It return the value of the dim at the sortPartition point. If the provided points is wrapping
// an OfflinePointWriter, the writer is destroyed in the process to save disk space.
func (b *RadixSelector) Select(points *PathSlice, partitionSlices []*PathSlice,
	from, to, partitionPoint int, dim, dimCommonPrefix int) ([]byte, error) {

	err := b.checkArgs(from, to, partitionPoint)
	if err != nil {
		return nil, err
	}

	if writer, ok := points.writer.(*HeapPointWriter); ok {
		partition := b.heapRadixSelect(writer, dim, from, to, partitionPoint, dimCommonPrefix)
		partitionSlices[0] = NewPathSlice(points.writer, from, partitionPoint-from)
		partitionSlices[1] = NewPathSlice(points.writer, partitionPoint, to-partitionPoint)
		return partition, nil
	}

	offlinePointWriter, ok := points.writer.(*OfflinePointWriter)
	if !ok {
		return nil, errors.New("unsupported pointWriter")
	}

	left := b.getPointWriter(partitionPoint-from, "left"+strconv.Itoa(dim))
	right := b.getPointWriter(to-partitionPoint, "right"+strconv.Itoa(dim))

	partitionSlices[0] = NewPathSlice(left, 0, partitionPoint-from)
	partitionSlices[1] = NewPathSlice(right, 0, to-partitionPoint)
	return b.buildHistogramAndPartition(offlinePointWriter, left, right, from, to,
		partitionPoint, 0, dimCommonPrefix, dim)
}

func (b *RadixSelector) checkArgs(from, to, partitionPoint int) error {
	if partitionPoint < from {
		return errors.New("partitionPoint must be >= from")
	}
	if partitionPoint >= to {
		return errors.New("partitionPoint must be < to")
	}
	return nil
}

func (b *RadixSelector) findCommonPrefixAndHistogram(points *OfflinePointWriter, from, to int, dim, dimCommonPrefix int) (int, error) {
	// find common prefix
	commonPrefixPosition := b.bytesSorted
	offset := dim * b.config.bytesPerDim

	reader, err := points.getReader(nil, from, to-from, b.offlineBuffer)
	if err != nil {
		return 0, err
	}
	// assert commonPrefixPosition > dimCommonPrefix;
	_, err = reader.Next()
	if err != nil {
		return 0, err
	}
	pointValue := reader.PointValue()
	packedValueDocID := pointValue.PackedValueDocIDBytes()

	// copy dimension
	copy(b.scratch[:b.config.bytesPerDim], packedValueDocID[offset:])
	// copy data dimensions and docID
	{
		size := (b.config.numDims-b.config.numIndexDims)*b.config.bytesPerDim + INTEGER_BYTES
		from := b.config.bytesPerDim
		to := b.config.bytesPerDim + size
		copy(b.scratch[from:to], packedValueDocID[b.config.packedIndexBytesLength:])
	}

	for i := from + 1; i < to; i++ {
		_, err := reader.Next()
		if err != nil {
			return 0, err
		}
		pointValue = reader.PointValue()
		if commonPrefixPosition == dimCommonPrefix {
			b.histogram[b.getBucket(offset, commonPrefixPosition, pointValue)]++
			// we do not need to check for common prefix anymore,
			// just finish the histogram and break
			for j := i + 1; j < to; j++ {
				_, err := reader.Next()
				if err != nil {
					return 0, err
				}
				pointValue = reader.PointValue()
				b.histogram[b.getBucket(offset, commonPrefixPosition, pointValue)]++
			}
			break
		}

		// Check common prefix and adjust histogram
		startIndex := dimCommonPrefix
		if dimCommonPrefix > b.config.bytesPerDim {
			startIndex = b.config.bytesPerDim
		}

		endIndex := commonPrefixPosition
		if commonPrefixPosition > b.config.bytesPerDim {
			endIndex = b.config.bytesPerDim
		}
		packedValueDocID = pointValue.PackedValueDocIDBytes()

		j := Mismatch(b.scratch[startIndex:endIndex], packedValueDocID[offset+startIndex:offset+endIndex])
		if j == -1 {
			if commonPrefixPosition > b.config.bytesPerDim {
				startTieBreak := b.config.packedIndexBytesLength
				endTieBreak := startTieBreak + commonPrefixPosition - b.config.bytesPerDim
				k := Mismatch(b.scratch[b.config.bytesPerDim:commonPrefixPosition],
					packedValueDocID[startTieBreak:endTieBreak])
				if k != -1 {
					commonPrefixPosition = b.config.bytesPerDim + k
					for idx := range b.histogram {
						b.histogram[idx] = 0
					}
					b.histogram[b.scratch[commonPrefixPosition]] = i - from
				}
			}
		} else {
			commonPrefixPosition = dimCommonPrefix + j
			for idx := range b.histogram {
				b.histogram[idx] = 0
			}
			b.histogram[b.scratch[commonPrefixPosition]] = int(i - from)
		}

		if commonPrefixPosition != b.bytesSorted {
			b.histogram[b.getBucket(offset, commonPrefixPosition, pointValue)]++
		}
	}

	// Build partition buckets up to commonPrefix
	for i := 0; i < commonPrefixPosition; i++ {
		b.partitionBucket[i] = int(b.scratch[i])
	}
	return commonPrefixPosition, nil
}

func (b *RadixSelector) getBucket(offset, commonPrefixPosition int, pointValue PointValue) int {
	bucket := 0
	if commonPrefixPosition < b.config.bytesPerDim {
		packedValue := pointValue.PackedValue()
		bucket = int(packedValue[offset+commonPrefixPosition])
	} else {
		packedValueDocID := pointValue.PackedValueDocIDBytes()
		bucket = int(packedValueDocID[b.config.packedIndexBytesLength+commonPrefixPosition-b.config.bytesPerDim])
	}
	return bucket
}

func (b *RadixSelector) buildHistogramAndPartition(points *OfflinePointWriter, left, right PointWriter,
	from, to, partitionPoint int, iteration, baseCommonPrefix, dim int) ([]byte, error) {

	// Find common prefix from baseCommonPrefix and build histogram
	commonPrefix, err := b.findCommonPrefixAndHistogram(points, from, to, dim, baseCommonPrefix)
	if err != nil {
		return nil, err
	}

	// If all equals we just partition the points
	if commonPrefix == b.bytesSorted {
		err := b.offlinePartition(points, left, right, nil, from, to, dim, commonPrefix-1, partitionPoint)
		if err != nil {
			return nil, err
		}
		return b.partitionPointFromCommonPrefix(), nil
	}

	leftCount := 0
	rightCount := 0

	// Count left points and record the partition point
	for i := 0; i < HISTOGRAM_SIZE; i++ {
		size := b.histogram[i]
		if leftCount+size > partitionPoint-from {
			b.partitionBucket[commonPrefix] = i
			break
		}
		leftCount += size
	}

	// Count right points
	for i := b.partitionBucket[commonPrefix] + 1; i < HISTOGRAM_SIZE; i++ {
		rightCount += b.histogram[i]
	}

	delta := b.histogram[b.partitionBucket[commonPrefix]]
	//assert leftCount + rightCount + delta == to - from : (leftCount + rightCount + delta) + " / " + (to - from);

	// Special case when points are equal except last byte, we can just tie-break
	if commonPrefix == b.bytesSorted-1 {
		tieBreakCount := partitionPoint - from - leftCount
		err := b.offlinePartition(points, left, right, nil, from, to, dim, commonPrefix, tieBreakCount)
		if err != nil {
			return nil, err
		}
		return b.partitionPointFromCommonPrefix(), nil
	}

	// Create the delta points writer
	tempDeltaPoints, err := b.getDeltaPointWriter(left, right, delta, iteration)
	if err != nil {
		return nil, err
	}
	// Divide the points. This actually destroys the current writer
	err = b.offlinePartition(points, left, right, tempDeltaPoints, from, to, dim, commonPrefix, 0)
	if err != nil {
		return nil, err
	}
	deltaPoints := tempDeltaPoints

	newPartitionPoint := partitionPoint - from - leftCount

	if writer, ok := deltaPoints.(*HeapPointWriter); ok {
		return b.heapPartition(writer, left, right, dim, 0,
			deltaPoints.Count(), newPartitionPoint, commonPrefix+1)
	}

	return b.buildHistogramAndPartition(deltaPoints.(*OfflinePointWriter), left, right, 0,
		deltaPoints.Count(), newPartitionPoint, iteration+1, commonPrefix+1, dim)
}

func (b *RadixSelector) offlinePartition(points *OfflinePointWriter, left, right, deltaPoints PointWriter,
	from, to, dim, bytePosition, numDocsTiebreak int) error {

	// assert bytePosition == bytesSorted -1 || deltaPoints != null;
	offset := dim * b.config.bytesPerDim
	tiebreakCounter := 0

	reader, err := points.getReader(nil, from, to-from, b.offlineBuffer)
	if err != nil {
		return err
	}

	for {
		ok, err := reader.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		if !ok {
			break
		}

		pointValue := reader.PointValue()
		bucket := b.getBucket(offset, bytePosition, pointValue)
		if bucket < b.partitionBucket[bytePosition] {
			// to the left side
			if err := left.AppendPoint(pointValue); err != nil {
				return err
			}
		} else if bucket > b.partitionBucket[bytePosition] {
			// to the right side
			if err := right.AppendPoint(pointValue); err != nil {
				return err
			}
		} else {
			if bytePosition == b.bytesSorted-1 {
				if tiebreakCounter < numDocsTiebreak {
					if err := left.AppendPoint(pointValue); err != nil {
						return err
					}
					tiebreakCounter++
				} else {
					if err := right.AppendPoint(pointValue); err != nil {
						return err
					}
				}
			} else {
				if err := deltaPoints.AppendPoint(pointValue); err != nil {
					return err
				}
			}
		}
	}

	// Delete original file
	return points.Destroy()
}

func (b *RadixSelector) partitionPointFromCommonPrefix() []byte {
	partition := make([]byte, b.config.bytesPerDim)
	for i := 0; i < b.config.bytesPerDim; i++ {
		partition[i] = byte(b.partitionBucket[i])
	}
	return partition
}

func (b *RadixSelector) heapPartition(points *HeapPointWriter, left, right PointWriter,
	dim, from, to, partitionPoint, commonPrefix int) ([]byte, error) {

	partition := b.heapRadixSelect(points, dim, from, to, partitionPoint, commonPrefix)
	for i := from; i < to; i++ {
		point := points.GetPackedValueSlice(i)
		if i < partitionPoint {
			err := left.AppendPoint(point)
			if err != nil {
				return nil, err
			}
		} else {
			err := right.AppendPoint(point)
			if err != nil {
				return nil, err
			}
		}
	}
	return partition, nil
}

func (b *RadixSelector) heapRadixSelect(points *HeapPointWriter, dim,
	from, to, partitionPoint, commonPrefixLength int) []byte {

	dimOffset := dim*b.config.BytesPerDim() + commonPrefixLength
	dimCmpBytes := b.config.BytesPerDim() - commonPrefixLength
	// dataOffset := i*r.selector.config.BytesPerDoc() + b.config.PackedIndexBytesLength() + (k - dimCmpBytes)
	dataOffset := b.config.PackedIndexBytesLength() - dimCmpBytes

	rSelector := &radixSelector{
		dimOffset:   dimOffset,
		dimCmpBytes: dimCmpBytes,
		dataOffset:  dataOffset,
		selector:    b,
		points:      points,
	}

	selector.NewRadixSelector(rSelector, b.bytesSorted-commonPrefixLength).
		SelectK(from, to, partitionPoint)

	pointValue := points.GetPackedValueSlice(partitionPoint)
	packedValue := pointValue.PackedValue()
	return slices.Clone(b.getDimValues(packedValue, dim)[:b.config.BytesPerDim()])
}

var _ selector.RadixSelector = &radixSelector{}

type radixSelector struct {
	dimOffset   int
	dimCmpBytes int
	dataOffset  int
	selector    *RadixSelector
	points      *HeapPointWriter
}

func (r *radixSelector) Swap(i, j int) {
	r.points.Swap(i, j)
}

func (r *radixSelector) ByteAt(i, k int) int {
	if k < r.dimCmpBytes {
		// dim bytes
		idx := i*r.selector.config.BytesPerDoc() + r.dimOffset + k
		if idx >= len(r.points.block) {
			return -1
		}
		return int(r.points.block[idx])
	}

	idx := i*r.selector.config.BytesPerDoc() + r.dataOffset + k
	b := r.points.block[idx]
	return int(b)
}

// HeapRadixSort
// Sort the heap writer by the specified dim. It is used to sort the leaves of the tree
func (b *RadixSelector) HeapRadixSort(points *HeapPointWriter, from, to, dim, commonPrefixLength int) {
	dimOffset := dim*b.config.BytesPerDim() + commonPrefixLength
	dimCmpBytes := b.config.BytesPerDim() - commonPrefixLength
	dataOffset := b.config.PackedIndexBytesLength() - dimCmpBytes

	msbSorter := &msbRadixSorter{
		dimOffset:   dimOffset,
		dimCmpBytes: dimCmpBytes,
		dataOffset:  dataOffset,
		selector:    b,
		points:      points,
		buf:         new(bytes.Buffer),
	}
	sorter.NewMsbRadixSorter(b.bytesSorted-commonPrefixLength, msbSorter).Sort(from, to)
}

var _ sorter.MSBRadixInterface = &msbRadixSorter{}

type msbRadixSorter struct {
	dimOffset   int // 维度的offset
	dimCmpBytes int // 当前比对的维度的需要比较的字节长度
	dataOffset  int //
	selector    *RadixSelector
	points      *HeapPointWriter
	buf         *bytes.Buffer
}

func (m *msbRadixSorter) Compare(i, j, skipBytes int) int {
	valueI := m.Value(i)
	valueJ := m.Value(j)

	if skipBytes < len(valueI) && skipBytes <= len(valueJ) {
		v1 := valueI[skipBytes:]
		v2 := valueJ[skipBytes:]

		cmp := bytes.Compare(v1, v2)
		if cmp != 0 {
			return cmp
		}
	}

	sliceI := m.points.GetPackedValueSlice(i)
	sliceJ := m.points.GetPackedValueSlice(j)

	dataOffset := m.selector.config.NumIndexDims() * m.selector.config.BytesPerDim()
	cmp := bytes.Compare(
		sliceI.PackedValue()[dataOffset:m.selector.config.PackedBytesLength()],
		sliceJ.PackedValue()[dataOffset:m.selector.config.PackedBytesLength()],
	)
	if cmp != 0 {
		return cmp
	}

	docIdI := sliceI.DocID()
	docIdJ := sliceJ.DocID()
	if docIdI > docIdJ {
		return 1
	} else if docIdI < docIdJ {
		return -1
	} else {
		return 0
	}
}

func (m *msbRadixSorter) Swap(i, j int) {
	m.points.Swap(i, j)
}

func (m *msbRadixSorter) ByteAt(i int, k int) int {
	if k < m.dimCmpBytes {
		// dim bytes
		b := m.points.block[i*m.selector.config.BytesPerDoc()+m.dimOffset+k]
		return int(b)
	} else {
		// data bytes
		b := m.points.block[i*m.selector.config.BytesPerDoc()+m.dataOffset+k]
		return int(b)
	}
}

func (m *msbRadixSorter) Value(i int) []byte {
	from := i*m.selector.config.BytesPerDoc() + m.dimOffset
	to := from + m.dimCmpBytes
	return m.points.block[from:to]
}

func (b *RadixSelector) getDeltaPointWriter(left, right PointWriter, delta int, iteration int) (PointWriter, error) {
	v, err := b.getMaxPointsSortInHeap(left, right)
	if err != nil {
		return nil, err
	}

	if delta <= (v) {
		return NewHeapPointWriter(b.config, delta), nil
	} else {
		return NewOfflinePointWriter(b.config, b.tempDir, b.tempFileNamePrefix, fmt.Sprintf("delta%d", iteration), delta), nil
	}
}

func (b *RadixSelector) getMaxPointsSortInHeap(left, right PointWriter) (int, error) {
	pointsUsed := 0
	if w, ok := left.(*HeapPointWriter); ok {
		pointsUsed += w.size
	}
	if w, ok := right.(*HeapPointWriter); ok {
		pointsUsed += w.size
	}
	//assert maxPointsSortInHeap >= pointsUsed;
	return b.maxPointsSortInHeap - pointsUsed, nil
}

func (b *RadixSelector) getPointWriter(count int, desc string) PointWriter {
	// As we recurse, we hold two on-heap point writers at any point. Therefore the
	// max size for these objects is half of the total points we can have on-heap.
	if count <= b.maxPointsSortInHeap/2 {
		return NewHeapPointWriter(b.config, count)
	}
	return NewOfflinePointWriter(b.config, b.tempDir, b.tempFileNamePrefix, desc, count)
}

func (b *RadixSelector) getDimValues(bs []byte, dim int) []byte {
	fromIndex := dim * b.config.BytesPerDim()
	toIndex := fromIndex + b.config.BytesPerDim()
	return bs[fromIndex:toIndex]
}

// PathSlice Sliced reference to points in an PointWriter.
type PathSlice struct {
	writer PointWriter
	start  int
	count  int
}

func NewPathSlice(writer PointWriter, start, count int) *PathSlice {
	return &PathSlice{
		writer: writer,
		start:  start,
		count:  count,
	}
}

func (p *PathSlice) PointWriter() PointWriter {
	return p.writer
}

func (p *PathSlice) Start() int {
	return p.start
}

func (p *PathSlice) Count() int {
	return p.count
}
