package lucene80

import (
	"context"
	"errors"
	"io"
	"math/bits"

	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
)

// jump-table time/space trade-offs to consider:
// The block offsets and the block indexes could be stored in more compressed form with
// two PackedInts or two MonotonicDirectReaders.
// The DENSE ranks (default 128 shorts = 256 bytes) could likewise be compressed. But as there is
// at least 4096 set bits in DENSE blocks, there will be at least one rank with 2^12 bits, so it
// is doubtful if there is much to gain here.

const (
	BLOCK_SIZE               = 65536
	DENSE_BLOCK_LONGS        = BLOCK_SIZE / 64
	DEFAULT_DENSE_RANK_POWER = 9
	MAX_ARRAY_LENGTH         = (1 << 12) - 1
)

var _ types.DocIdSetIterator = &IndexedDISI{}

type IndexedDISI struct {
	slice               store.IndexInput
	jumpTableEntryCount int
	denseRankPower      int8
	jumpTable           store.RandomAccessInput // Skip blocks of 64K bits
	denseRankTable      []byte
	cost                int64

	block             int
	blockEnd          int64
	denseBitmapOffset int64
	nextBlockIndex    int
	method            Method

	doc   int
	index int

	// SPARSE variables
	exists bool

	// DENSE variables
	word      int64
	wordIndex int
	// number of one bits encountered so far, including those of `word`
	numberOfOnes int
	// Used with rank for jumps inside of DENSE as they are absolute instead of relative
	denseOrigoIndex int

	// ALL variables
	gap int
}

func NewIndexedDISI(in store.IndexInput, offset int64, length int, jumpTableEntryCount int, denseRankPower int8) (*IndexedDISI, error) {
	blockSlice, err := createBlockSlice(in, "docs", offset, length, jumpTableEntryCount)
	if err != nil {
		return nil, err
	}

	jumpTable, err := createJumpTable(in, offset, length, jumpTableEntryCount)
	if err != nil {
		return nil, err
	}
	return newIndexedDISI(blockSlice, jumpTable, jumpTableEntryCount, denseRankPower)
}

func createJumpTable(slice store.IndexInput, offset int64, length int, jumpTableEntryCount int) (store.RandomAccessInput, error) {
	if jumpTableEntryCount <= 0 {
		return nil, nil
	} else {
		jumpTableBytes := jumpTableEntryCount * 4 * 2
		return slice.RandomAccessSlice(offset+int64(length)-int64(jumpTableBytes), int64(jumpTableBytes))
	}
}

func createBlockSlice(slice store.IndexInput, sliceDescription string, offset int64, length int, jumpTableEntryCount int) (store.IndexInput, error) {
	var jumpTableBytes int
	if jumpTableEntryCount < 0 {
		jumpTableBytes = 0
	} else {
		jumpTableBytes = jumpTableEntryCount * 4 * 2
	}
	return slice.Slice(sliceDescription, offset, int64(length-jumpTableBytes))
}

func newIndexedDISI(blockSlice store.IndexInput, jumpTable store.RandomAccessInput,
	jumpTableEntryCount int, denseRankPower int8) (*IndexedDISI, error) {

	if (denseRankPower < 7 || denseRankPower > 15) && denseRankPower != -1 {
		return nil, errors.New("denseRankPower must be 0 or -1")
	}

	this := &IndexedDISI{
		slice:               blockSlice,
		jumpTable:           jumpTable,
		jumpTableEntryCount: jumpTableEntryCount,
		denseRankPower:      denseRankPower,
	}

	rankIndexShift := denseRankPower - 7

	if denseRankPower == -1 {
		this.denseRankTable = nil
	} else {
		this.denseRankTable = make([]byte, DENSE_BLOCK_LONGS>>rankIndexShift)
	}

	return this, nil
}

func (i *IndexedDISI) DocID() int {
	return i.doc
}

func (i *IndexedDISI) NextDoc(ctx context.Context) (int, error) {
	return i.Advance(ctx, i.doc+1)
}

