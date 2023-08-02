package packed

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPagedGrowableWriter(t *testing.T) {
	for i := 0; i < 10; i++ {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))

		size := 1 + r.Intn(1<<10)
		pageSize := 1 << (6 + rand.Intn(12))
		startBitsPerValue := 2 + r.Intn(20)
		acceptableOverheadRatio := 1.0

		writer, err := NewPagedGrowableWriter(size, pageSize, startBitsPerValue, acceptableOverheadRatio)
		assert.Nil(t, err)

		for j := 0; j < 100; j++ {
			index := r.Intn(size)
			value := uint64(r.Intn(1 << startBitsPerValue))

			writer.Set(index, value)
			assert.EqualValues(t, value, writer.GetTest(index))
		}
	}

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

	assert.EqualValues(t, 1, writer.GetTest(0))
	assert.EqualValues(t, 2, writer.GetTest(999))

	newWriter, err := writer.NewUnfilledCopy(100)
	assert.Nil(t, err)

	assert.Panics(t, func() {
		newWriter.Set(99, 3)
		newWriter.Set(999, 4)
	})
}

//func TestPagedGrowableWriter_Resize(t *testing.T) {
//	for i := 0; i < 10; i++ {
//		r := rand.New(rand.NewSource(time.Now().UnixNano()))
//
//		size := 1000 + r.Intn(1<<20)
//		pageSize := 1 << (6 + rand.Intn(12))
//		startBitsPerValue := 2 + r.Intn(20)
//		acceptableOverheadRatio := 1.0
//
//		writer, err := NewPagedGrowableWriter(size, pageSize, startBitsPerValue, acceptableOverheadRatio)
//		assert.Nil(t, err)
//
//		for j := 0; j < 100; j++ {
//			index := r.Intn(size)
//			value := uint64(r.Intn(1 << startBitsPerValue))
//
//			writer.Set(index, value)
//			assert.EqualValues(t, value, writer.GetTest(index))
//		}
//
//		newSize := size / 2
//		newWriter := writer.Resize(newSize)
//
//		newWriter.Set(newSize-1, 3)
//		assert.EqualValues(t, 1, newWriter.GetTest(0))
//		assert.EqualValues(t, 3, newWriter.GetTest(newSize-1))
//
//		//for j := 0; j < 100; j++ {
//		//	index := r.Intn(newSize)
//		//	value := uint64(r.Intn(1 << startBitsPerValue))
//		//
//		//	newWriter.Set(index, value)
//		//	assert.EqualValues(t, value, newWriter.GetTest(index))
//		//}
//
//	}
//}
