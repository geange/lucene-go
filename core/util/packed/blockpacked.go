package packed

import (
	"context"
	"errors"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util/zigzag"
	"io"
	"math"
)

var _ types.LongValues = &BlockPackedReader{}

func NewBlockPackedReader(ctx context.Context, in store.IndexInput,
	packedIntsVersion, blockSize, valueCount int, direct bool) (*BlockPackedReader, error) {
	blockShift := checkBlockSize(blockSize, MIN_BLOCK_SIZE, MAX_BLOCK_SIZE)
	blockMask := blockSize - 1

	reader := &BlockPackedReader{
		blockShift: blockShift,
		blockMask:  blockMask,
		valueCount: valueCount,
		minValues:  nil,
		sumBPV:     0,
	}

	numBlocks, err := getNumBlocks(valueCount, blockSize)
	if err != nil {
		return nil, err
	}

	reader.subReaders = make([]Reader, numBlocks)

	var minValues []uint64
	var sumBPV int64

	for i := 0; i < numBlocks; i++ {
		token, err := in.ReadByte()
		if err != nil {
			return nil, err
		}
		bitsPerValue := int(token >> ABP_BPV_SHIFT)
		sumBPV += int64(bitsPerValue)

		if bitsPerValue > 64 {
			return nil, errors.New("corrupted Block")
		}

		if token&ABP_MIN_VALUE_EQUALS_0 == 0 {
			if len(minValues) == 0 {
				minValues = make([]uint64, numBlocks)
			}
			n, err := in.ReadUvarint(ctx)
			if err != nil {
				return nil, err
			}
			minValues[i] = uint64(zigzag.Decode(n + 1))
		}

		if bitsPerValue == 0 {
			reader.subReaders[i] = NewNullReader(blockSize)
		} else {
			size := min(blockSize, valueCount-i*blockSize)
			if direct {
				pointer := in.GetFilePointer()
				subReader, err := getDirectReaderNoHeader(ctx, in, FormatPacked, packedIntsVersion, size, bitsPerValue)
				if err != nil {
					return nil, err
				}
				reader.subReaders[i] = subReader
				if _, err := in.Seek(pointer+int64(FormatPacked.ByteCount(packedIntsVersion, size, bitsPerValue)), io.SeekStart); err != nil {
					return nil, err
				}
			} else {
				subReader, err := getReaderNoHeader(ctx, in, FormatPacked, packedIntsVersion, size, bitsPerValue)
				if err != nil {
					return nil, err
				}
				reader.subReaders[i] = subReader
			}
		}
	}

	reader.minValues = minValues
	reader.sumBPV = sumBPV

	return reader, nil
}

// BlockPackedReader
// Provides random access to a stream written with BlockPackedWriter.
// lucene.internal
type BlockPackedReader struct {
	blockShift int
	blockMask  int
	valueCount int
	minValues  []uint64
	subReaders []Reader
	sumBPV     int64
}

func (b *BlockPackedReader) Get(index int) (uint64, error) {

	block := index >> b.blockShift
	idx := (index) & b.blockMask

	minValue := int64(0)
	if len(b.minValues) > block {
		minValue = int64(b.minValues[block])
	}

	num, err := b.subReaders[block].Get(idx)
	if err != nil {
		return 0, err
	}

	return uint64(int64(num) + minValue), nil
}

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
	*abstractBlockPackedWriter
}

func NewBlockPackedWriter(out store.DataOutput, blockSize int) *BlockPackedWriter {
	res := &BlockPackedWriter{}
	res.abstractBlockPackedWriter = newAbstractBlockPackedWriter(out, blockSize)
	res.abstractBlockPackedWriter.flusher = res
	return res
}

var _ BlockPackedFlusher = &BlockPackedWriter{}

func (b *BlockPackedWriter) Flush(ctx context.Context) error {
	minValue, maxValue := int64(math.MaxInt64), int64(math.MinInt64)
	for i := 0; i < b.off; i++ {
		v := int64(b.values[i])
		minValue = min(minValue, v)
		maxValue = max(maxValue, v)
	}

	delta := maxValue - minValue

	var bitsRequired int
	if delta == 0 {
		bitsRequired = 0
	} else {
		bitsRequired = unsignedBitsRequired(uint64(delta))
	}

	if bitsRequired == 64 {
		// no need to delta-encode
		minValue = 0
	} else if minValue > 0 {
		// make min as small as possible so that writeVLong requires fewer bytes
		minValue = int64(max(0, uint64(maxValue)-MaxValue(bitsRequired)))
	}

	token := bitsRequired << ABP_BPV_SHIFT
	if minValue == 0 {
		token |= ABP_MIN_VALUE_EQUALS_0
	}

	if err := b.out.WriteByte(byte(token)); err != nil {
		return err
	}

	if minValue != 0 {
		n := zigzag.Encode(minValue) - 1
		if err := b.out.WriteUvarint(ctx, n); err != nil {
			return err
		}
	}

	if bitsRequired > 0 {
		if minValue != 0 {
			for i := 0; i < b.off; i++ {
				b.values[i] = uint64(int64(b.values[i]) - minValue)
			}
		}
		if err := b.writeValues(bitsRequired); err != nil {
			return err
		}
	}

	b.off = 0
	return nil
}

func (b *BlockPackedWriter) Add(ctx context.Context, u uint64) error {
	return b.add(ctx, u)
}

// BlockPackedReaderIterator
// Reader for sequences of longs written with BlockPackedWriter.
// BlockPackedWriter
// lucene.internal
type BlockPackedReaderIterator struct {
	in                store.DataInput
	packedIntsVersion int
	valueCount        int
	blockSize         int
	values            []uint64
	valuesRef         []uint64
	blocks            []byte
	off               int
	ord               int
}

