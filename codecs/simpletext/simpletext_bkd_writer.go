package simpletext

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/bits-and-blooms/bitset"
	"github.com/geange/lucene-go/codecs/bkd"
	"github.com/geange/lucene-go/codecs/utils"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util"
	"github.com/geange/lucene-go/core/util/numeric"
	"math"
	"sort"
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
	config              *bkd.BKDConfig
	scratch             *bytes.Buffer
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
	pointWriter         bkd.PointWriter
	finished            bool
	tempInput           store.IndexOutput
	maxPointsSortInHeap int

	// Minimum per-dim values, packed
	// 记录每个维度的最小的值
	minPackedValue []byte

	// Maximum per-dim values, packed
	// 记录每个维度的最大的值
	maxPackedValue []byte

	// 数据点的数量
	pointCount int64

	// An upper bound on how many points the caller will add (includes deletions)
	totalPointCount int64

	maxDoc int
}

func NewSimpleTextBKDWriter(maxDoc int, tempDir store.Directory, tempFileNamePrefix string,
	config *bkd.BKDConfig, maxMBSortInHeap float64, totalPointCount int64) *SimpleTextBKDWriter {
	return &SimpleTextBKDWriter{
		config:              config,
		scratch:             new(bytes.Buffer),
		tempDir:             store.NewTrackingDirectoryWrapper(tempDir),
		tempFileNamePrefix:  tempFileNamePrefix,
		maxMBSortInHeap:     maxMBSortInHeap,
		scratchDiff:         make([]byte, config.BytesPerDim),
		scratch1:            make([]byte, config.PackedBytesLength),
		scratch2:            make([]byte, config.PackedBytesLength),
		scratchBytesRef1:    new(bytes.Buffer),
		scratchBytesRef2:    new(bytes.Buffer),
		commonPrefixLengths: make([]int, config.NumDims),
		docsSeen:            bitset.New(uint(maxDoc)),
		pointWriter:         nil,
		finished:            false,
		tempInput:           nil,
		maxPointsSortInHeap: int(maxMBSortInHeap*1024*1024) / (config.BytesPerDoc * config.NumDims),
		minPackedValue:      make([]byte, config.PackedIndexBytesLength),
		maxPackedValue:      make([]byte, config.PackedIndexBytesLength),
		pointCount:          0,
		totalPointCount:     totalPointCount,
		maxDoc:              maxDoc,
	}
}

func (s *SimpleTextBKDWriter) Add(packedValue []byte, docID int) error {
	if len(packedValue) != s.config.PackedBytesLength {
		return fmt.Errorf("packedValue should be length=%d (got: %d)",
			s.config.PackedBytesLength, len(packedValue))
	}

	if s.pointCount >= s.totalPointCount {
		return fmt.Errorf("totalPointCount=%d was passed when we were created, but we just hit %d values",
			s.totalPointCount, s.pointCount+1)
	}

	if s.pointCount == 0 {
		// assert pointWriter == null : "Point writer is already initialized";
		if s.pointWriter != nil {
			return errors.New("point writer is already initialized")
		}

		//total point count is an estimation but the final point count must be equal or lower to that number.
		if int(s.totalPointCount) > s.maxPointsSortInHeap {
			pointWriter := bkd.NewOfflinePointWriter(s.config, s.tempDir, s.tempFileNamePrefix, "spill", 0)

			s.pointWriter = pointWriter
			s.tempInput = pointWriter.GetIndexOutput()
		} else {
			s.pointWriter = bkd.NewHeapPointWriter(s.config, int(s.totalPointCount))
		}

		copy(s.minPackedValue, packedValue[:s.config.PackedIndexBytesLength])
		copy(s.maxPackedValue, packedValue[:s.config.PackedIndexBytesLength])
	} else {
		for dim := 0; dim < s.config.NumIndexDims; dim++ {
			fromIndex := dim * s.config.BytesPerDim
			toIndex := fromIndex + s.config.BytesPerDim

			if bytes.Compare(packedValue[fromIndex:toIndex], s.minPackedValue[fromIndex:toIndex]) < 0 {
				copy(s.minPackedValue[fromIndex:toIndex], packedValue[fromIndex:toIndex])
			}

			if bytes.Compare(packedValue[fromIndex:toIndex], s.maxPackedValue[fromIndex:toIndex]) > 0 {
				copy(s.maxPackedValue[fromIndex:toIndex], packedValue[fromIndex:toIndex])
			}
		}
	}

	if err := s.pointWriter.Append(packedValue, docID); err != nil {
		return err
	}
	s.pointCount++
	s.docsSeen.Set(uint(docID))
	return nil
}

