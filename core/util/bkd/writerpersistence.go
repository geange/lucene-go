package bkd

import (
	"context"
	"github.com/geange/lucene-go/codecs/utils"
	"github.com/geange/lucene-go/core/store"
	"github.com/samber/lo"
)

// 写入索引的
func (w *Writer) writeIndex(ctx context.Context, metaOut, indexOut store.IndexOutput, countPerLeaf int, leafNodes LeafNodes, dataStartFP int64) error {

	// 计算
	packedIndex, err := w.packIndex(ctx, leafNodes)
	if err != nil {
		return err
	}

	config := w.config

	numDims := config.NumDims()
	numIndexDims := config.NumIndexDims()
	bytesPerDim := config.BytesPerDim()
	packedIndexBytesLength := config.PackedIndexBytesLength()
	numLeaves := leafNodes.NumLeaves()

	if err := utils.WriteHeader(ctx, metaOut, CODEC_NAME, VERSION_CURRENT); err != nil {
		return err
	}
	if err := metaOut.WriteUvarint(ctx, uint64(numDims)); err != nil {
		return err
	}
	if err := metaOut.WriteUvarint(ctx, uint64(numIndexDims)); err != nil {
		return err
	}
	if err := metaOut.WriteUvarint(ctx, uint64(countPerLeaf)); err != nil {
		return err
	}
	if err := metaOut.WriteUvarint(ctx, uint64(bytesPerDim)); err != nil {
		return err
	}
	if err := metaOut.WriteUvarint(ctx, uint64(numLeaves)); err != nil {
		return err
	}
	if _, err := metaOut.Write(w.minPackedValue[:packedIndexBytesLength]); err != nil {
		return err
	}
	if _, err := metaOut.Write(w.maxPackedValue[:packedIndexBytesLength]); err != nil {
		return err
	}
	if err := metaOut.WriteUvarint(ctx, uint64(w.pointCount)); err != nil {
		return err
	}
	if err := metaOut.WriteUvarint(ctx, uint64(w.docsSeen.Len())); err != nil {
		return err
	}
	if err := metaOut.WriteUvarint(ctx, uint64(len(packedIndex))); err != nil {
		return err
	}
	if err := metaOut.WriteUint64(ctx, uint64(dataStartFP)); err != nil {
		return err
	}

	// If metaOut and indexOut are the same file, we account for the fact that
	// writing a long makes the index start 8 bytes later.
	fp := indexOut.GetFilePointer()
	if metaOut == indexOut {
		fp += 8
	}
	if err := metaOut.WriteUint64(ctx, uint64(fp)); err != nil {
		return err
	}
	if _, err := indexOut.Write(packedIndex); err != nil {
		return err
	}
	return nil
}

func (w *Writer) writeLeafBlockDocs(out store.DataOutput, docIDs []int) error {
	if err := out.WriteUvarint(nil, uint64(len(docIDs))); err != nil {
		return err
	}
	return WriteDocIds(nil, docIDs, out)
}

type packedValuesFunc func(int) []byte

func (w *Writer) writeLeafBlockPackedValues(out store.DataOutput, commonPrefixLengths []int,
	count, sortedDim int, packedValues packedValuesFunc, leafCardinality int) error {
	config := w.config

	prefixLenSum := lo.Sum(commonPrefixLengths)
	if prefixLenSum == config.PackedBytesLength() {
		// all values in this block are equal
		return out.WriteByte(255)
	}

	// estimate if storing the values with cardinality is cheaper than storing all values.
	compressedByteOffset := sortedDim*config.BytesPerDim() + commonPrefixLengths[sortedDim]
	highCardinalityCost := 0
	lowCardinalityCost := 0
	if count == leafCardinality {
		// all values in this block are different
		// 高基数，所有值都不一样
		highCardinalityCost = 0
		lowCardinalityCost = 1
	} else {
		// compute cost of runLen compression
		numRunLens := 0
		for i := 0; i < count; {
			// do run-length compression on the byte at compressedByteOffset
			runLen := runLen(packedValues, i, min(i+0xff, count), compressedByteOffset)
			// assert runLen <= 0xff;
			numRunLens++
			i += runLen
		}
		// Add cost of runLen compression
		highCardinalityCost = count*(config.PackedBytesLength()-prefixLenSum-1) + 2*numRunLens
		// +1 is the byte needed for storing the cardinality
		lowCardinalityCost = leafCardinality * (config.PackedBytesLength() - prefixLenSum + 1)
	}

	if lowCardinalityCost <= highCardinalityCost {
		if err := out.WriteByte(254); err != nil {
			return err
		}
		return w.writeLowCardinalityLeafBlockPackedValues(out, commonPrefixLengths, count, packedValues)
	}

	// 写高基数
	// 写排序的维度是哪个
	if err := out.WriteByte(byte(sortedDim)); err != nil {
		return err
	}
	return w.writeHighCardinalityLeafBlockPackedValues(out, commonPrefixLengths, count, sortedDim, packedValues, compressedByteOffset)
}

