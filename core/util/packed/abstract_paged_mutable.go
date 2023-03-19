package packed

import (
	. "github.com/geange/lucene-go/math"
)

type AbstractPagedMutable interface {
	Get(index int) uint64
	Set(index int, value uint64)
	Resize(newSize int) AbstractPagedMutable
	Grow(minSize int) AbstractPagedMutable
	GrowOne() AbstractPagedMutable
	SubMutables() []Mutable
	GetSubMutableByIndex(index int) Mutable
	SetSubMutableByIndex(index int, value Mutable)
}

// AbstractPagedMutableDefault Base implementation for PagedMutable and PagedGrowableWriter.
// lucene.internal
type AbstractPagedMutableDefault struct {
	spi          AbstractPagedMutableSPI
	size         int
	pageShift    int
	pageMask     int
	subMutables  []Mutable
	bitsPerValue int
}

/**
  AbstractPagedMutableDefault(int bitsPerValue, long size, int pageSize) {
    this.bitsPerValue = bitsPerValue;
    this.size = size;
    pageShift = checkBlockSize(pageSize, MIN_BLOCK_SIZE, MAX_BLOCK_SIZE);
    pageMask = pageSize - 1;
    final int numPages = numBlocks(size, pageSize);
    subMutables = new PackedInts.Mutable[numPages];
  }
*/

func newAbstractPagedMutable(spi AbstractPagedMutableSPI, bitsPerValue, size, pageSize int) *AbstractPagedMutableDefault {

	mutable := &AbstractPagedMutableDefault{
		spi:          spi,
		bitsPerValue: bitsPerValue,
		size:         size,
		pageShift:    checkBlockSize(pageSize, MIN_BLOCK_SIZE, MAX_BLOCK_SIZE),
		pageMask:     pageSize - 1,
	}

	numPages, _ := numBlocks(size, pageSize)
	mutable.subMutables = make([]Mutable, numPages)
	return mutable
}

type AbstractPagedMutableSPI interface {
	NewMutable(valueCount, bitsPerValue int) Mutable
	NewUnfilledCopy(newSize int) AbstractPagedMutable
}

const (
	MIN_BLOCK_SIZE = 1 << 6
	MAX_BLOCK_SIZE = 1 << 30
)

func (a *AbstractPagedMutableDefault) fillPages() error {
	numPages, err := numBlocks(a.size, a.pageSize())
	if err != nil {
		return err
	}
	for i := 0; i < numPages; i++ {
		valueCount := a.pageSize()
		if i == numPages-1 {
			// do not allocate for more entries than necessary on the last page
			valueCount = a.lastPageSize(a.size)
		}
		a.subMutables[i] = a.spi.NewMutable(valueCount, a.bitsPerValue)
	}
	return nil
}

func (a *AbstractPagedMutableDefault) lastPageSize(size int) int {
	sz := a.indexInPage(size)
	if sz == 0 {
		return a.pageSize()
	}
	return sz
}

func (a *AbstractPagedMutableDefault) pageSize() int {
	return a.pageMask + 1
}

func (a *AbstractPagedMutableDefault) Size() int {
	return a.size
}

func (a *AbstractPagedMutableDefault) pageIndex(index int) int {
	return index >> a.pageShift
}

func (a *AbstractPagedMutableDefault) indexInPage(index int) int {
	return index & a.pageMask
}

func (a *AbstractPagedMutableDefault) Get(index int) uint64 {
	pageIndex := a.pageIndex(index)
	indexInPage := a.indexInPage(index)
	return a.subMutables[pageIndex].Get(indexInPage)
}

func (a *AbstractPagedMutableDefault) Set(index int, value uint64) {
	pageIndex := a.pageIndex(index)
	indexInPage := a.indexInPage(index)
	a.subMutables[pageIndex].Set(indexInPage, value)
}

func (a *AbstractPagedMutableDefault) SubMutables() []Mutable {
	return a.subMutables
}

func (a *AbstractPagedMutableDefault) GetSubMutableByIndex(index int) Mutable {
	return a.subMutables[index]
}

func (a *AbstractPagedMutableDefault) SetSubMutableByIndex(index int, value Mutable) {
	a.subMutables[index] = value
}

// Resize Create a new copy of size newSize based on the content of this buffer.
// This method is much more efficient than creating a new instance and copying values one by one.
func (a *AbstractPagedMutableDefault) Resize(newSize int) AbstractPagedMutable {
	ucopy := a.spi.NewUnfilledCopy(newSize)
	numCommonPages := Min(len(ucopy.SubMutables()), len(a.subMutables))
	copyBuffer := make([]uint64, 1024)

	size := len(ucopy.SubMutables())

	subMutables := ucopy.SubMutables()
	for i := range subMutables {
		valueCount := a.pageSize()
		if i < size-1 {
			valueCount = a.lastPageSize(newSize)
		}

		bpv := a.bitsPerValue
		if i < numCommonPages {
			bpv = a.subMutables[i].GetBitsPerValue()
		}

		ucopy.SetSubMutableByIndex(i, a.spi.NewMutable(valueCount, bpv))
		if i < numCommonPages {
			copyLength := Min(valueCount, a.subMutables[i].Size())
			PackedIntsCopyBuff(a.subMutables[i], 0, subMutables[i], 0, copyLength, copyBuffer)
		}
	}
	return ucopy
}

func (a *AbstractPagedMutableDefault) Grow(minSize int) AbstractPagedMutable {
	if minSize <= a.Size() {
		return a
	}

	extra := minSize >> 3
	if extra < 3 {
		extra = 3
	}
	newSize := minSize + extra
	return a.Resize(newSize)
}

// GrowOne Similar to ArrayUtil.grow(long[]).
func (a *AbstractPagedMutableDefault) GrowOne() AbstractPagedMutable {
	return a.Grow(a.Size() + 1)
}