// NewBlockPackedReaderIterator
// Sole constructor.
// blockSize: the number of values of a block, must be equal to the block size of the BlockPackedWriter which has been used to write the stream
func NewBlockPackedReaderIterator(in store.DataInput,
	packedIntsVersion, blockSize, valueCount int) *BlockPackedReaderIterator {
	checkBlockSize(blockSize, MIN_BLOCK_SIZE, MAX_BLOCK_SIZE)
	res := &BlockPackedReaderIterator{}
	res.packedIntsVersion = packedIntsVersion
	res.blockSize = blockSize
	res.values = make([]uint64, blockSize)
	res.Reset(in, valueCount)
	return res
}

// Reset the current reader to wrap a stream of valueCount values contained in in. The block size remains unchanged.
func (b *BlockPackedReaderIterator) Reset(in store.DataInput, valueCount int) {
	b.in = in
	b.valueCount = valueCount
	b.off = b.blockSize
	b.ord = 0
}

// Skip exactly count values.
func (b *BlockPackedReaderIterator) Skip(ctx context.Context, count int) error {
	if b.ord+count > b.valueCount || b.ord+count < 0 {
		return io.EOF
	}

	// 1. skip buffered values
	skipBuffer := min(count, b.blockSize-b.off)
	b.off += skipBuffer
	b.ord += skipBuffer
	count -= skipBuffer
	if count == 0 {
		return nil
	}

	// 2. skip as many blocks as necessary
	for count >= b.blockSize {
		token, err := b.in.ReadByte()
		if err != nil {
			return err
		}
		bitsPerValue := int(token >> ABP_BPV_SHIFT)
		if bitsPerValue > 64 {
			return errors.New("corrupted")
		}
		if (token & ABP_MIN_VALUE_EQUALS_0) == 0 {
			if _, err := b.in.ReadUvarint(ctx); err != nil {
				return err
			}
		}
		blockBytes := FormatPacked.ByteCount(b.packedIntsVersion, b.blockSize, bitsPerValue)
		if err := b.skipBytes(blockBytes); err != nil {
			return err
		}
		b.ord += b.blockSize
		count -= b.blockSize
	}
	if count == 0 {
		return nil
	}

	// 3. skip last values
	if err := b.refill(ctx); err != nil {
		return err
	}
	b.ord += count
	b.off += count
	return nil
}

func (b *BlockPackedReaderIterator) skipBytes(count int) error {
	if ii, ok := b.in.(store.IndexInput); ok {
		if _, err := ii.Seek(ii.GetFilePointer()+int64(count), io.SeekStart); err != nil {
			return err
		}
		return nil
	}

	if len(b.blocks) == 0 {
		b.blocks = make([]byte, b.blockSize)
	}
	skipped := 0
	for skipped < count {
		toSkip := min(len(b.blocks), count-skipped)
		if _, err := b.in.Read(b.blocks[:toSkip]); err != nil {
			return err
		}
		skipped += toSkip
	}
	return nil
}

// Next Read the next value.
func (b *BlockPackedReaderIterator) Next(ctx context.Context) (uint64, error) {
	if b.ord == b.valueCount {
		return 0, io.EOF
	}
	if b.off == b.blockSize {
		if err := b.refill(ctx); err != nil {
			return 0, err
		}
	}
	value := b.values[b.off]
	b.off++
	b.ord++
	return value, nil
}

func (b *BlockPackedReaderIterator) NextSlices(ctx context.Context, count int) ([]uint64, error) {
	if b.ord == b.valueCount {
		return nil, io.EOF
	}

	if b.off == b.blockSize {
		if err := b.refill(ctx); err != nil {
			return nil, err
		}
	}

	count = min(count, b.blockSize-b.off, b.valueCount-b.ord)

	b.valuesRef = b.values[b.off : b.off+count]

	b.off += count
	b.ord += count

	return b.valuesRef, nil
}

// Read between 1 and count values.
func (b *BlockPackedReaderIterator) refill(ctx context.Context) error {
	token, err := b.in.ReadByte()
	if err != nil {
		return err
	}
	minEquals0 := (token & ABP_MIN_VALUE_EQUALS_0) != 0
	bitsPerValue := int(token >> ABP_BPV_SHIFT)
	if bitsPerValue > 64 {
		return errors.New("corrupted")
	}
	var minValue int64
	if minEquals0 {
		minValue = 0
	} else {
		num, err := b.in.ReadUvarint(ctx)
		if err != nil {
			return err
		}
		minValue = zigzag.Decode(num + 1)
	}

	if bitsPerValue == 0 {
		for i := range b.values {
			b.values[i] = uint64(minValue)
		}
	} else {
		decoder, err := GetDecoder(FormatPacked, b.packedIntsVersion, bitsPerValue)
		if err != nil {
			return err
		}

		iterations := b.blockSize / decoder.ByteValueCount()
		blocksSize := iterations * decoder.ByteBlockCount()
		if len(b.blocks) < blocksSize {
			b.blocks = make([]byte, blocksSize)
		}

		valueCount := min(b.valueCount-b.ord, b.blockSize)
		blocksCount := FormatPacked.ByteCount(b.packedIntsVersion, valueCount, bitsPerValue)

		if _, err := b.in.Read(b.blocks[:blocksCount]); err != nil {
			return err
		}

		decoder.DecodeBytes(b.blocks, b.values, iterations)

		if minValue != 0 {
			for i := 0; i < valueCount; i++ {
				b.values[i] = uint64(int64(b.values[i]) + minValue)
			}
		}
	}
	b.off = 0
	return nil
}

// Ord
// Return the offset of the next value to read.
func (b *BlockPackedReaderIterator) Ord() int {
	return b.ord
}