func (w *Writer) writeLowCardinalityLeafBlockPackedValues(out store.DataOutput, commonPrefixLengths []int,
	count int, packedValues packedValuesFunc) error {
	config := w.config

	if config.numIndexDims != 1 {
		if err := w.writeActualBounds(out, commonPrefixLengths, count, packedValues); err != nil {
			return err
		}
	}
	value := packedValues(0)
	arraycopy(value, 0, w.scratch1, 0, config.packedBytesLength)
	cardinality := 1
	for i := 1; i < count; i++ {
		value = packedValues(i)
		for dim := 0; dim < config.numDims; dim++ {
			start := dim*config.bytesPerDim + commonPrefixLengths[dim]
			end := dim*config.bytesPerDim + config.bytesPerDim
			if Mismatch(value[start:end], w.scratch1[start:end]) != -1 {
				if err := out.WriteUvarint(nil, uint64(cardinality)); err != nil {
					return err
				}
				for j := 0; j < config.numDims; j++ {
					offset := j*config.bytesPerDim + commonPrefixLengths[j]
					size := config.bytesPerDim - commonPrefixLengths[j]
					if _, err := out.Write(w.scratch1[offset : offset+size]); err != nil {
						return err
					}
				}
				arraycopy(value, 0, w.scratch1, 0, config.packedBytesLength)
				cardinality = 1
				break
			} else if dim == config.numDims-1 {
				cardinality++
			}
		}
	}
	if err := out.WriteUvarint(nil, uint64(cardinality)); err != nil {
		return err
	}
	for i := 0; i < config.numDims; i++ {
		offset := i*config.bytesPerDim + commonPrefixLengths[i]
		size := config.bytesPerDim - commonPrefixLengths[i]
		if _, err := out.Write(w.scratch1[offset : offset+size]); err != nil {
			return err
		}
	}
	return nil
}

func (w *Writer) writeHighCardinalityLeafBlockPackedValues(out store.DataOutput, commonPrefixLengths []int,
	count, sortedDim int, packedValues packedValuesFunc, compressedByteOffset int) error {
	if w.config.numIndexDims != 1 {
		if err := w.writeActualBounds(out, commonPrefixLengths, count, packedValues); err != nil {
			return err
		}
	}

	commonPrefixLengths[sortedDim]++

	for i := 0; i < count; i++ {
		runLen := runLen(packedValues, i, min(i+0xff, count), compressedByteOffset)
		first := packedValues(i)
		prefixByte := first[compressedByteOffset]
		if err := out.WriteByte(prefixByte); err != nil {
			return err
		}
		if err := out.WriteByte(byte(runLen)); err != nil {
			return err
		}
		if err := w.writeLeafBlockPackedValuesRange(out, commonPrefixLengths, i, i+runLen, packedValues); err != nil {
			return err
		}
		i += runLen
	}
	return nil
}

func (w *Writer) writeActualBounds(out store.DataOutput, commonPrefixLengths []int,
	count int, packedValues packedValuesFunc) error {
	config := w.config

	for dim := 0; dim < config.NumIndexDims(); dim++ {
		commonPrefixLength := commonPrefixLengths[dim]
		suffixLength := config.BytesPerDim() - commonPrefixLength
		if suffixLength > 0 {
			minMax, err := w.computeMinMax(count, packedValues, dim*config.BytesPerDim()+commonPrefixLength, suffixLength)
			if err != nil {
				return err
			}
			if _, err := out.Write(minMax[0]); err != nil {
				return err
			}
			if _, err := out.Write(minMax[1]); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *Writer) writeLeafBlockPackedValuesRange(out store.DataOutput, commonPrefixLengths []int, start, end int,
	packedValues packedValuesFunc) error {
	config := w.config

	for i := start; i < end; i++ {
		ref := packedValues(i)
		// assert ref.length == config.packedBytesLength;

		for dim := 0; dim < config.NumDims(); dim++ {
			prefix := commonPrefixLengths[dim]
			fromIndex := dim*config.BytesPerDim() + prefix
			size := config.BytesPerDim() - prefix
			toIndex := fromIndex + size

			if _, err := out.Write(ref[fromIndex:toIndex]); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *Writer) writeCommonPrefixes(ctx context.Context, out store.DataOutput, commonPrefixes []int, packedValue []byte) error {
	config := w.config
	numDims := config.NumDims()
	bytesPerDim := config.BytesPerDim()

	for dim := 0; dim < numDims; dim++ {
		if err := out.WriteUvarint(ctx, uint64(commonPrefixes[dim])); err != nil {
			return err
		}
		start := dim * bytesPerDim
		end := start + commonPrefixes[dim]
		if _, err := out.Write(packedValue[start:end]); err != nil {
			return err
		}
	}
	return nil
}
