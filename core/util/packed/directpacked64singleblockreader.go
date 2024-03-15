package packed

import (
	"context"
	"errors"
	"github.com/geange/lucene-go/core/store"
	"io"
)

var _ Reader = &DirectPacked64SingleBlockReader{}

type DirectPacked64SingleBlockReader struct {
	in             store.IndexInput
	bitsPerValue   int
	valueCount     int
	startPointer   int64
	valuesPerBlock int
	mask           uint64
}

func (d *DirectPacked64SingleBlockReader) GetBulk(index int, arr []uint64) int {
	for i := 0; i < len(arr); i++ {
		idx := index + i
		n, err := d.Get(idx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return i
			}
			return -1
		}
		arr[i] = n
	}
	return len(arr)
}

func (d *DirectPacked64SingleBlockReader) Size() int {
	//TODO implement me
	return d.valueCount
}

func (d *DirectPacked64SingleBlockReader) Get(index int) (uint64, error) {
	blockOffset := index / d.valuesPerBlock
	skip := uint64(blockOffset) << 3

	if _, err := d.in.Seek(d.startPointer+int64(skip), io.SeekStart); err != nil {
		return 0, err
	}

	block, err := d.in.ReadUint64(context.TODO())
	if err != nil {
		return 0, err
	}

	offsetInBlock := index % d.valuesPerBlock
	return (block >> (offsetInBlock * d.bitsPerValue)) & d.mask, nil
}

func NewDirectPacked64SingleBlockReader(bitsPerValue, valueCount int,
	in store.IndexInput) *DirectPacked64SingleBlockReader {

	return &DirectPacked64SingleBlockReader{
		in:             in,
		bitsPerValue:   bitsPerValue,
		valueCount:     valueCount,
		startPointer:   in.GetFilePointer(),
		valuesPerBlock: 64 / bitsPerValue,
		mask:           ^(^uint64(0) << bitsPerValue),
	}
}
