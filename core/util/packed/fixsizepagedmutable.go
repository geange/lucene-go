package packed

var _ PagedMutableBuilder = &FixSizePagedMutable{}

type FixSizePagedMutableBuilder struct {
}

func NewFixSizePagedMutableBuilder() FixSizePagedMutableBuilder {
	return FixSizePagedMutableBuilder{}
}

func (b *FixSizePagedMutableBuilder) New(size, pageSize, bitsPerValue int, acceptableOverheadRatio float64) (*FixSizePagedMutable, error) {
	m, err := b.NewWithFormatAndBits(size, pageSize, fastestFormatAndBits(pageSize, bitsPerValue, acceptableOverheadRatio))
	if err != nil {
		return nil, err
	}

	if err := m.fillPages(); err != nil {
		return nil, err
	}
	return m, nil
}

func (b *FixSizePagedMutableBuilder) NewWithFormatAndBits(size, pageSize int, formatAndBits *FormatAndBits) (*FixSizePagedMutable, error) {
	return b.NewWithFormat(size, pageSize, formatAndBits.bitsPerValue, formatAndBits.format)
}

func (b *FixSizePagedMutableBuilder) NewWithFormat(size, pageSize, bitsPerValue int, format Format) (*FixSizePagedMutable, error) {
	m := &FixSizePagedMutable{
		format:  format,
		builder: b,
	}

	pagedMutable, err := newPagedMutable(m, bitsPerValue, size, pageSize)
	if err != nil {
		return nil, err
	}
	m.basePagedMutable = pagedMutable
	return m, nil
}

// FixSizePagedMutable
// A FixSizePagedMutable. This class slices data into fixed-size blocks which have the same number of bits per value.
// It can be a useful replacement for PackedInts.Mutable to store more than 2B values.
type FixSizePagedMutable struct {
	*basePagedMutable

	format  Format
	builder *FixSizePagedMutableBuilder
}

func (p *FixSizePagedMutable) NewMutable(valueCount, bitsPerValue int) Mutable {
	return getMutable(valueCount, p.bitsPerValue, p.format)
}

func (p *FixSizePagedMutable) NewUnfilledCopy(newSize int) (PagedMutable, error) {
	return p.builder.NewWithFormat(newSize, p.pageSize(), p.bitsPerValue, p.format)
}
