package fst

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBytesEnumDoNext(t *testing.T) {
	builder, err := NewBuilder(BYTE1, NewBoxManager[int64]())
	assert.Nil(t, err)

	writeItems := []struct {
		key   string
		value int64
	}{
		{
			key:   "mop",
			value: 100,
		},
		{
			key:   "moth",
			value: 91,
		},
		{
			key:   "pop",
			value: 72,
		},
		{
			key:   "star",
			value: 83,
		},
		{
			key:   "stop",
			value: 54,
		},
		{
			key:   "top",
			value: 55,
		},
	}
	ctx := context.Background()
	for _, item := range writeItems {
		err = builder.Add(ctx, []rune(item.key), NewIntBox[int64](item.value))
		assert.Nil(t, err)
	}

	fst, err := builder.Finish(ctx)
	assert.Nil(t, err)

	fstEnum, err := NewEnum[byte](fst)
	assert.Nil(t, err)

	for _, item := range writeItems {
		next, err := fstEnum.Next(context.TODO())
		assert.Nil(t, err)
		assert.Equal(t, item.key, string(next.GetInput()))
		assert.Equal(t, item.value, next.GetOutput().(*IntBox[int64]).Value())
	}
}

func TestBytesEnumSeekExact(t *testing.T) {
	builder, err := NewBuilder(BYTE1, NewBoxManager[int64]())
	assert.Nil(t, err)

	writeItems := []struct {
		key   string
		value int64
	}{
		{
			key:   "mop",
			value: 100,
		},
		{
			key:   "moth",
			value: 91,
		},
		{
			key:   "pop",
			value: 72,
		},
		{
			key:   "star",
			value: 83,
		},
		{
			key:   "stop",
			value: 54,
		},
		{
			key:   "top",
			value: 55,
		},
	}
	ctx := context.Background()
	for _, item := range writeItems {
		err = builder.Add(ctx, []rune(item.key), NewIntBox[int64](item.value))
		assert.Nil(t, err)
	}

	fst, err := builder.Finish(ctx)
	assert.Nil(t, err)

	fstEnum, err := NewEnum[byte](fst)
	assert.Nil(t, err)

	items := []struct {
		key   string
		value int64
		exist bool
	}{
		{
			key:   "mop",
			value: 100,
			exist: true,
		},
		{
			key:   "moth",
			value: 91,
			exist: true,
		},
		{
			key:   "moth1",
			value: 91,
			exist: false,
		},
		{
			key:   "top",
			value: 55,
			exist: true,
		},
		{
			key:   "toz",
			value: 55,
			exist: false,
		},
	}

	for _, item := range items {
		next, ok, err := fstEnum.SeekExact(context.TODO(), []byte(item.key))
		assert.Nil(t, err)
		assert.Equal(t, item.exist, ok)
		if item.exist {
			assert.Equal(t, item.value, next.GetOutput().(*IntBox[int64]).Value())
		}
	}
}

func TestBytesEnumSeekExactArcsForBinarySearch(t *testing.T) {
	builder, err := NewBuilder(BYTE1, NewBoxManager[int64]())
	assert.Nil(t, err)

	writeItems := []struct {
		key   string
		value int64
	}{
		{
			key:   "m1",
			value: 100,
		},
		{
			key:   "m2",
			value: 91,
		},
		{
			key:   "m3",
			value: 72,
		},
		{
			key:   "m4",
			value: 83,
		},
		{
			key:   "m5",
			value: 54,
		},
		{
			key:   "m6",
			value: 55,
		},
		{
			key:   "m7",
			value: 55,
		},
		{
			key:   "m8",
			value: 67,
		},
		{
			key:   "m9",
			value: 59,
		},
		{

			key:   "ma",
			value: 12,
		},
		{
			key:   "mb",
			value: 13,
		},
		{
			key:   "mc",
			value: 14,
		},
	}
	ctx := context.Background()
	for _, item := range writeItems {
		err = builder.Add(ctx, []rune(item.key), NewIntBox[int64](item.value))
		assert.Nil(t, err)
	}

	fst, err := builder.Finish(ctx)
	assert.Nil(t, err)

	fstEnum, err := NewEnum[byte](fst)
	assert.Nil(t, err)

	items := []struct {
		key   string
		value int64
		exist bool
	}{
		{
			key:   "m1",
			value: 100,
			exist: true,
		},
		{
			key:   "mc",
			value: 14,
			exist: true,
		},
	}

	for _, item := range items {
		next, ok, err := fstEnum.SeekExact(context.TODO(), []byte(item.key))
		assert.Nil(t, err)
		assert.Equal(t, item.exist, ok)
		if item.exist {
			assert.Equal(t, item.value, next.GetOutput().(*IntBox[int64]).Value())
		}
	}
}