func (i *IndexedDISI) Advance(ctx context.Context, target int) (int, error) {
	targetBlock := target & 0xFFFF0000
	if i.block < targetBlock {
		if err := i.advanceBlock(ctx, targetBlock); err != nil {
			return 0, err
		}
	}
	if i.block == targetBlock {
		found, err := i.method.AdvanceWithinBlock(ctx, i, target)
		if err != nil {
			return 0, err
		}
		if found {
			return i.doc, nil
		}
		if err := i.readBlockHeader(ctx); err != nil {
			return 0, err
		}
	}
	_, err := i.method.AdvanceWithinBlock(ctx, i, i.block)
	if err != nil {
		return 0, err
	}
	return i.doc, nil
}

func (i *IndexedDISI) SlowAdvance(ctx context.Context, target int) (int, error) {
	return types.SlowAdvanceWithContext(ctx, i, target)
}

func (i *IndexedDISI) Cost() int64 {
	return i.cost
}

func (i *IndexedDISI) AdvanceExact(ctx context.Context, target int) (bool, error) {
	targetBlock := target & 0xFFFF0000
	if i.block < targetBlock {
		if err := i.advanceBlock(ctx, targetBlock); err != nil {
			return false, err
		}
	}

	found := false

	if i.block == targetBlock {
		inBlock, err := i.method.AdvanceExactWithinBlock(ctx, i, target)
		if err != nil {
			return false, err
		}
		found = inBlock
	}

	i.doc = target
	return found, nil
}

func (i *IndexedDISI) advanceBlock(ctx context.Context, targetBlock int) error {
	blockIndex := targetBlock >> 16
	// If the destination block is 2 blocks or more ahead, we use the jump-table.
	if i.jumpTable != nil && blockIndex >= (i.block>>16)+2 {
		// If the jumpTableEntryCount is exceeded, there are no further bits. Last entry is always NO_MORE_DOCS
		var inRangeBlockIndex int
		if blockIndex < i.jumpTableEntryCount {
			inRangeBlockIndex = blockIndex
		} else {
			inRangeBlockIndex = i.jumpTableEntryCount - 1
		}

		integerBYTES := 4

		index, err := i.jumpTable.ReadU32(int64(inRangeBlockIndex * integerBYTES * 2))
		if err != nil {
			return err
		}

		offset, err := i.jumpTable.ReadU32(int64(inRangeBlockIndex*integerBYTES*2 + integerBYTES))
		if err != nil {
			return err
		}

		i.nextBlockIndex = int(index - 1) // -1 to compensate for the always-added 1 in readBlockHeader
		if _, err := i.slice.Seek(int64(offset), io.SeekStart); err != nil {
			return err
		}
		if err := i.readBlockHeader(ctx); err != nil {
			return err
		}
		return nil
	}

	// Fallback to iteration of blocks
	for {
		if _, err := i.slice.Seek(i.blockEnd, io.SeekStart); err != nil {
			return err
		}
		if err := i.readBlockHeader(ctx); err != nil {
			return err
		}

		if !(i.block < targetBlock) {
			break
		}
	}
	return nil
}

