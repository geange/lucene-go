package fst

import (
	"context"
	"encoding/binary"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/geange/lucene-go/core/store"
)

func TestNewBuilderWriteDiffDataOutput(t *testing.T) {
	ctx := context.Background()

	builder, err := NewBuilder[int64](BYTE1, NewPositiveIntOutputs())
	assert.Nil(t, err)

	items := []struct {
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

	for _, item := range items {
		err := builder.Add(ctx, []rune(item.key), item.value)
		assert.Nil(t, err)
	}

	count := builder.GetTermCount()
	assert.Equal(t, len(items), count)

	fst, err := builder.Finish(ctx)
	assert.Nil(t, err)

	metaOutput := store.NewBufferDataOutput()
	dataOutput := store.NewBufferDataOutput()

	err = fst.Save(ctx, metaOutput, dataOutput)
	assert.Nil(t, err)

	metaBytes := metaOutput.Bytes()
	assert.Equal(t, 16, len(metaBytes))

	dataBytes := dataOutput.Bytes()
	bs := binary.AppendVarint(nil, 41)
	assert.NotNil(t, bs)

	data := []byte{
		0, 'h', 15, 't', 6, 9, 'p', 25, 'o', 6, 'p', 15, 'o', 6, 'r', 15,
		11, 'o', 2, 15, 29, 'a', 16, 't', 6, 13, 55, 't', 18, 24, 54, 's',
		16, 13, 72, 'p', 16, 9, 91, 'm', 16,
	}

	assert.Equal(t, data, dataBytes)
}

func TestNewBuilderWithOptions(t *testing.T) {
	ctx := context.Background()

	options := []BuilderOption{
		WithMinSuffixCount1(0),
		WithMinSuffixCount2(0),
		WithDoShareSuffix(true),
		WithDoShareNonSingletonNodes(true),
		WithShareMaxTailLength(math.MaxInt32),
		WithAllowFixedLengthArcs(true),
		WithBytesPageBits(15),
	}

	builder, err := NewBuilder[int64](BYTE1, NewPositiveIntOutputs(), options...)
	assert.Nil(t, err)

	items := []struct {
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

	for _, item := range items {
		err := builder.Add(ctx, []rune(item.key), item.value)
		assert.Nil(t, err)
	}

	fst, err := builder.Finish(ctx)
	assert.Nil(t, err)

	metaOutput := store.NewBufferDataOutput()
	dataOutput := store.NewBufferDataOutput()

	err = fst.Save(ctx, metaOutput, dataOutput)
	assert.Nil(t, err)

	metaBytes := metaOutput.Bytes()
	assert.Equal(t, 16, len(metaBytes))

	dataBytes := dataOutput.Bytes()
	bs := binary.AppendVarint(nil, 41)
	assert.NotNil(t, bs)

	data := []byte{
		0, 'h', 15, 't', 6, 9, 'p', 25, 'o', 6, 'p', 15, 'o', 6, 'r', 15,
		11, 'o', 2, 15, 29, 'a', 16, 't', 6, 13, 55, 't', 18, 24, 54, 's',
		16, 13, 72, 'p', 16, 9, 91, 'm', 16,
	}

	assert.Equal(t, data, dataBytes)
}

func TestNewBuilderWriteSameDataOutput(t *testing.T) {
	ctx := context.Background()

	builder, err := NewBuilder[int64](BYTE1, NewPositiveIntOutputs())
	assert.Nil(t, err)

	items := []struct {
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

	for _, item := range items {
		err := builder.Add(ctx, []rune(item.key), item.value)
		assert.Nil(t, err)
	}

	fst, err := builder.Finish(ctx)
	assert.Nil(t, err)

	output := store.NewBufferDataOutput()

	err = fst.Save(ctx, output, output)
	assert.Nil(t, err)

	dataBytes := output.Bytes()

	data := []byte{
		0, 'h', 15, 't', 6, 9, 'p', 25, 'o', 6, 'p', 15, 'o', 6, 'r', 15,
		11, 'o', 2, 15, 29, 'a', 16, 't', 6, 13, 55, 't', 18, 24, 54, 's',
		16, 13, 72, 'p', 16, 9, 91, 'm', 16,
	}

	assert.Equal(t, data, dataBytes[16:])
}

func TestNewBuilderWithBYTE1(t *testing.T) {
	ctx := context.Background()

	builder, err := NewBuilder[int64](BYTE4, NewPositiveIntOutputs())
	assert.Nil(t, err)

	items := []struct {
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

	for _, item := range items {
		err := builder.Add(ctx, []rune(item.key), item.value)
		assert.Nil(t, err)
	}

	fst, err := builder.Finish(ctx)
	assert.Nil(t, err)

	output := store.NewBufferDataOutput()

	err = fst.Save(ctx, output, output)
	assert.Nil(t, err)

	fstEnum, err := NewIntsFSTEnum(fst)
	assert.Nil(t, err)

	next, ok, err := fstEnum.SeekExact(context.TODO(), str2Ints("top"))
	assert.Nil(t, err)
	assert.True(t, ok)
	assert.Equal(t, int64(55), next.GetOutput())
}

func TestNewBuilderAddWithString(t *testing.T) {
	ctx := context.Background()

	builder, err := NewBuilder[int64](BYTE1, NewPositiveIntOutputs())
	assert.Nil(t, err)

	items := []struct {
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

	for _, item := range items {
		err := builder.AddStr(ctx, item.key, item.value)
		assert.Nil(t, err)
	}

	fst, err := builder.Finish(ctx)
	if !assert.Nil(t, err) {
		t.FailNow()
	}

	metaOutput := store.NewBufferDataOutput()
	dataOutput := store.NewBufferDataOutput()

	err = fst.Save(ctx, metaOutput, dataOutput)
	assert.Nil(t, err)

	metaBytes := metaOutput.Bytes()
	assert.Equal(t, 16, len(metaBytes))

	dataBytes := dataOutput.Bytes()
	bs := binary.AppendVarint(nil, 41)
	assert.NotNil(t, bs)

	data := []byte{
		0, 'h', 15, 't', 6, 9, 'p', 25, 'o', 6, 'p', 15, 'o', 6, 'r', 15,
		11, 'o', 2, 15, 29, 'a', 16, 't', 6, 13, 55, 't', 18, 24, 54, 's',
		16, 13, 72, 'p', 16, 9, 91, 'm', 16,
	}

	assert.Equal(t, data, dataBytes)
}

func str2Ints(s string) []int {
	runes := []rune(s)

	items := make([]int, 0, len(runes))
	for _, v := range runes {
		items = append(items, int(v))
	}
	return items
}