// GetPointCount How many points have been added so far
func (s *SimpleTextBKDWriter) GetPointCount() int64 {
	return s.pointCount
}

// WriteField Write a field from a MutablePointValues. This way of writing points is faster than regular
// writes with add since there is opportunity for reordering points before writing them to disk.
// This method does not use transient disk in order to reorder points.
func (s *SimpleTextBKDWriter) WriteField(out store.IndexOutput, fieldName string, reader index.MutablePointValues) (int64, error) {
	if s.config.NumIndexDims == 1 {
		return s.writeField1Dim(out, fieldName, reader)
	} else {
		return s.writeFieldNDims(out, fieldName, reader)
	}
}

// Finish Writes the BKD tree to the provided IndexOutput and returns the file offset where index was written.
func (s *SimpleTextBKDWriter) Finish(out store.IndexOutput) (int64, error) {
	// TODO: specialize the 1D case?  it's much faster at indexing time (no partitioning on recurse...)

	// Catch user silliness:
	if s.pointCount == 0 {
		return 0, errors.New("must index at least one point")
	}

	// Catch user silliness:
	if s.finished == true {
		return 0, errors.New("already finished")
	}

	//mark as finished
	s.finished = true

	if err := s.pointWriter.Close(); err != nil {
		return 0, err
	}
	points := bkd.NewPathSlice(s.pointWriter, 0, s.pointCount)
	//clean up pointers
	s.tempInput = nil
	s.pointWriter = nil

	// 计算 innerNodeCount 的数量
	countPerLeaf := s.pointCount
	innerNodeCount := 1

	for countPerLeaf > int64(s.config.MaxPointsInLeafNode) {
		countPerLeaf = (countPerLeaf + 1) / 2
		innerNodeCount *= 2
	}

	numLeaves := innerNodeCount

	if err := s.checkMaxLeafNodeCount(numLeaves); err != nil {
		return 0, err
	}

	// NOTE: we could save the 1+ here, to use a bit less heap at search time, but then we'd need a somewhat costly check at each
	// step of the recursion to recompute the split dim:

	// Indexed by nodeID, but first (root) nodeID is 1.
	// We do 1+ because the lead byte at each recursion says which dim we split on.
	// 使用 nodeID 进行索引。第一个节点的从index=1开始。index=0的数据用于说明我们使用哪个维度进行拆分
	// 1+s.config.BytesPerDim 表示 第一个字节用于存储维度信息，剩余字节用于存储维度的值
	splitPackedValues := make([]byte, numLeaves*(1+s.config.BytesPerDim))

	// +1 because leaf count is power of 2 (e.g. 8), and innerNodeCount is power of 2 minus 1 (e.g. 7)
	leafBlockFPs := make([]int64, numLeaves)

	//We re-use the selector so we do not need to create an object every time.
	radixSelector := bkd.NewBKDRadixSelector(s.config, s.maxPointsSortInHeap, s.tempDir, s.tempFileNamePrefix)

	err := s.buildV2(1, numLeaves, points, out,
		radixSelector, s.minPackedValue, s.maxPackedValue,
		splitPackedValues, leafBlockFPs, make([]int, s.config.MaxPointsInLeafNode))
	if err != nil {
		return 0, err
	}

	// Write index:
	indexFP := out.GetFilePointer()
	if err := s.writeIndex(out, leafBlockFPs, splitPackedValues); err != nil {
		return 0, err
	}
	return indexFP, nil
}

/* In the 2+D case, we recursively pick the split dimension, compute the
 * median value and partition other values around it. */
