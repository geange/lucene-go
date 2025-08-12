package packed

import (
	"context"
	"errors"
	"math"

	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
)

// DirectMonotonicReader
// Retrieves an instance previously written by DirectMonotonicWriter.
type DirectMonotonicReader struct {
	blockShift  int
	readers     []types.LongValues
	mins        []int64
	avgs        []float32
	bpvs        []byte
	nonZeroBpvs int
}

// BinarySearch
// Return the index of a key if it exists, or its insertion point
func (d *DirectMonotonicReader) BinarySearch(fromIndex, toIndex, key int64) (int64, error) {
	if fromIndex < 0 || fromIndex > toIndex {
		return 0, errors.New("illegal argument exception")
	}
	lo := fromIndex
	hi := toIndex - 1

	for lo <= hi {
		mid := (lo + hi) >> 1
		// Try to run as many iterations of the binary search as possible without
		// hitting the direct readers, since they might hit a page fault.
		bounds := d.getBounds(int(mid))
		if bounds[1] < key {
			lo = mid + 1
		} else if bounds[0] > key {
			hi = mid - 1
		} else {
			midVal, err := d.Get(int(mid))
			if err != nil {
				return 0, err
			}
			if midVal < key {
				lo = mid + 1
			} else if midVal > key {
				hi = mid - 1
			} else {
				return mid, nil
			}
		}
	}

	return -1 - lo, nil
}

func (d *DirectMonotonicReader) Get(index int) (int64, error) {
	block := index >> d.blockShift
	blockIndex := index & ((1 << d.blockShift) - 1)
	delta, err := d.readers[block].Get(blockIndex)
	if err != nil {
		return 0, err
	}
	return d.mins[block] + int64(d.avgs[block]*float32(blockIndex)) + int64(delta), nil
}

// Get lower/ upper bounds for the value at a given index without hitting the direct reader.
func (d *DirectMonotonicReader) getBounds(index int) []int64 {
	block := int(index >> d.blockShift)
	blockIndex := index & ((1 << d.blockShift) - 1)
	lowerBound := d.mins[block] + int64(d.avgs[block]*float32(blockIndex))
	upperBound := lowerBound + int64(1<<d.bpvs[block]) - 1
	if d.bpvs[block] == 64 || upperBound < lowerBound { // overflow
		return []int64{math.MinInt64, math.MaxInt64}
	} else {
		return []int64{lowerBound, upperBound}
	}
}

type Meta struct {
	blockShift int
	numBlocks  int
	mins       []int64
	avgs       []float32
	bpvs       []byte
	offsets    []int64
}

func NewMeta(numValues int64, blockShift int) *Meta {
	meta := &Meta{
		blockShift: blockShift,
	}
	numBlocks := numValues >> blockShift
	if (numBlocks << blockShift) < numValues {
		numBlocks += 1
	}
	meta.numBlocks = int(numBlocks)
	meta.mins = make([]int64, numBlocks)
	meta.avgs = make([]float32, numBlocks)
	meta.bpvs = make([]byte, numBlocks)
	meta.offsets = make([]int64, numBlocks)
	return meta
}

// LoadMeta
// Load metadata from the given IndexInput.
func LoadMeta(metaIn store.IndexInput, numValues int64, blockShift int) (*Meta, error) {
	meta := NewMeta(numValues, blockShift)
	for i := 0; i < meta.numBlocks; i++ {
		u64, err := metaIn.ReadUint64(context.Background())
		if err != nil {
			return nil, err
		}
		meta.mins[i] = int64(u64)

		u32, err := metaIn.ReadUint32(context.Background())
		if err != nil {
			return nil, err
		}
		meta.avgs[i] = math.Float32frombits(u32)

		offsetU64, err := metaIn.ReadUint64(context.Background())
		if err != nil {
			return nil, err
		}
		meta.offsets[i] = int64(offsetU64)

		bpv, err := metaIn.ReadByte()
		if err != nil {
			return nil, err
		}
		meta.bpvs[i] = bpv
	}
	return meta, nil
}

var (
	_ types.LongValues = &emptyLongValues{}

	empty = &emptyLongValues{}
)

type emptyLongValues struct {
}

func (e *emptyLongValues) Get(index int) (uint64, error) {
	return 0, nil
}

func DirectMonotonicReaderGetInstance(meta *Meta, data store.RandomAccessInput) (*DirectMonotonicReader, error) {
	readers := make([]types.LongValues, meta.numBlocks)

	for i := 0; i < len(meta.mins); i++ {
		if meta.bpvs[i] == 0 {
			readers[i] = empty
		} else {
			instance, err := DirectReaderGetInstance(data, int(meta.bpvs[i]), int(meta.offsets[i]))
			if err != nil {
				return nil, err
			}
			readers[i] = instance
		}
	}

	return newDirectMonotonicReader(meta.blockShift, readers, meta.mins, meta.avgs, meta.bpvs)
}

func newDirectMonotonicReader(blockShift int, readers []types.LongValues,
	mins []int64, avgs []float32, bpvs []byte) (*DirectMonotonicReader, error) {

	if len(readers) != len(mins) || len(readers) != len(avgs) || len(readers) != len(bpvs) {
		return nil, errors.New("illegal argument exception")
	}

	nonZeroBpvs := 0
	for _, bpv := range bpvs {
		if bpv != 0 {
			nonZeroBpvs++
		}
	}

	reader := &DirectMonotonicReader{
		blockShift:  blockShift,
		readers:     readers,
		mins:        mins,
		avgs:        avgs,
		bpvs:        bpvs,
		nonZeroBpvs: nonZeroBpvs,
	}
	return reader, nil
}
