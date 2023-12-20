package fst

import (
	"context"
	"encoding/binary"
	"testing"

	"github.com/geange/lucene-go/core/store"
	"github.com/stretchr/testify/assert"
)

func TestNewBuilderWriteDiffDataOutput(t *testing.T) {
	ctx := context.Background()

	builder, err := NewBuilder(BYTE1, NewBoxManager[int64]())
	assert.Nil(t, err)

	// "mop", "moth", "pop", "star", "stop", "top"
	// 100, 91, 72, 83, 54, 55
	err = builder.Add(ctx, []rune("mop"), NewIntBox[int64](100))
	assert.Nil(t, err)

	err = builder.Add(ctx, []rune("moth"), NewIntBox[int64](91))
	assert.Nil(t, err)

	err = builder.Add(ctx, []rune("pop"), NewIntBox[int64](72))
	assert.Nil(t, err)

	err = builder.Add(ctx, []rune("star"), NewIntBox[int64](83))
	assert.Nil(t, err)

	err = builder.Add(ctx, []rune("stop"), NewIntBox[int64](54))
	assert.Nil(t, err)

	err = builder.Add(ctx, []rune("top"), NewIntBox[int64](55))
	assert.Nil(t, err)

	fst, err := builder.Finish(ctx)
	assert.Nil(t, err)

	metaOutput := store.NewByteBuffersDataOutput()
	dataOutput := store.NewByteBuffersDataOutput()

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

	builder, err := NewBuilder(BYTE1, NewBoxManager[int64]())
	assert.Nil(t, err)

	// "mop", "moth", "pop", "star", "stop", "top"
	// 100, 91, 72, 83, 54, 55
	err = builder.Add(ctx, []rune("mop"), NewIntBox[int64](100))
	assert.Nil(t, err)

	err = builder.Add(ctx, []rune("moth"), NewIntBox[int64](91))
	assert.Nil(t, err)

	err = builder.Add(ctx, []rune("pop"), NewIntBox[int64](72))
	assert.Nil(t, err)

	err = builder.Add(ctx, []rune("star"), NewIntBox[int64](83))
	assert.Nil(t, err)

	err = builder.Add(ctx, []rune("stop"), NewIntBox[int64](54))
	assert.Nil(t, err)

	err = builder.Add(ctx, []rune("top"), NewIntBox[int64](55))
	assert.Nil(t, err)

	fst, err := builder.Finish(ctx)
	assert.Nil(t, err)

	output := store.NewByteBuffersDataOutput()

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