func TestBytesEnumSeekCeil(t *testing.T) {
	builder, err := NewBuilder(BYTE1, NewBoxManager[int64]())
	assert.Nil(t, err)

	writeItems := []struct {
		key   string
		value int64
	}{
		{
			key:   "mop",
			value: 100,
		},
		{
			key:   "moth",
			value: 91,
		},
		{
			key:   "pop",
			value: 72,
		},
		{
			key:   "star",
			value: 83,
		},
		{
			key:   "stop",
			value: 54,
		},
		{
			key:   "top",
			value: 55,
		},
	}
	ctx := context.Background()
	for _, item := range writeItems {
		err = builder.Add(ctx, []rune(item.key), NewIntBox[int64](item.value))
		assert.Nil(t, err)
	}

	fst, err := builder.Finish(ctx)
	assert.Nil(t, err)

	fstEnum, err := NewEnum[byte](fst)
	assert.Nil(t, err)

	items := []struct {
		key    string
		target string
		value  int64
		exist  bool
	}{
		{
			key:    "mop",
			target: "mop",
			value:  100,
			exist:  true,
		},
		{
			key:    "moq",
			target: "moth",
			value:  91,
			exist:  true,
		},
	}

	for _, item := range items {
		next, ok, err := fstEnum.SeekCeil(context.TODO(), []byte(item.key))
		assert.Nil(t, err)
		assert.Equal(t, item.exist, ok)
		if item.exist {
			assert.Equal(t, item.value, next.GetOutput().(*IntBox[int64]).Value())
		}
	}
}

func TestBytesEnumSeekFloor(t *testing.T) {
	builder, err := NewBuilder(BYTE1, NewBoxManager[int64]())
	assert.Nil(t, err)

	writeItems := []struct {
		key   string
		value int64
	}{
		{
			key:   "mop",
			value: 100,
		},
		{
			key:   "moth",
			value: 91,
		},
		{
			key:   "pop",
			value: 72,
		},
		{
			key:   "star",
			value: 83,
		},
		{
			key:   "stop",
			value: 54,
		},
		{
			key:   "top",
			value: 55,
		},
	}
	ctx := context.Background()
	for _, item := range writeItems {
		err = builder.Add(ctx, []rune(item.key), NewIntBox[int64](item.value))
		assert.Nil(t, err)
	}

	fst, err := builder.Finish(ctx)
	assert.Nil(t, err)

	fstEnum, err := NewEnum[byte](fst)
	assert.Nil(t, err)

	items := []struct {
		key    string
		target string
		value  int64
		exist  bool
	}{
		{
			key:    "mop",
			target: "mop",
			value:  100,
			exist:  true,
		},
		{
			key:    "poo",
			target: "moth",
			value:  91,
			exist:  true,
		},
	}

	for _, item := range items {
		next, ok, err := fstEnum.SeekFloor(context.TODO(), []byte(item.key))
		assert.Nil(t, err)
		assert.Equal(t, item.exist, ok)
		if item.exist {
			assert.Equal(t, item.value, next.GetOutput().(*IntBox[int64]).Value())
		}
	}
}