func (i *IndexedDISI) readBlockHeader(ctx context.Context) error {
	n, err := i.slice.ReadUint16(ctx)
	if err != nil {
		return err
	}
	i.block = int(uint32(n) << 16)

	n2, err := i.slice.ReadUint16(ctx)
	if err != nil {
		return err
	}

	numValues := 1 + int(n2)
	i.index = i.nextBlockIndex
	i.nextBlockIndex = i.index + numValues
	if numValues <= MAX_ARRAY_LENGTH {
		i.method = &MethodSPARSE{}
		i.blockEnd = i.slice.GetFilePointer() + int64(numValues<<1)
	} else if numValues == 65536 {
		i.method = &MethodALL{}
		i.blockEnd = i.slice.GetFilePointer()
		i.gap = i.block - i.index - 1
	} else {
		i.method = &MethodDENSE{}
		fp := i.slice.GetFilePointer()
		if i.denseRankTable != nil {
			fp += int64(len(i.denseRankTable))
		}
		i.denseBitmapOffset = fp
		i.blockEnd = i.denseBitmapOffset + (1 << 13)
		// Performance consideration: All rank (default 128 * 16 bits) are loaded up front. This should be fast with the
		// reusable byte[] buffer, but it is still wasted if the DENSE block is iterated in small steps.
		// If this results in too great a performance regression, a heuristic strategy might work where the rank data
		// are loaded on first in-block advance, if said advance is > X docIDs. The hope being that a small first
		// advance means that subsequent advances will be small too.
		// Another alternative is to maintain an extra slice for DENSE rank, but IndexedDISI is already slice-heavy.
		if i.denseRankPower != -1 {
			if _, err := i.slice.Read(i.denseRankTable); err != nil {
				return err
			}
		}
		i.wordIndex = -1
		i.numberOfOnes = i.index + 1
		i.denseOrigoIndex = i.numberOfOnes
	}
	return nil
}

func (i *IndexedDISI) Index() int {
	return i.index
}

type Method interface {
	// AdvanceWithinBlock
	// Advance to the first doc from the block that is equal to or greater than target.
	// Return true if there is such a doc and false otherwise.
	AdvanceWithinBlock(ctx context.Context, disi *IndexedDISI, target int) (bool, error)

	// AdvanceExactWithinBlock
	// Advance the iterator exactly to the position corresponding to the given target and
	// return whether this document exists.
	AdvanceExactWithinBlock(ctx context.Context, disi *IndexedDISI, target int) (bool, error)
}

var _ Method = &MethodSPARSE{}

type MethodSPARSE struct{}

func (m *MethodSPARSE) AdvanceWithinBlock(ctx context.Context, disi *IndexedDISI, target int) (bool, error) {
	targetInBlock := target & 0xFFFF
	// TODO: binary search
	for disi.index < disi.nextBlockIndex {
		docU16, err := disi.slice.ReadUint16(ctx)
		if err != nil {
			return false, err
		}
		doc := int(docU16)
		disi.index++
		if doc >= targetInBlock {
			disi.doc = disi.block | doc
			disi.exists = true
			return true, nil
		}
	}
	return false, nil
}

func (m *MethodSPARSE) AdvanceExactWithinBlock(ctx context.Context, disi *IndexedDISI, target int) (bool, error) {
	targetInBlock := target & 0xFFFF
	// TODO: binary search
	if target == disi.doc {
		return disi.exists, nil
	}
	for disi.index < disi.nextBlockIndex {
		docU16, err := disi.slice.ReadUint16(ctx)
		if err != nil {
			return false, err
		}
		doc := int(docU16)
		disi.index++
		if doc >= targetInBlock {
			if doc != targetInBlock {
				disi.index--
				if _, err := disi.slice.Seek(disi.slice.GetFilePointer()-2, io.SeekStart); err != nil {
					return false, err
				}
				break
			}
			disi.exists = true
			return true, nil
		}
	}
	disi.exists = false
	return false, nil
}

var _ Method = &MethodDENSE{}

type MethodDENSE struct {
}

