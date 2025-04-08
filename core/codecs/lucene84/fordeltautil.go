package lucene84

import (
	"context"

	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/packed"
)

var (
	IDENTITY_PLUS_ONE = make([]uint64, BLOCK_SIZE)
)

func init() {
	for i := 0; i < BLOCK_SIZE; i++ {
		IDENTITY_PLUS_ONE[i] = uint64(i + 1)
	}
}

// ForDeltaUtil
// Utility class to encode sequences of 128 small positive integers.
type ForDeltaUtil struct {
	codec *IntCodec
}

func NewForDeltaUtil() *ForDeltaUtil {
	return &ForDeltaUtil{
		codec: NewIntCodec(),
	}
}

func (f *ForDeltaUtil) EncodeDeltas(ctx context.Context, longs []uint64, out store.DataOutput) error {
	if longs[0] == 1 && allEqual(longs) { // happens with very dense postings
		return out.WriteByte(0)
	}

	or := uint64(0)
	for _, l := range longs {
		or |= l
	}
	bitsPerValue, err := packed.BitsRequired(int64(or))
	if err != nil {
		return err
	}

	if err := out.WriteByte(byte(bitsPerValue)); err != nil {
		return err
	}
	return f.codec.Encode(longs, bitsPerValue, out)
}

func (f *ForDeltaUtil) DecodeAndPrefixSum(ctx context.Context, in store.DataInput, base uint64, longs []uint64) error {
	bitsPerValue, err := in.ReadByte()
	if err != nil {
		return err
	}
	if bitsPerValue == 0 {
		prefixSumOfOnes(longs, base)
	} else {
		f.codec.DecodeAndPrefixSum(ctx, int(bitsPerValue), in, base, longs)
	}
	return nil
}

func prefixSumOfOnes(arr []uint64, base uint64) {
	copy(arr[:BLOCK_SIZE], IDENTITY_PLUS_ONE)
	// This loop gets auto-vectorized
	for i := 0; i < BLOCK_SIZE; i++ {
		arr[i] += base
	}
}

// Skip a sequence of 128 longs.
func (f *ForDeltaUtil) Skip(ctx context.Context, in store.DataInput) error {
	bitsPerValue, err := in.ReadByte()
	if err != nil {
		return err
	}
	if bitsPerValue != 0 {
		return in.SkipBytes(ctx, numBytes(int(bitsPerValue)))
	}
	return nil
}