func (s *SimpleTextBKDWriter) writeFieldNDims(out store.IndexOutput, fieldName string, values index.MutablePointValues) (int64, error) {
	if s.pointCount != 0 {
		return 0, errors.New("cannot mix add and writeField")
	}

	// Catch user silliness:
	if s.finished {
		return 0, errors.New("already finished")
	}

	// Mark that we already finished:
	s.finished = true
	countPerLeaf := values.Size()
	s.pointCount = countPerLeaf
	innerNodeCount := int64(1)

	for countPerLeaf > int64(s.config.MaxPointsInLeafNode) {
		countPerLeaf = (countPerLeaf + 1) / 2
		innerNodeCount *= 2
	}

	numLeaves := int(innerNodeCount)

	if err := s.checkMaxLeafNodeCount(numLeaves); err != nil {
		return 0, err
	}

	splitPackedValues := make([]byte, numLeaves*(s.config.BytesPerDim+1))
	leafBlockFPs := make([]int64, numLeaves)

	// compute the min/max for this slice
	for i := range s.minPackedValue {
		s.minPackedValue[i] = 0xff
	}
	for i := range s.maxPackedValue {
		s.maxPackedValue[i] = 0
	}

	for i := 0; i < int(s.pointCount); i++ {
		values.GetValue(i, s.scratchBytesRef1)

		for dim := 0; dim < s.config.NumIndexDims; dim++ {
			offset := dim * s.config.BytesPerDim
			bs := s.scratchBytesRef1.Bytes()

			if bytes.Compare(
				bs[offset:offset+s.config.BytesPerDim],
				s.minPackedValue[offset:offset+s.config.BytesPerDim]) < 0 {

				copy(s.minPackedValue[offset:offset+s.config.BytesPerDim], bs[offset:])
			}

			if bytes.Compare(
				bs[offset:offset+s.config.BytesPerDim],
				s.maxPackedValue[offset:offset+s.config.BytesPerDim]) > 0 {

				copy(s.maxPackedValue[offset:offset+s.config.BytesPerDim], bs[offset:])
			}
		}

		s.docsSeen.Set(uint(values.GetDocCount()))
	}

	spareDocIds := make([]int, s.config.MaxPointsInLeafNode)
	s.buildV1(1, numLeaves, values, 0, int(s.pointCount), out,
		s.minPackedValue, s.maxPackedValue, splitPackedValues, leafBlockFPs, spareDocIds)

	indexFP := out.GetFilePointer()
	if err := s.writeIndex(out, leafBlockFPs, splitPackedValues); err != nil {
		return 0, err
	}
	return indexFP, nil
}

func (s *SimpleTextBKDWriter) writeIndex(out store.IndexOutput, leafBlockFPs []int64, splitPackedValues []byte) error {
	w := utils.NewTextWriter(out)

	w.WriteBytes(NUM_DATA_DIMS)
	w.WriteInt(s.config.NumDims)
	w.NewLine()

	w.WriteBytes(NUM_INDEX_DIMS)
	w.WriteInt(s.config.NumIndexDims)
	w.NewLine()

	w.WriteBytes(BYTES_PER_DIM)
	w.WriteInt(s.config.BytesPerDim)
	w.NewLine()

	w.WriteBytes(MAX_LEAF_POINTS)
	w.WriteInt(s.config.MaxPointsInLeafNode)
	w.NewLine()

	w.WriteBytes(INDEX_COUNT)
	w.WriteInt(len(leafBlockFPs))
	w.NewLine()

	w.WriteBytes(MIN_VALUE)
	w.WriteString(util.BytesToString(s.minPackedValue))
	w.NewLine()

	w.WriteBytes(MAX_VALUE)
	w.WriteString(util.BytesToString(s.maxPackedValue))
	w.NewLine()

	w.WriteBytes(POINT_COUNT)
	w.WriteLong(s.pointCount)
	w.NewLine()

	w.WriteBytes(DOC_COUNT)
	w.WriteLong(int64(s.docsSeen.Len()))
	w.NewLine()

	for i := 0; i < len(leafBlockFPs); i++ {
		w.WriteBytes(BLOCK_FP)
		w.WriteLong(leafBlockFPs[i])
		w.NewLine()
	}

	// assert (splitPackedValues.length % (1 + config.bytesPerDim)) == 0;
	count := len(splitPackedValues) / (1 + s.config.BytesPerDim)
	// assert count == leafBlockFPs.length;

	w.WriteBytes(SPLIT_COUNT)
	w.WriteInt(count)
	w.NewLine()

	for i := 0; i < count; i++ {
		w.WriteBytes(SPLIT_DIM)
		w.WriteInt(int(splitPackedValues[i*(1+s.config.BytesPerDim)] & 0xff))
		w.NewLine()
		w.WriteBytes(SPLIT_VALUE)

		offset := 1 + (i * (1 + s.config.BytesPerDim))
		endOffset := offset + s.config.BytesPerDim
		values := splitPackedValues[offset:endOffset]
		w.WriteString(util.BytesToString(values))
		w.NewLine()
	}
	return nil
}

