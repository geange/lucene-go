package bkd

import (
	"encoding/binary"
)

var _ PointWriter = &HeapPointWriter{}

// HeapPointWriter Utility class to write new points into in-heap arrays.
// lucene.internal
type HeapPointWriter struct {
	block      []byte
	size       int
	config     *BKDConfig
	scratch    []byte
	nextWrite  int
	closed     bool
	pointValue *HeapPointValue
}

func NewHeapPointWriter(config *BKDConfig, size int) *HeapPointWriter {
	writer := &HeapPointWriter{
		block:      make([]byte, config.BytesPerDoc*size),
		size:       size,
		config:     config,
		scratch:    make([]byte, config.BytesPerDoc),
		nextWrite:  0,
		closed:     false,
		pointValue: nil,
	}
	if size > 0 {
		writer.pointValue = NewHeapPointValue(config, writer.block)
	}
	return writer
}

func (h *HeapPointWriter) Close() error {
	h.closed = true
	return nil
}

func (h *HeapPointWriter) Swap(i, j int) {
	indexI := i * h.config.BytesPerDoc
	indexJ := j * h.config.BytesPerDoc

	copy(h.scratch, h.block[indexI:indexI+h.config.BytesPerDoc])
	copy(h.block[indexI:], h.block[indexJ:indexJ+h.config.BytesPerDoc])
	copy(h.block[indexJ:], h.scratch[:h.config.BytesPerDoc])
}

func (h *HeapPointWriter) ComputeCardinality(form, to int, commonPrefixLengths []int) int {
	leafCardinality := 1
	for i := form + 1; i < to; i++ {
		for dim := 0; dim < h.config.NumDims; dim++ {
			start := dim*h.config.BytesPerDim + commonPrefixLengths[dim]
			end := dim*h.config.BytesPerDim + h.config.BytesPerDim

			if Mismatch(h.block[i*h.config.BytesPerDoc+start:i*h.config.BytesPerDoc+end],
				h.block[(i-1)*h.config.BytesPerDoc+start:(i-1)*h.config.BytesPerDoc+end]) != -1 {
				leafCardinality++
				break
			}
		}
	}
	return leafCardinality
}

// Mismatch 返回a、b两个字节数组
// * 第一个不相同字节的索引
// * 完全一致则返回-1，
// * 返回较短的数组的长度
func Mismatch(a, b []byte) int {
	size := len(a)
	if len(b) < size {
		size = len(b)
	}

	for i := 0; i < size; i++ {
		if a[i] != b[i] {
			return i
		}
	}

	if len(a) == len(b) {
		return -1
	}
	return size
}

func (h *HeapPointWriter) GetPackedValueSlice(index int) PointValue {
	h.pointValue.SetOffset(index * h.config.BytesPerDoc)
	return h.pointValue
}

func (h *HeapPointWriter) Append(packedValue []byte, docID int) error {
	//assert closed == false : "point writer is already closed";
	//assert packedValue.length == config.packedBytesLength : "[packedValue] must have length [" + config.packedBytesLength + "] but was [" + packedValue.length + "]";
	//assert nextWrite < size : "nextWrite=" + (nextWrite + 1) + " vs size=" + size;
	copy(h.block[h.nextWrite*h.config.BytesPerDoc:], packedValue[:h.config.PackedBytesLength])
	position := h.nextWrite*h.config.BytesPerDoc + h.config.PackedBytesLength
	binary.BigEndian.PutUint32(h.block[position:], uint32(docID))
	h.nextWrite++
	return nil
}

func (h *HeapPointWriter) AppendValue(pointValue PointValue) error {
	//assert closed == false : "point writer is already closed";
	//assert nextWrite < size : "nextWrite=" + (nextWrite + 1) + " vs size=" + size;
	packedValueDocID := pointValue.PackedValueDocIDBytes()
	//assert packedValueDocID.length == config.bytesPerDoc : "[packedValue] must have length [" + (config.bytesPerDoc) + "] but was [" + packedValueDocID.length + "]";
	copy(h.block[h.nextWrite*h.config.BytesPerDoc:], packedValueDocID[:h.config.BytesPerDoc])
	h.nextWrite++
	return nil
}

func (h *HeapPointWriter) GetReader(start, length int64) (PointReader, error) {
	//assert closed : "point writer is still open and trying to get a reader";
	//assert start + length <= size: "start=" + start + " length=" + length + " docIDs.length=" + size;
	//assert start + length <= nextWrite: "start=" + start + " length=" + length + " nextWrite=" + nextWrite;
	return NewHeapPointReader(h.config, h.block, int(start), int(start+length)), nil
}

func (h *HeapPointWriter) Count() int64 {
	return int64(h.nextWrite)
}

func (h *HeapPointWriter) Destroy() error {
	return nil
}
