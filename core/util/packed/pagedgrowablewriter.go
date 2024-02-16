package packed

var _ AbstractPagedMutableSPI = &PagedGrowableWriter{}

// PagedGrowableWriter A PagedGrowableWriter. This class slices data into fixed-size blocks which have
// independent numbers of bits per value and grow on-demand.
// You should use this class instead of the LongValues related ones only when you need random
// write-access. Otherwise this class will likely be slower and less memory-efficient.
// lucene.internal
type PagedGrowableWriter struct {
	*BaseAbstractPagedMutable

	acceptableOverheadRatio float64
}

func NewPagedGrowableWriter(size, pageSize, startBitsPerValue int, acceptableOverheadRatio float64) (*PagedGrowableWriter, error) {
	writer := &PagedGrowableWriter{
		BaseAbstractPagedMutable: nil,
		acceptableOverheadRatio:  acceptableOverheadRatio,
	}
	return writer.NewPagedGrowableWriter(size, pageSize, startBitsPerValue, acceptableOverheadRatio, true)
}

func (p *PagedGrowableWriter) NewPagedGrowableWriter(size, pageSize, startBitsPerValue int,
	acceptableOverheadRatio float64, fillPages bool) (*PagedGrowableWriter, error) {

	p.BaseAbstractPagedMutable = newAbstractPagedMutable(p, startBitsPerValue, size, pageSize)
	p.acceptableOverheadRatio = acceptableOverheadRatio
	if fillPages {
		if err := p.fillPages(); err != nil {
			return nil, err
		}
	}
	return p, nil
}

func (p *PagedGrowableWriter) NewMutable(valueCount, bitsPerValue int) Mutable {
	return NewGrowableWriter(bitsPerValue, valueCount, p.acceptableOverheadRatio)
}

func (p *PagedGrowableWriter) NewUnfilledCopy(newSize int) AbstractPagedMutable {
	writer, err := p.NewPagedGrowableWriter(newSize, p.pageSize(), p.bitsPerValue, p.acceptableOverheadRatio, false)
	if err != nil {
		return nil
	}
	return writer
}