func (s *SimpleTextBKDWriter) checkMaxLeafNodeCount(numLeaves int) error {
	if (1+s.config.BytesPerDim)*numLeaves > math.MaxInt32 {
		return fmt.Errorf("too many nodes; increase config.maxPointsInLeafNode (currently %d) and reindex",
			s.config.MaxPointsInLeafNode)
	}
	return nil
}

func (s *SimpleTextBKDWriter) buildV1(nodeID, leafNodeOffset int, reader index.MutablePointValues, from, to int,
	out store.IndexOutput, minPackedValue, maxPackedValue, splitPackedValues []byte,
	leafBlockFPs []int64, spareDocIds []int) error {

	if nodeID >= leafNodeOffset {
		// leaf node
		count := to - from
		// assert count <= config.maxPointsInLeafNode;
		for i := range s.commonPrefixLengths {
			s.commonPrefixLengths[i] = s.config.BytesPerDim
		}
		reader.GetValue(from, s.scratchBytesRef1)
		for i := from + 1; i < to; i++ {
			reader.GetValue(i, s.scratchBytesRef2)
			for dim := 0; dim < s.config.NumDims; dim++ {
				offset := dim * s.config.BytesPerDim
				for j := 0; j < s.commonPrefixLengths[dim]; j++ {
					if s.scratchBytesRef1.Bytes()[offset+j] != s.scratchBytesRef2.Bytes()[offset+j] {
						s.commonPrefixLengths[dim] = j
						break
					}
				}
			}
		}

		// Find the dimension that has the least number of unique bytes at commonPrefixLengths[dim]
		usedBytes := make([]*bitset.BitSet, s.config.NumDims)
		for dim := 0; dim < s.config.NumDims; dim++ {
			if s.commonPrefixLengths[dim] < s.config.BytesPerDim {
				usedBytes[dim] = bitset.New(256)
			}
		}

		for i := from + 1; i < to; i++ {
			for dim := 0; dim < s.config.NumDims; dim++ {
				if usedBytes[dim] != nil {
					b := reader.GetByteAt(i, dim*s.config.BytesPerDim+s.commonPrefixLengths[dim])
					usedBytes[dim].Set(uint(b))
				}
			}
		}

		sortedDim := 0
		sortedDimCardinality := math.MaxInt32
		for dim := 0; dim < s.config.NumDims; dim++ {
			if usedBytes[dim] != nil {
				cardinality := int(usedBytes[dim].Len())
				if cardinality < sortedDimCardinality {
					sortedDim = dim
					sortedDimCardinality = cardinality
				}
			}
		}

		// sort by sortedDim
		index.SortByDim(s.config, sortedDim, s.commonPrefixLengths,
			reader, from, to, s.scratchBytesRef1, s.scratchBytesRef2)

		// Save the block file pointer:
		leafBlockFPs[nodeID-leafNodeOffset] = out.GetFilePointer()

		// Write doc IDs
		docIDs := spareDocIds
		for i := from; i < to; i++ {
			docIDs[i-from] = reader.GetDocID(i)
		}
		if err := s.writeLeafBlockDocs(out, docIDs, 0, count); err != nil {
			return err
		}

		// Write the common prefixes:
		reader.GetValue(from, s.scratchBytesRef1)
		copy(s.scratch1, s.scratchBytesRef1.Bytes()[:s.config.PackedBytesLength])

		// Write the full values:
		packedValues := func(i int) []byte {
			reader.GetValue(from+i, s.scratchBytesRef1)
			return s.scratchBytesRef1.Bytes()
		}
		return s.writeLeafBlockPackedValues(out, s.commonPrefixLengths, count, sortedDim, packedValues)
	} else {
		// inner node

		// compute the split dimension and partition around it
		splitDim := s.split(minPackedValue, maxPackedValue)
		mid := (from + to + 1) >> 1
		commonPrefixLen := s.config.BytesPerDim
		for i := 0; i < s.config.BytesPerDim; i++ {
			if minPackedValue[splitDim*s.config.BytesPerDim+i] != maxPackedValue[splitDim*s.config.BytesPerDim+i] {
				commonPrefixLen = i
				break
			}
		}

		index.Partition(s.config, s.maxDoc, splitDim, commonPrefixLen,
			reader, from, to, mid, s.scratchBytesRef1, s.scratchBytesRef2)

		address := nodeID * (1 + s.config.BytesPerDim)
		splitPackedValues[address] = byte(splitDim)
		reader.GetValue(mid, s.scratchBytesRef1)
		offset := splitDim * s.config.BytesPerDim
		copy(splitPackedValues[address+1:], s.scratchBytesRef1.Bytes()[offset:offset+s.config.BytesPerDim])

		minSplitPackedValue := make([]byte, s.config.PackedIndexBytesLength)
		copy(minSplitPackedValue, minPackedValue)
		maxSplitPackedValue := make([]byte, s.config.PackedIndexBytesLength)
		copy(maxSplitPackedValue, maxPackedValue)

		srcPos := splitDim * s.config.BytesPerDim
		srcEndPos := srcPos + s.config.BytesPerDim
		destPos := splitDim * s.config.BytesPerDim
		copy(minSplitPackedValue[destPos:], s.scratchBytesRef1.Bytes()[srcPos:srcEndPos])
		copy(maxSplitPackedValue[destPos:], s.scratchBytesRef1.Bytes()[srcPos:srcEndPos])

		if err := s.buildV1(nodeID*2, leafNodeOffset, reader, from, mid, out,
			minPackedValue, maxSplitPackedValue, splitPackedValues, leafBlockFPs, spareDocIds); err != nil {
			return err
		}
		if err := s.buildV1(nodeID*2+1, leafNodeOffset, reader, mid, to, out,
			minSplitPackedValue, maxPackedValue, splitPackedValues, leafBlockFPs, spareDocIds); err != nil {
			return err
		}
	}
	return nil
}

