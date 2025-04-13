package lucene84

import (
	"context"
	"slices"

	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/packed"
)

type PForUtil struct {
	forUtil *ForUtil
}

func allEqual(l []uint64) bool {
	for i := 1; i < len(l); i++ {
		if l[i] != l[0] {
			return false
		}
	}
	return true
}

func (p *PForUtil) Encode(ctx context.Context, longs []uint64, out store.DataOutput) error {
	// At most 7 exceptions
	top8 := make([]int64, 8)
	for i := range top8 {
		top8[i] = -1
	}
	for i := 0; i < BLOCK_SIZE; i++ {
		if int64(longs[i]) > top8[0] {
			top8[0] = int64(longs[i])
			slices.Sort(top8) // For only 8 entries we just sort on every iteration instead of maintaining a PQ
		}
	}

	maxBitsRequired, err := packed.BitsRequired(top8[7])
	if err != nil {
		return err
	}
	// We store the patch on a byte, so we can't decrease the number of bits required by more than 8
	n, err := packed.BitsRequired(top8[0])
	if err != nil {
		return err
	}
	patchedBitsRequired := max(n, maxBitsRequired-8)
	numExceptions := 0
	maxUnpatchedValue := uint64((1 << patchedBitsRequired) - 1)
	for i := 1; i < 8; i++ {
		if top8[i] > int64(maxUnpatchedValue) {
			numExceptions++
		}
	}
	exceptions := make([]byte, numExceptions*2)
	if numExceptions > 0 {
		exceptionCount := 0
		for i := 0; i < BLOCK_SIZE; i++ {
			if longs[i] > maxUnpatchedValue {
				exceptions[exceptionCount*2] = byte(i)
				exceptions[exceptionCount*2+1] = byte(longs[i] >> patchedBitsRequired)
				longs[i] &= maxUnpatchedValue
				exceptionCount++
			}
		}
	}

	if allEqual(longs) && maxBitsRequired <= 8 {
		for i := 0; i < numExceptions; i++ {
			exceptions[2*i+1] = exceptions[2*i+1] << patchedBitsRequired
		}
		if err := out.WriteByte(byte(numExceptions << 5)); err != nil {
			return err
		}
		if err := out.WriteUvarint(ctx, longs[0]); err != nil {
			return err
		}
	} else {
		token := byte((numExceptions << 5) | patchedBitsRequired)
		if err := out.WriteByte(token); err != nil {
			return err
		}
		if err := p.forUtil.Encode(longs, patchedBitsRequired, out); err != nil {
			return err
		}
	}
	if _, err := out.Write(exceptions); err != nil {
		return err
	}
	return nil
}

func (p *PForUtil) Decode(ctx context.Context, in store.DataInput, longs []uint64) error {
	token, err := in.ReadByte()
	if err != nil {
		return err
	}
	bitsPerValue := int(token & 0x1f)
	numExceptions := int(token >> 5)
	if bitsPerValue == 0 {
		num, err := in.ReadUvarint(ctx)
		if err != nil {
			return err
		}
		for i := range longs {
			longs[i] = num
		}
	} else {
		if err := p.forUtil.Decode(ctx, bitsPerValue, in, longs); err != nil {
			return err
		}
	}
	for i := 0; i < numExceptions; i++ {
		idx, err := in.ReadByte()
		if err != nil {
			return err
		}
		n, err := in.ReadByte()
		if err != nil {
			return err
		}
		longs[idx] |= uint64(n) << bitsPerValue
	}
	return nil
}

func (p *PForUtil) Skip(ctx context.Context, in store.DataInput) error {
	token, err := in.ReadByte()
	if err != nil {
		return err
	}
	bitsPerValue := int(token & 0x1f)
	numExceptions := int(token >> 5)
	if bitsPerValue == 0 {
		if _, err := in.ReadUvarint(ctx); err != nil {
			return err
		}
		if err := in.SkipBytes(ctx, numExceptions<<1); err != nil {
			return err
		}
	} else {
		if err := in.SkipBytes(ctx, numBytes(bitsPerValue)+(numExceptions<<1)); err != nil {
			return err
		}
	}
	return nil
}
