package packed

import "github.com/geange/lucene-go/math"

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

// abstractPagedMutable Base implementation for PagedMutable and PagedGrowableWriter.
// lucene.internal
type abstractPagedMutable struct {
	spi          abstractPagedMutableSPI
	size         int
	pageShift    int
	pageMask     int
	subMutables  []Mutable
	bitsPerValue int
}

/**
  abstractPagedMutable(int bitsPerValue, long size, int pageSize) {
    this.bitsPerValue = bitsPerValue;
    this.size = size;
    pageShift = checkBlockSize(pageSize, MIN_BLOCK_SIZE, MAX_BLOCK_SIZE);
    pageMask = pageSize - 1;
    final int numPages = numBlocks(size, pageSize);
    subMutables = new PackedInts.Mutable[numPages];
  }
*/

func newAbstractPagedMutable(bitsPerValue, size, pageSize int) *abstractPagedMutable {

	mutable := &abstractPagedMutable{
		bitsPerValue: bitsPerValue,
		size:         size,
		pageShift:    checkBlockSize(pageSize, MIN_BLOCK_SIZE, MAX_BLOCK_SIZE),
		pageMask:     pageSize - 1,
	}

	numPages, _ := numBlocks(size, pageSize)
	mutable.subMutables = make([]Mutable, numPages)
	return mutable
}

type abstractPagedMutableSPI interface {
	newMutable(valueCount, bitsPerValue int) Mutable
	newUnfilledCopy(newSize int) AbstractPagedMutable
}

const (
	MIN_BLOCK_SIZE = 1 << 6
	MAX_BLOCK_SIZE = 1 << 30
)

func (a *abstractPagedMutable) fillPages() error {
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
		a.subMutables[i] = a.spi.newMutable(valueCount, a.bitsPerValue)
	}
	return nil
}

func (a *abstractPagedMutable) lastPageSize(size int) int {
	sz := a.indexInPage(size)
	if sz == 0 {
		return a.pageSize()
	}
	return sz
}

func (a *abstractPagedMutable) pageSize() int {
	return a.pageMask + 1
}

func (a *abstractPagedMutable) Size() int {
	return a.size
}

func (a *abstractPagedMutable) pageIndex(index int) int {
	return index >> a.pageShift
}

func (a *abstractPagedMutable) indexInPage(index int) int {
	return index & a.pageMask
}

func (a *abstractPagedMutable) Get(index int) uint64 {
	pageIndex := a.pageIndex(index)
	indexInPage := a.indexInPage(index)
	return a.subMutables[pageIndex].Get(indexInPage)
}

func (a *abstractPagedMutable) Set(index int, value uint64) {
	pageIndex := a.pageIndex(index)
	indexInPage := a.indexInPage(index)
	a.subMutables[pageIndex].Set(indexInPage, value)
}

func (a *abstractPagedMutable) SubMutables() []Mutable {
	return a.subMutables
}

func (a *abstractPagedMutable) GetSubMutableByIndex(index int) Mutable {
	return a.subMutables[index]
}

func (a *abstractPagedMutable) SetSubMutableByIndex(index int, value Mutable) {
	a.subMutables[index] = value
}

// Resize Create a new copy of size newSize based on the content of this buffer.
// This method is much more efficient than creating a new instance and copying values one by one.
func (a *abstractPagedMutable) Resize(newSize int) AbstractPagedMutable {
	ucopy := a.spi.newUnfilledCopy(newSize)
	numCommonPages := math.Min(len(ucopy.SubMutables()), len(a.subMutables))
	copyBuffer := make([]uint64, 1024)

	size := len(ucopy.SubMutables())

	for i, mutable := range ucopy.SubMutables() {
		valueCount := a.pageSize()
		if i < size-1 {
			valueCount = a.lastPageSize(newSize)
		}

		bpv := a.bitsPerValue
		if i < numCommonPages {
			bpv = a.subMutables[i].GetBitsPerValue()
		}

		ucopy.SetSubMutableByIndex(i, a.spi.newMutable(valueCount, bpv))
		if i < numCommonPages {
			copyLength := math.Min(valueCount, a.subMutables[i].Size())
			PackedIntsCopyBuff(a.subMutables[i], 0, mutable, 0, copyLength, copyBuffer)
		}
	}
	return ucopy
}

func (a *abstractPagedMutable) Grow(minSize int) AbstractPagedMutable {
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
func (a *abstractPagedMutable) GrowOne() AbstractPagedMutable {
	return a.Grow(a.Size() + 1)
}