func (s *SimpleTextBKDWriter) writeLeafBlockPackedValues(out store.IndexOutput, commonPrefixLengths []int,
	count, sortedDim int, packedValues func(int) []byte) error {

	for i := 0; i < count; i++ {
		packedValue := packedValues(i)
		// NOTE: we don't do prefix coding, so we ignore commonPrefixLengths
		if err := utils.WriteBytes(out, BLOCK_VALUE); err != nil {
			return err
		}
		if err := utils.WriteString(out, util.BytesToString(packedValue)); err != nil {
			return err
		}
		if err := utils.Newline(out); err != nil {
			return err
		}
	}
	return nil
}

func (s *SimpleTextBKDWriter) split(minPackedValue, maxPackedValue []byte) int {
	// Find which dim has the largest span so we can split on it:
	splitDim := -1
	for dim := 0; dim < s.config.NumIndexDims; dim++ {
		_ = numeric.Subtract(s.config.BytesPerDim, dim, maxPackedValue, minPackedValue, s.scratchDiff)
		if splitDim == -1 ||
			bytes.Compare(s.scratchDiff[:s.config.BytesPerDim], s.scratch1[:s.config.BytesPerDim]) > 0 {
			copy(s.scratch1, s.scratchDiff[:s.config.BytesPerDim])
			splitDim = dim
		}
	}
	return splitDim
}

func (s *SimpleTextBKDWriter) writeLeafBlockDocs(out store.IndexOutput, docIDs []int, start, count int) error {
	w := utils.NewTextWriter(out)
	if err := w.WriteBytes(BLOCK_COUNT); err != nil {
		return err
	}
	if err := w.WriteInt(count); err != nil {
		return err
	}
	if err := w.NewLine(); err != nil {
		return err
	}

	for i := 0; i < count; i++ {
		if err := w.WriteBytes(BLOCK_DOC_ID); err != nil {
			return err
		}
		if err := w.WriteInt(docIDs[start+i]); err != nil {
			return err
		}
		if err := w.NewLine(); err != nil {
			return err
		}
	}
	return nil
}

