package packed

import (
	"github.com/geange/lucene-go/core/store"
)

var _ Reader = &DirectPacked64SingleBlockReader{}

type DirectPacked64SingleBlockReader struct {
	in             store.IndexInput
	bitsPerValue   int
	startPointer   int64
	valuesPerBlock int
	mask           uint64
}

func (d *DirectPacked64SingleBlockReader) GetBulk(index int, arr []uint64) int {
	//TODO implement me
	panic("implement me")
}

func (d *DirectPacked64SingleBlockReader) Size() int {
	//TODO implement me
	panic("implement me")
}

func (d *DirectPacked64SingleBlockReader) Get(index int) (uint64, error) {
	//TODO implement me
	panic("implement me")
}

func NewDirectPacked64SingleBlockReader(bitsPerValue, valueCount int,
	in store.IndexInput) *DirectPacked64SingleBlockReader {

	panic("")
}