func (m *MethodDENSE) AdvanceWithinBlock(ctx context.Context, disi *IndexedDISI, target int) (bool, error) {
	targetInBlock := target & 0xFFFF
	targetWordIndex := targetInBlock >> 6

	// If possible, skip ahead using the rank cache
	// If the distance between the current position and the target is < rank-longs
	// there is no sense in using rank
	if disi.denseRankPower != -1 && targetWordIndex-disi.wordIndex >= (1<<(disi.denseRankPower-6)) {
		if err := rankSkip(ctx, disi, targetInBlock); err != nil {
			return false, err
		}
	}

	for i := disi.wordIndex + 1; i <= targetWordIndex; i++ {
		word, err := disi.slice.ReadUint64(ctx)
		if err != nil {
			return false, err
		}
		disi.word = int64(word)

		disi.numberOfOnes += bits.OnesCount64(word)
	}
	disi.wordIndex = targetWordIndex

	leftBits := uint64(disi.word >> target)
	if leftBits != 0 {
		disi.doc = target + bits.TrailingZeros64(leftBits)
		disi.index = disi.numberOfOnes - bits.OnesCount64(uint64(leftBits))
		return true, nil
	}

	// There were no set bits at the wanted position. Move forward until one is reached
	for disi.wordIndex+1 < 1024 {
		disi.wordIndex++
		// This could use the rank cache to skip empty spaces >= 512 bits, but it seems unrealistic
		// that such blocks would be DENSE
		word, err := disi.slice.ReadUint64(ctx)
		if err != nil {
			return false, err
		}
		disi.word = int64(word)

		if disi.word != 0 {
			disi.index = disi.numberOfOnes
			disi.numberOfOnes += bits.OnesCount64(word)
			disi.doc = disi.block | (disi.wordIndex << 6) | bits.TrailingZeros64(word)
			return true, nil
		}
	}
	// No set bits in the block at or after the wanted position.
	return false, nil
}

func (m *MethodDENSE) AdvanceExactWithinBlock(ctx context.Context, disi *IndexedDISI, target int) (bool, error) {
	targetInBlock := target & 0xFFFF
	targetWordIndex := targetInBlock >> 6

	// If possible, skip ahead using the rank cache
	// If the distance between the current position and the target is < rank-longs
	// there is no sense in using rank
	if disi.denseRankPower != -1 && targetWordIndex-disi.wordIndex >= (1<<(disi.denseRankPower-6)) {
		if err := rankSkip(ctx, disi, targetInBlock); err != nil {
			return false, err
		}
	}

	for i := disi.wordIndex + 1; i <= targetWordIndex; i++ {
		word, err := disi.slice.ReadUint64(ctx)
		if err != nil {
			return false, err
		}
		disi.word = int64(word)

		disi.numberOfOnes += bits.OnesCount64(word)
	}
	disi.wordIndex = targetWordIndex

	leftBits := uint64(disi.word >> target)
	disi.index = disi.numberOfOnes - bits.OnesCount64(leftBits)
	return (leftBits & 1) != 0, nil
}

func rankSkip(ctx context.Context, disi *IndexedDISI, targetInBlock int) error {
	// Resolve the rank as close to targetInBlock as possible (maximum distance is 8 longs)
	// Note: rankOrigoOffset is tracked on block open, so it is absolute (e.g. don't add origo)
	rankIndex := targetInBlock >> disi.denseRankPower // Default is 9 (8 longs: 2^3 * 2^6 = 512 docIDs)

	rank := uint32(disi.denseRankTable[rankIndex<<1])<<8 | uint32(disi.denseRankTable[(rankIndex<<1)+1])

	// Position the counting logic just after the rank point
	rankAlignedWordIndex := rankIndex << disi.denseRankPower >> 6
	if _, err := disi.slice.Seek(disi.denseBitmapOffset+rankAlignedWordIndex*8, io.SeekStart); err != nil {
		return err
	}
	rankWord, err := disi.slice.ReadUint64(ctx)
	if err != nil {
		return err
	}
	denseNOO := int(rank) + bits.OnesCount64(rankWord)

	disi.wordIndex = rankAlignedWordIndex
	disi.word = int64(rankWord)
	disi.numberOfOnes = disi.denseOrigoIndex + denseNOO
	return nil
}

var _ Method = &MethodALL{}

type MethodALL struct {
}

func (m *MethodALL) AdvanceWithinBlock(ctx context.Context, disi *IndexedDISI, target int) (bool, error) {
	disi.doc = target
	disi.index = target - disi.gap
	return true, nil
}

func (m *MethodALL) AdvanceExactWithinBlock(ctx context.Context, disi *IndexedDISI, target int) (bool, error) {
	disi.index = target - disi.gap
	return true, nil
}

func WriteBitSet(ctx context.Context, it types.DocIdSetIterator, out store.IndexOutput, denseRankPower byte) (uint16, error) {
	panic("")
}
