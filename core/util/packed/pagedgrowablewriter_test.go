package packed

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPagedGrowableWriter(t *testing.T) {

	size := 1000
	pageSize := 1 << 8
	startBitsPerValue := 2
	acceptableOverheadRatio := 1.0

	writer, err := NewPagedGrowableWriter(size, pageSize, startBitsPerValue, acceptableOverheadRatio)
	assert.Nil(t, err)

	writer.Set(0, 1)
	writer.Set(999, 2)

	assert.EqualValues(t, 1, writer.Get(0))
	assert.EqualValues(t, 2, writer.Get(999))
}

func TestPagedGrowableWriter_NewUnfilledCopy(t *testing.T) {
	size := 1000
	pageSize := 1 << 8
	startBitsPerValue := 2
	acceptableOverheadRatio := 1.0

	writer, err := NewPagedGrowableWriter(size, pageSize, startBitsPerValue, acceptableOverheadRatio)
	assert.Nil(t, err)

	writer.Set(0, 1)
	writer.Set(999, 2)

	assert.EqualValues(t, 1, writer.Get(0))
	assert.EqualValues(t, 2, writer.Get(999))

	newWriter := writer.NewUnfilledCopy(100)

	assert.Panics(t, func() {
		newWriter.Set(99, 3)
		newWriter.Set(999, 4)
	})
}

func TestPagedGrowableWriter_Resize(t *testing.T) {
	size := 1000
	pageSize := 1 << 8
	startBitsPerValue := 2
	acceptableOverheadRatio := 1.0

	writer, err := NewPagedGrowableWriter(size, pageSize, startBitsPerValue, acceptableOverheadRatio)
	assert.Nil(t, err)

	writer.Set(0, 1)
	writer.Set(999, 2)

	assert.EqualValues(t, 1, writer.Get(0))
	assert.EqualValues(t, 2, writer.Get(999))

	newWriter := writer.Resize(100)

	newWriter.Set(99, 3)
	assert.EqualValues(t, 1, newWriter.Get(0))
	assert.EqualValues(t, 3, newWriter.Get(99))
}
