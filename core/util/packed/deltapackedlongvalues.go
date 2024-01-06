package packed

type DeltaPackedLongValues struct {
	LongValues

	mins []int64
}

func (d *DeltaPackedLongValues) Get(block int, element int) int64 {
	return d.mins[block] + d.values[block].Get(element)
}
