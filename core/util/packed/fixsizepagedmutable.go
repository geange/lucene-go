package packed

var _ PagedMutableSPI = &FixSizePagedMutable{}

// FixSizePagedMutable
// A FixSizePagedMutable. This class slices data into fixed-size blocks which have the same number of bits per value.
// It can be a useful replacement for PackedInts.Mutable to store more than 2B values.
type FixSizePagedMutable struct {
	*BasePagedMutable

	format Format
}

func NewPagedMutable(size int64, pageSize, bitsPerValue int,
	acceptableOverheadRatio float64) (*FixSizePagedMutable, error) {

	m := NewPagedMutableV1(size, pageSize,
		fastestFormatAndBits(pageSize, bitsPerValue, acceptableOverheadRatio))
	if err := m.fillPages(); err != nil {
		return nil, err
	}
	return m, nil
}

func NewPagedMutableV1(size int64, pageSize int, formatAndBits *FormatAndBits) *FixSizePagedMutable {
	return NewPagedMutableV3(size, pageSize, formatAndBits.bitsPerValue, formatAndBits.format)
}

func NewPagedMutableV3(size int64, pageSize, bitsPerValue int,
	format Format) *FixSizePagedMutable {
	m := &FixSizePagedMutable{
		format: format,
	}
	m.BasePagedMutable = newPagedMutable(m, bitsPerValue, int(size), pageSize)
	return m
}

func (p *FixSizePagedMutable) NewMutable(valueCount, bitsPerValue int) Mutable {
	return getMutable(valueCount, p.bitsPerValue, p.format)
}

func (p *FixSizePagedMutable) NewUnfilledCopy(newSize int) PagedMutable {
	return NewPagedMutableV3(int64(newSize), p.pageSize(), p.bitsPerValue, p.format)
}
