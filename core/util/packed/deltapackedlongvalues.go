package packed

type DeltaPackedLongValues struct {
	LongValues

	mins []uint64
}

func (d *DeltaPackedLongValues) Get(block int, element int) uint64 {
	return d.mins[block] + d.values[block].Get(element)
}
