package packed

import "slices"

type MonotonicLongValues struct {
	*DeltaPackedLongValues

	averages []float32
}

func NewMonotonicLongValues(pageShift int, pageMask uint64,
	values []Reader, mins []int64, averages []float32, size int) *MonotonicLongValues {

	return &MonotonicLongValues{
		DeltaPackedLongValues: NewDeltaPackedLongValues(pageShift, pageMask, values, mins, size),
		averages:              averages,
	}
}

func (m *MonotonicLongValues) Get(index int) (uint64, error) {
	return m.getWithFnGet(index, m.get)
}

func (m *MonotonicLongValues) Iterator() PackedLongValuesIterator {
	return m.iteratorWithSPI(m)
}

func (m *MonotonicLongValues) get(block int, element int) (uint64, error) {
	value, err := m.values[block].Get(element)
	if err != nil {
		return 0, err
	}

	add := expected(m.mins[block], m.averages[block], element)
	return uint64(add + int64(value)), nil
}

type MonotonicLongValuesBuilder struct {
	*DeltaPackedLongValuesBuilder

	averages []float32
}

func NewMonotonicLongValuesBuilder(pageSize int, acceptableOverheadRatio float64) *MonotonicLongValuesBuilder {
	return &MonotonicLongValuesBuilder{
		DeltaPackedLongValuesBuilder: NewDeltaPackedLongValuesBuilder(pageSize, acceptableOverheadRatio),
		averages:                     make([]float32, 0),
	}
}

func (m *MonotonicLongValuesBuilder) Build() (*MonotonicLongValues, error) {
	if err := m.finish(); err != nil {
		return nil, err
	}

	m.pending = nil
	values := slices.Clone(m.values)
	mins := slices.Clone(m.mins)
	averages := slices.Clone(m.averages)
	return NewMonotonicLongValues(m.pageShift, m.pageMask, values, mins, averages, m.size), nil
}

func (m *MonotonicLongValuesBuilder) Add(value int64) error {
	return m.addWithFnPack(value, m.pack)
}

func (m *MonotonicLongValuesBuilder) finish() error {
	return m.finishWithFnPack(m.pack)
}

func (m *MonotonicLongValuesBuilder) pack() error {
	return m.packWithFnPackValues(m.packValues)
}

func (m *MonotonicLongValuesBuilder) packValues(values []int64, numValues int, acceptableOverheadRatio float64) error {
	var average float32 = 0
	if numValues != 1 {
		average = float32(values[numValues-1]-values[0]) / float32(numValues-1)
	}
	for i := 0; i < numValues; i++ {
		values[i] -= expected(0, average, i)
	}
	if err := m.DeltaPackedLongValuesBuilder.packValues(values, numValues, acceptableOverheadRatio); err != nil {
		return err
	}
	m.averages = append(m.averages, average)
	return nil
}
