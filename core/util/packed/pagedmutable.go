package packed

var _ AbstractPagedMutableSPI = &PagedMutable{}

// PagedMutable
// A PagedMutable. This class slices data into fixed-size blocks which have the same number of bits per value.
// It can be a useful replacement for PackedInts.Mutable to store more than 2B values.
type PagedMutable struct {
	*BaseAbstractPagedMutable

	format Format
}

func NewPagedMutable(size int64, pageSize, bitsPerValue int,
	acceptableOverheadRatio float64) *PagedMutable {

	m := NewPagedMutableV1(size, pageSize,
		fastestFormatAndBits(pageSize, bitsPerValue, acceptableOverheadRatio))
	m.fillPages()
	return m
}

func NewPagedMutableV1(size int64, pageSize int, formatAndBits *FormatAndBits) *PagedMutable {
	return NewPagedMutableV3(size, pageSize, formatAndBits.bitsPerValue, formatAndBits.format)
}

func NewPagedMutableV3(size int64, pageSize, bitsPerValue int,
	format Format) *PagedMutable {
	m := &PagedMutable{
		format: format,
	}
	m.BaseAbstractPagedMutable = newAbstractPagedMutable(m, bitsPerValue, int(size), pageSize)
	return m
}

func (p *PagedMutable) NewMutable(valueCount, bitsPerValue int) Mutable {
	return getMutable(valueCount, p.bitsPerValue, p.format)
}

func (p *PagedMutable) NewUnfilledCopy(newSize int) AbstractPagedMutable {
	return NewPagedMutableV3(int64(newSize), p.pageSize(), p.bitsPerValue, p.format)
}
