package packed

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

// BaseAbstractPagedMutable
// Base implementation for PagedMutable and PagedGrowableWriter.
// lucene.internal
type BaseAbstractPagedMutable struct {
	spi          AbstractPagedMutableSPI
	size         int
	pageShift    int
	pageMask     int
	subMutables  []Mutable
	bitsPerValue int
}

func newAbstractPagedMutable(spi AbstractPagedMutableSPI, bitsPerValue, size, pageSize int) *BaseAbstractPagedMutable {

	m := &BaseAbstractPagedMutable{
		spi:          spi,
		bitsPerValue: bitsPerValue,
		size:         size,
		pageShift:    checkBlockSize(pageSize, MIN_BLOCK_SIZE, MAX_BLOCK_SIZE),
		pageMask:     pageSize - 1,
	}

	numPages, _ := numBlocks(size, pageSize)
	m.subMutables = make([]Mutable, numPages)
	return m
}

type AbstractPagedMutableSPI interface {
	NewMutable(valueCount, bitsPerValue int) Mutable
	NewUnfilledCopy(newSize int) AbstractPagedMutable
}

const (
	MIN_BLOCK_SIZE = 1 << 6
	MAX_BLOCK_SIZE = 1 << 30
)

func (a *BaseAbstractPagedMutable) fillPages() error {
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

func (a *BaseAbstractPagedMutable) lastPageSize(size int) int {
	sz := a.indexInPage(size)
	if sz == 0 {
		return a.pageSize()
	}
	return sz
}

func (a *BaseAbstractPagedMutable) pageSize() int {
	return a.pageMask + 1
}

func (a *BaseAbstractPagedMutable) Size() int {
	return a.size
}

func (a *BaseAbstractPagedMutable) pageIndex(index int) int {
	return index >> a.pageShift
}

func (a *BaseAbstractPagedMutable) indexInPage(index int) int {
	return index & a.pageMask
}

func (a *BaseAbstractPagedMutable) Get(index int) uint64 {
	pageIndex := a.pageIndex(index)
	indexInPage := a.indexInPage(index)
	return a.subMutables[pageIndex].Get(indexInPage)
}

func (a *BaseAbstractPagedMutable) Set(index int, value uint64) {
	pageIndex := a.pageIndex(index)
	indexInPage := a.indexInPage(index)
	a.subMutables[pageIndex].Set(indexInPage, value)
}

func (a *BaseAbstractPagedMutable) SubMutables() []Mutable {
	return a.subMutables
}

func (a *BaseAbstractPagedMutable) GetSubMutableByIndex(index int) Mutable {
	return a.subMutables[index]
}

func (a *BaseAbstractPagedMutable) SetSubMutableByIndex(index int, value Mutable) {
	a.subMutables[index] = value
}

// Resize
// Create a new copy of size newSize based on the content of this buffer.
// This method is much more efficient than creating a new instance and copying values one by one.
func (a *BaseAbstractPagedMutable) Resize(newSize int) AbstractPagedMutable {
	ucopy := a.spi.NewUnfilledCopy(newSize)
	numCommonPages := min(len(ucopy.SubMutables()), len(a.subMutables))
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
			copyLength := min(valueCount, a.subMutables[i].Size())
			PackedIntsCopyBuff(a.subMutables[i], 0, subMutables[i], 0, copyLength, copyBuffer)
		}
	}
	return ucopy
}

func (a *BaseAbstractPagedMutable) Grow(minSize int) AbstractPagedMutable {
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
func (a *BaseAbstractPagedMutable) GrowOne() AbstractPagedMutable {
	return a.Grow(a.Size() + 1)
}