func (s *SimpleTextBKDWriter) buildV2(nodeID, leafNodeOffset int, points *bkd.PathSlice,
	out store.IndexOutput, radixSelector *bkd.BKDRadixSelector,
	minPackedValue, maxPackedValue, splitPackedValues []byte,
	leafBlockFPs []int64, spareDocIds []int) error {

	if nodeID >= leafNodeOffset {
		// Leaf node: write block
		// We can write the block in any order so by default we write it sorted by the dimension that has the
		// least number of unique bytes at commonPrefixLengths[dim], which makes compression more efficient
		var heapSource *bkd.HeapPointWriter

		if writer, ok := points.PointWriter().(*bkd.HeapPointWriter); ok {
			heapSource = writer
		} else {
			var err error
			heapSource, err = s.switchToHeap(writer)
			if err != nil {
				return err
			}
		}

		from, to := points.Start(), points.Start()+points.Count()

		//we store common prefix on scratch1
		s.computeCommonPrefixLength(heapSource, s.scratch1)

		sortedDim := 0
		sortedDimCardinality := math.MaxInt32
		//usedBytes := bitset.New(uint(s.config.NumDims))
		usedBytes := make([]*bitset.BitSet, 0)

		for dim := 0; dim < s.config.NumDims; dim++ {
			if s.commonPrefixLengths[dim] < s.config.BytesPerDim {
				usedBytes = append(usedBytes, bitset.New(256))
			} else {
				usedBytes = append(usedBytes, bitset.New(0))
			}
		}

		//Find the dimension to compress
		// 计算使用哪个维度进行索引
		for dim := 0; dim < s.config.NumDims; dim++ {
			prefix := s.commonPrefixLengths[dim]
			if prefix < s.config.BytesPerDim {
				offset := dim * s.config.BytesPerDim
				count := int(heapSource.Count())
				for i := 0; i < count; i++ {
					value := heapSource.GetPackedValueSlice(i)
					packedValue := value.PackedValue()
					bucket := packedValue[offset+prefix] & 0xff
					usedBytes[dim].Set(uint(bucket))
				}
				cardinality := int(usedBytes[dim].Count())
				if cardinality < sortedDimCardinality {
					sortedDim = dim
					sortedDimCardinality = cardinality
				}
			}
		}

		// sort the chosen dimension
		radixSelector.HeapRadixSort(heapSource, int(from), int(to), sortedDim, s.commonPrefixLengths[sortedDim])

		// Save the block file pointer:
		leafBlockFPs[nodeID-leafNodeOffset] = out.GetFilePointer()

		// Write docIDs first, as their own chunk, so that at intersect time we can add all docIDs w/o
		// loading the values:
		count := int(to - from)
		// assert count > 0: "nodeID=" + nodeID + " leafNodeOffset=" + leafNodeOffset;
		docIDs := spareDocIds
		for i := 0; i < count; i++ {
			docIDs[i] = heapSource.GetPackedValueSlice(int(from) + i).DocID()
		}
		if err := s.writeLeafBlockDocs(out, spareDocIds, 0, count); err != nil {
			return err
		}

		// TODO: minor opto: we don't really have to write the actual common prefixes, because BKDReader on recursing can regenerate it for us
		// from the index, much like how terms dict does so from the FST:

		// Write the full values:
		packedValues := func(i int) []byte {
			value := heapSource.GetPackedValueSlice(int(from) + i)
			return value.PackedValue()
		}
		return s.writeLeafBlockPackedValues(out, s.commonPrefixLengths, count, sortedDim, packedValues)
	}

	splitDim := 0
	if s.config.NumIndexDims > 1 {
		splitDim = s.split(minPackedValue, maxPackedValue)
	} else {
		splitDim = 0
	}

	// assert nodeID < splitPackedValues.length : "nodeID=" + nodeID + " splitValues.length=" + splitPackedValues.length;

	// How many points will be in the left tree:
	rightCount := points.Count() / 2
	leftCount := points.Count() - rightCount

	// 计算最大和最小的值的前缀
	fromIndex := splitDim * s.config.BytesPerDim
	toIndex := splitDim*s.config.BytesPerDim + s.config.BytesPerDim
	commonPrefixLen := bkd.Mismatch(minPackedValue[fromIndex:toIndex], maxPackedValue[fromIndex:toIndex])
	if commonPrefixLen == -1 {
		commonPrefixLen = s.config.BytesPerDim
	}

	pathSlices := make([]*bkd.PathSlice, 2)

	splitValue, err := radixSelector.Select(points, pathSlices,
		points.Start(), points.Start()+points.Count(), points.Start()+leftCount,
		splitDim, commonPrefixLen)
	if err != nil {
		return err
	}

	address := nodeID * (1 + s.config.BytesPerDim)
	splitPackedValues[address] = byte(splitDim)

	destPos := address + 1
	destEnd := destPos + s.config.BytesPerDim
	copy(splitPackedValues[destPos:destEnd], splitValue)

	minSplitPackedValue := make([]byte, s.config.PackedIndexBytesLength)
	copy(minSplitPackedValue, minPackedValue[:s.config.PackedIndexBytesLength])

	maxSplitPackedValue := make([]byte, s.config.PackedIndexBytesLength)
	copy(maxSplitPackedValue, maxPackedValue[:s.config.PackedIndexBytesLength])

	destPos = splitDim * s.config.BytesPerDim
	copy(minSplitPackedValue[destPos:], splitValue[:s.config.BytesPerDim])
	copy(maxSplitPackedValue[destPos:], splitValue[:s.config.BytesPerDim])

	// Recurse on left tree:
	err = s.buildV2(2*nodeID, leafNodeOffset, pathSlices[0], out, radixSelector,
		minPackedValue, maxSplitPackedValue, splitPackedValues, leafBlockFPs, spareDocIds)
	if err != nil {
		return err
	}

	// TODO: we could "tail recurse" here?  have our parent discard its refs as we recurse right?
	// Recurse on right tree:
	return s.buildV2(2*nodeID+1, leafNodeOffset, pathSlices[1], out, radixSelector,
		minSplitPackedValue, maxPackedValue, splitPackedValues, leafBlockFPs, spareDocIds)

}

