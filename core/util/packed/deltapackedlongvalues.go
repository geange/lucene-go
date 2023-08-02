package packed

import "slices"

type DeltaPackedLongValues struct {
	*PackedLongValues

	mins []int64
}

func NewDeltaPackedLongValues(pageShift int, pageMask uint64, values []Reader, mins []int64, size int) *DeltaPackedLongValues {
	return &DeltaPackedLongValues{
		PackedLongValues: NewPackedLongValues(values, pageShift, pageMask, size),
		mins:             mins,
	}
}

func (d *DeltaPackedLongValues) Get(index int) (uint64, error) {
	return d.getWithFnGet(index, d.get)
}

func (d *DeltaPackedLongValues) get(block int, element int) (uint64, error) {
	value, err := d.values[block].Get(element)
	if err != nil {
		return 0, err
	}
	return uint64(d.mins[block] + int64(value)), nil
}

func (d *DeltaPackedLongValues) Iterator() PackedLongValuesIterator {
	return d.iteratorWithSPI(d)
}

type DeltaPackedLongValuesBuilder struct {
	*PackedLongValuesBuilder
	mins []int64
}

func NewDeltaPackedLongValuesBuilder(pageSize int, acceptableOverheadRatio float64) *DeltaPackedLongValuesBuilder {
	return &DeltaPackedLongValuesBuilder{
		PackedLongValuesBuilder: NewPackedLongValuesBuilder(pageSize, acceptableOverheadRatio),
		mins:                    make([]int64, 0),
	}
}

func (d *DeltaPackedLongValuesBuilder) Add(value int64) error {
	return d.addWithFnPack(value, d.pack)
}

func (d *DeltaPackedLongValuesBuilder) finish() error {
	return d.finishWithFnPack(d.pack)
}

func (d *DeltaPackedLongValuesBuilder) pack() error {
	return d.packWithFnPackValues(d.packValues)
}

func (d *DeltaPackedLongValuesBuilder) packValues(values []int64, numValues int, acceptableOverheadRatio float64) error {
	var minValue int64
	for _, v := range values {
		minValue = min(minValue, v)
	}

	for i := range values {
		values[i] -= minValue
	}
	if err := d.PackedLongValuesBuilder.packValues(values, numValues, acceptableOverheadRatio); err != nil {
		return err
	}
	d.mins = append(d.mins, minValue)
	return nil
}

func (d *DeltaPackedLongValuesBuilder) Build() (*DeltaPackedLongValues, error) {
	if err := d.finish(); err != nil {
		return nil, err
	}

	d.pending = nil
	values := slices.Clone(d.values)
	mins := slices.Clone(d.mins)
	return NewDeltaPackedLongValues(d.pageShift, d.pageMask, values, mins, d.size), nil
}
