package packed

type DeltaPackedLongValues struct {
	LongValues

	mins []uint64
}

func (d *DeltaPackedLongValues) Get(block int, element int) (uint64, error) {
	value, err := d.values[block].Get(element)
	if err != nil {
		return 0, err
	}
	return d.mins[block] + value, nil
}
