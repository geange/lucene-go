package packed

import (
	"context"
	"math"

	"github.com/geange/lucene-go/core/util/zigzag"
)

// BlockPackedWriter
// A writer for large sequences of longs.
// The sequence is divided into fixed-size blocks and for each block, the difference between each value
// and the minimum value of the block is encoded using as few bits as possible. Memory usage of this
// class is proportional to the block size. Each block has an overhead between 1 and 10 bytes to store
// the minimum value and the number of bits per value of the block.
// Format:
// <BLock>BlockCount
// BlockCount: ⌈ ValueCount / BlockSize ⌉
// Block: <Header, (Ints)>
// Header: <Token, (MinValue)>
// Token: a byte, first 7 bits are the number of bits per value (bitsPerValue). If the 8th bit is 1, then MinValue (see next) is 0, otherwise MinValue and needs to be decoded
// MinValue: a zigzag-encoded  variable-length long whose value should be added to every int from the block to restore the original values
// Ints: If the number of bits per value is 0, then there is nothing to decode and all ints are equal to MinValue. Otherwise: BlockSize packed ints encoded on exactly bitsPerValue bits per value. They are the subtraction of the original values and MinValue
// 请参阅: BlockPackedReaderIterator, BlockPackedReader
// lucene.internal
type BlockPackedWriter struct {
}

var _ BlockPackedFlusher = &BlockPackedWriter{}

func (b BlockPackedWriter) Flush(ctx context.Context, w *AbstractBlockPackedWriter) error {
	// 如果使用int64转uint64，如果一个数为负数，会导致转换后的uint64的整形过大，因此会导致delta过大，导致空间利用率较低。
	minValue := int64(math.MaxInt64)
	maxValue := int64(math.MinInt64)

	for _, v := range w.values[:w.valuesOffset] {
		num := int64(v)
		minValue = min(minValue, num)
		maxValue = max(maxValue, num)
	}

	delta := uint64(maxValue - minValue)

	var bitsRequired int
	if delta == 0 {
		bitsRequired = 0
	} else {
		bitsRequired = unsignedBitsRequired(delta)
	}
	if bitsRequired == 64 {
		// no need to delta-encode
		minValue = 0
	} else if minValue > 0 {
		// make min as small as possible so that writeVLong requires fewer bytes
		minValue = max(0, maxValue-int64(MaxValue(bitsRequired)))
	}

	var token byte
	if minValue == 0 {
		token = byte(bitsRequired)<<w.BPV_SHIFT | byte(w.MIN_VALUE_EQUALS_0)
	} else {
		token = byte(bitsRequired)<<w.BPV_SHIFT | 0
	}
	if err := w.out.WriteByte(token); err != nil {
		return err
	}

	if minValue != 0 {
		if err := w.out.WriteUvarint(ctx, zigzag.Encode(minValue)-1); err != nil {
			return err
		}
	}

	if bitsRequired > 0 {
		if minValue != 0 {
			for i := 0; i < w.valuesOffset; i++ {
				w.values[i] = uint64(int64(w.values[i]) - minValue)
			}
		}

		if err := w.writeValues(bitsRequired); err != nil {
			return err
		}
	}
	w.valuesOffset = 0
	return nil
}
