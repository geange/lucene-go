package packed

type PagedMutable interface {
	Get(index int) (uint64, error)
	GetTest(index int) uint64
	Set(index int, value uint64)
	Resize(newSize int) PagedMutable
	Grow(minSize int) PagedMutable
	GrowOne() PagedMutable
	SubMutables() []Mutable
	GetSubMutableByIndex(index int) Mutable
	SetSubMutableByIndex(index int, value Mutable)
}

// basePagedMutable
// Base implementation for FixSizePagedMutable and PagedGrowableWriter.
// lucene.internal
type basePagedMutable struct {
	spi          PagedMutableBuilder
	size         int
	pageShift    int
	pageMask     int
	subMutables  []Mutable
	bitsPerValue int
}

func newPagedMutable(spi PagedMutableBuilder, bitsPerValue, size, pageSize int) (*basePagedMutable, error) {

	m := &basePagedMutable{
		spi:          spi,
		bitsPerValue: bitsPerValue,
		size:         size,
		pageShift:    checkBlockSize(pageSize, MIN_BLOCK_SIZE, MAX_BLOCK_SIZE),
		pageMask:     pageSize - 1,
	}

	numPages, err := getNumBlocks(size, pageSize)
	if err != nil {
		return nil, err
	}
	m.subMutables = make([]Mutable, numPages)
	return m, nil
}

type PagedMutableBuilder interface {
	NewMutable(valueCount, bitsPerValue int) Mutable
	NewUnfilledCopy(newSize int) (PagedMutable, error)
}

const (
	MIN_BLOCK_SIZE = 1 << 6
	MAX_BLOCK_SIZE = 1 << 30
)

func (a *basePagedMutable) fillPages() error {
	numPages, err := getNumBlocks(a.size, a.pageSize())
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

func (a *basePagedMutable) lastPageSize(size int) int {
	sz := a.indexInPage(size)
	if sz == 0 {
		return a.pageSize()
	}
	return sz
}

func (a *basePagedMutable) pageSize() int {
	return a.pageMask + 1
}

func (a *basePagedMutable) Size() int {
	return a.size
}

func (a *basePagedMutable) pageIndex(index int) int {
	return index >> a.pageShift
}

func (a *basePagedMutable) indexInPage(index int) int {
	return index & a.pageMask
}

func (a *basePagedMutable) Get(index int) (uint64, error) {
	pageIndex := a.pageIndex(index)
	indexInPage := a.indexInPage(index)
	return a.subMutables[pageIndex].Get(indexInPage)
}

func (a *basePagedMutable) GetTest(index int) uint64 {
	v, _ := a.Get(index)
	return v
}

func (a *basePagedMutable) Set(index int, value uint64) {
	pageIndex := a.pageIndex(index)
	indexInPage := a.indexInPage(index)
	a.subMutables[pageIndex].Set(indexInPage, value)
}

func (a *basePagedMutable) SubMutables() []Mutable {
	return a.subMutables
}

func (a *basePagedMutable) GetSubMutableByIndex(index int) Mutable {
	return a.subMutables[index]
}

func (a *basePagedMutable) SetSubMutableByIndex(index int, value Mutable) {
	a.subMutables[index] = value
}

// Resize
// Create a new copy of size newSize based on the content of this buffer.
// This method is much more efficient than creating a new instance and copying values one by one.
func (a *basePagedMutable) Resize(newSize int) PagedMutable {
	ucopy, _ := a.spi.NewUnfilledCopy(newSize)
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
			CopyValuesWithBuffer(a.subMutables[i], 0, subMutables[i], 0, copyLength, copyBuffer)
		}
	}
	return ucopy
}

func (a *basePagedMutable) Grow(minSize int) PagedMutable {
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
func (a *basePagedMutable) GrowOne() PagedMutable {
	return a.Grow(a.Size() + 1)
}
