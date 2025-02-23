package compressing

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/geange/lucene-go/core/store"
)

func TestRWZFloat(t *testing.T) {
	out := store.NewBufferDataOutput()

	nums := []float32{
		3.2,
		1.0,
		1000,
	}

	for _, num := range nums {
		err := WriteZFloat(context.Background(), out, num)
		assert.Nil(t, err)
	}
	bs := out.Bytes()

	in := store.NewBytesDataInput(bs)

	for _, num := range nums {
		actNum, err := ReadZFloat(context.Background(), in)
		assert.Nil(t, err)
		assert.Equal(t, num, actNum)
	}
}

func TestRWZDouble(t *testing.T) {
	out := store.NewBufferDataOutput()

	nums := []float64{
		3.2,
		1.0,
		1000,
		-1,
		-2,
		-100,
		-10e100 - 1000,
	}

	for _, num := range nums {
		err := WriteZDouble(context.Background(), out, num)
		assert.Nil(t, err)
	}
	bs := out.Bytes()

	in := store.NewBytesDataInput(bs)

	for _, num := range nums {
		actNum, err := ReadZDouble(context.Background(), in)
		assert.Nil(t, err)
		assert.Equal(t, num, actNum)
	}
}

func TestRWTLong(t *testing.T) {
	out := store.NewBufferDataOutput()

	nums := []int64{
		1735660800_000,
		1735664459_000,
		1735664460_000,
		1735664400_000,
		86400000_1000,
	}

	for _, num := range nums {
		err := WriteTLong(context.Background(), out, num)
		assert.Nil(t, err)
	}
	bs := out.Bytes()

	in := store.NewBytesDataInput(bs)

	for _, num := range nums {
		actNum, err := ReadTLong(context.Background(), in)
		assert.Nil(t, err)
		assert.Equal(t, num, actNum)
	}
}