func (s *SimpleTextBKDWriter) computeCommonPrefixLength(heapPointWriter *bkd.HeapPointWriter, commonPrefix []byte) {
	for i := range s.commonPrefixLengths {
		s.commonPrefixLengths[i] = s.config.BytesPerDim
	}
	value := heapPointWriter.GetPackedValueSlice(0)
	packedValue := value.PackedValue()
	for dim := 0; dim < s.config.NumDims; dim++ {
		start := +dim * s.config.BytesPerDim
		end := start + s.config.BytesPerDim
		copy(commonPrefix[start:], packedValue[start:end])
	}

	count := int(heapPointWriter.Count())
	for i := 1; i < count; i++ {
		value = heapPointWriter.GetPackedValueSlice(i)
		packedValue = value.PackedValue()
		for dim := 0; dim < s.config.NumDims; dim++ {
			if s.commonPrefixLengths[dim] != 0 {
				fromIndex := dim * s.config.BytesPerDim
				toIndex := dim*s.config.BytesPerDim + s.commonPrefixLengths[dim]
				j := bkd.Mismatch(commonPrefix[fromIndex:toIndex], packedValue[fromIndex:toIndex])
				if j != -1 {
					s.commonPrefixLengths[dim] = j
				}
			}
		}
	}
}

//func (s *SimpleTextBKDWriter) writeLeafBlockDocs(out store.IndexOutput, docIDs []int, start, count int) {
//
//}

// Pull a partition back into heap once the point count is low enough while recursing.
func (s *SimpleTextBKDWriter) switchToHeap(source bkd.PointWriter) (*bkd.HeapPointWriter, error) {
	count := source.Count()

	reader, err := source.GetReader(0, count)
	if err != nil {
		return nil, err
	}

	writer := bkd.NewHeapPointWriter(s.config, int(count))

	for i := 0; i < int(count); i++ {
		if _, err := reader.Next(); err != nil {
			return nil, err
		}

		if err := writer.AppendValue(reader.PointValue()); err != nil {
			return nil, err
		}
	}
	return writer, nil
}

/* In the 1D case, we can simply sort points in ascending order and use the
 * same writing logic as we use at merge time. */
func (s *SimpleTextBKDWriter) writeField1Dim(out store.IndexOutput, fieldName string, values index.MutablePointValues) (int64, error) {
	panic("")
}

func SortK(data sort.Interface, k int) {
	begin, end := 0, data.Len()-1
	sortN(begin, end, k, data)
}

func sortN(from, to, k int, data sort.Interface) {
	for from < to {
		loc := partition(from, to, data)
		if loc == k {
			return
		}
		if loc < k {
			sortN(loc+1, to, k, data)
			return
		}

		sortN(from, loc-1, k, data)
	}
}

func partition(begin, end int, data sort.Interface) int {
	i, j := begin+1, end

	for i < j {
		if data.Less(begin, i) {
			data.Swap(i, j)
			j--
		} else {
			i++
		}
	}

	// 如果 values[begin] <= values[i]
	// !data.Less(begin, i) && !data.Less(i, begin) => values[begin] == values[i]
	if data.Less(begin, i) || (!data.Less(begin, i) && !data.Less(i, begin)) {
		i--
	}
	data.Swap(i, begin)
	return i
}
