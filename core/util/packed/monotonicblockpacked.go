package packed

import (
	"context"
	"math"

	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
)

var _ types.LongValues = &MonotonicBlockPackedReader{}

// MonotonicBlockPackedReader
// Provides random access to a stream written with MonotonicBlockPackedWriter.
// lucene.internal
type MonotonicBlockPackedReader struct {
	blockShift int
	blockMask  uint64
	valueCount int
	minValues  []uint64
	averages   []float32
	subReaders []Reader
	sumBPV     uint64
}

// NewMonotonicBlockPackedReader
// IndexInput in, int packedIntsVersion, int blockSize, long valueCount, boolean direct
func NewMonotonicBlockPackedReader(ctx context.Context, in store.IndexInput,
	packedIntsVersion, blockSize, valueCount int, direct bool) (*MonotonicBlockPackedReader, error) {

	numBlocks, err := getNumBlocks(valueCount, blockSize)
	if err != nil {
		return nil, err
	}

	reader := &MonotonicBlockPackedReader{
		valueCount: valueCount,
		blockShift: checkBlockSize(blockSize, MIN_BLOCK_SIZE, ABP_MAX_BLOCK_SIZE),
		blockMask:  uint64(blockSize - 1),
		minValues:  make([]uint64, numBlocks),
		averages:   make([]float32, numBlocks),
		subReaders: make([]Reader, numBlocks),
	}

	sumBPV := 0

	for i := 0; i < numBlocks; i++ {
		minValue, err := in.ReadZInt64(ctx)
		if err != nil {
			return nil, err
		}
		reader.minValues[i] = uint64(minValue)

		u32, err := in.ReadUint32(ctx)
		if err != nil {
			return nil, err
		}
		average := math.Float32frombits(u32)
		reader.averages[i] = average

		bitsPerValue, err := in.ReadUvarint(ctx)
		if err != nil {
			return nil, err
		}
		if bitsPerValue == 0 {
			reader.subReaders[i] = NewNullReader(blockSize)
		} else {
			size := min(blockSize, valueCount-i*blockSize)
			if direct {
				pointer := in.GetFilePointer()
				readerNoHeader, err := getDirectReaderNoHeader(ctx, in, FormatPacked, packedIntsVersion, size, int(bitsPerValue))
				if err != nil {
					return nil, err
				}
				reader.subReaders[i] = readerNoHeader
				if _, err := in.Seek(
					pointer+int64(FormatPacked.ByteCount(packedIntsVersion, size, int(bitsPerValue))), 0); err != nil {
					return nil, err
				}
			} else {
				readerNoHeader, err := getReaderNoHeader(ctx, in, FormatPacked, packedIntsVersion, size, int(bitsPerValue))
				if err != nil {
					return nil, err
				}
				reader.subReaders[i] = readerNoHeader
			}
		}
	}

	reader.sumBPV = uint64(sumBPV)
	return reader, nil
}

func (m *MonotonicBlockPackedReader) Get(index int) (uint64, error) {
	block := index >> m.blockShift
	idx := int(uint64(index) & m.blockMask)

	value, err := m.subReaders[block].Get(idx)
	if err != nil {
		return 0, err
	}

	return expected(m.minValues[block], m.averages[block], idx) + value, nil
}

// Size
// Returns the number of values
func (m *MonotonicBlockPackedReader) Size() int {
	return m.valueCount
}

func expected(origin uint64, average float32, index int) uint64 {
	return origin + uint64(average*float32(index))
}

var _ BlockPackedFlusher = &MonotonicBlockPackedWriter{}

// MonotonicBlockPackedWriter
// A writer for large monotonically increasing sequences of positive longs.
// The sequence is divided into fixed-size blocks and for each block, values are modeled after a linear function f: x → A × x + B. The block encodes deltas from the expected values computed from this function using as few bits as possible.
// Format:
// <BLock>BlockCount
// BlockCount: ⌈ ValueCount / BlockSize ⌉
// Block: <Header, (Ints)>
// Header: <B, A, BitsPerValue>
// B: the B from f: x → A × x + B using a zig-zag encoded vLong
// A: the A from f: x → A × x + B encoded using Float.floatToIntBits(float) on 4 bytes
// BitsPerValue: a variable-length int
// Ints: if BitsPerValue is 0, then there is nothing to read and all values perfectly match the result of the function. Otherwise, these are the packed deltas from the expected value (computed from the function) using exactly BitsPerValue bits per value.
// 请参阅:
// MonotonicBlockPackedReader
// lucene.internal
type MonotonicBlockPackedWriter struct {
	*AbstractBlockPackedWriter
}

func NewMonotonicBlockPackedWriter(out store.DataOutput, blockSize int) *MonotonicBlockPackedWriter {
	writer := newAbstractBlockPackedWriter(out, blockSize)
	res := &MonotonicBlockPackedWriter{writer}
	writer.flusher = res
	return res
}

func (m *MonotonicBlockPackedWriter) Add(ctx context.Context, v uint64) error {
	return m.AbstractBlockPackedWriter.Add(ctx, v)
}

func (m *MonotonicBlockPackedWriter) Flush(ctx context.Context) error {
	var avg float32
	if m.off == 1 {
		avg = 0
	} else {
		avg = float32(m.values[m.off-1]-m.values[0]) / float32(m.off-1)
	}
	minValue := m.values[0]

	// adjust min so that all deltas will be positive
	for i := 1; i < m.off; i++ {
		actual := m.values[i]
		expect := expected(minValue, avg, i)
		if expect > actual {
			minValue = minValue - (expect - actual)
		}
	}

	maxDelta := uint64(0)
	for i := 0; i < m.off; i++ {
		m.values[i] = m.values[i] - expected(minValue, avg, i)
		maxDelta = max(maxDelta, m.values[i])
	}

	if err := m.out.WriteZInt64(ctx, int64(minValue)); err != nil {
		return err
	}
	if err := m.out.WriteUint32(ctx, math.Float32bits(float32(avg))); err != nil {
		return err
	}
	if maxDelta == 0 {
		if err := m.out.WriteUvarint(ctx, 0); err != nil {
			return err
		}
	} else {
		bitsRequired, err := BitsRequired(int64(maxDelta))
		if err != nil {
			return err
		}
		if err := m.out.WriteUvarint(ctx, uint64(bitsRequired)); err != nil {
			return err
		}
		if err := m.writeValues(bitsRequired); err != nil {
			return err
		}
	}

	m.off = 0

	return nil
}
