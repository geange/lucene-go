package fst

import (
	"context"
	"testing"

	"github.com/geange/lucene-go/core/store"
	"github.com/stretchr/testify/assert"
)

func TestNewFstV1(t *testing.T) {
	metaBytes := []byte{
		63, 215, 108, 23, 3, 70, 83, 84,
		0, 0, 0, 7, 0, 0, 40, 41,
	}

	dataBytes := []byte{
		0, 'h', 15, 't', 6, 9, 'p', 25, 'o', 6, 'p', 15, 'o', 6, 'r', 15,
		11, 'o', 2, 15, 29, 'a', 16, 't', 6, 13, 55, 't', 18, 24, 54, 's',
		16, 13, 72, 'p', 16, 9, 91, 'm', 16,
	}

	metaInput := store.NewBytesInput(metaBytes)
	dataInput := store.NewBytesInput(dataBytes)

	ctx := context.Background()
	manager := NewBoxManager[int64]()
	_, err := NewFstV1(ctx, manager, metaInput, dataInput)
	assert.Nil(t, err)
}

func TestReadFirstTarget(t *testing.T) {
	metaBytes := []byte{
		63, 215, 108, 23, 3, 70, 83, 84,
		0, 0, 0, 7, 0, 0, 40, 41,
	}

	dataBytes := []byte{
		0, 'h', 15, 't', 6, 9, 'p', 25, 'o', 6, 'p', 15, 'o', 6, 'r', 15,
		11, 'o', 2, 15, 29, 'a', 16, 't', 6, 13, 55, 't', 18, 24, 54, 's',
		16, 13, 72, 'p', 16, 9, 91, 'm', 16,
	}

	metaInput := store.NewBytesInput(metaBytes)
	dataInput := store.NewBytesInput(dataBytes)

	ctx := context.Background()
	manager := NewBoxManager[int64]()
	fst, err := NewFstV1(ctx, manager, metaInput, dataInput)
	assert.Nil(t, err)

	follow := &Arc{}
	follow, err = fst.GetFirstArc(follow)
	assert.Nil(t, err)

	reader := newReverseBytesReader(dataBytes)
	arc := &Arc{}
	follow, err = fst.ReadFirstTargetArc(ctx, reader, follow, arc)
	assert.Nil(t, err)
	assert.Equal(t, int('m'), arc.Label())

	follow, err = fst.ReadFirstTargetArc(ctx, reader, follow, arc)
	assert.Nil(t, err)
	assert.Equal(t, int('o'), arc.Label())

	follow, err = fst.ReadFirstTargetArc(ctx, reader, follow, arc)
	assert.Nil(t, err)
	assert.Equal(t, int('p'), arc.Label())
}

func TestReadNextArc(t *testing.T) {
	metaBytes := []byte{
		63, 215, 108, 23, 3, 70, 83, 84,
		0, 0, 0, 7, 0, 0, 40, 41,
	}

	dataBytes := []byte{
		0, 'h', 15, 't', 6, 9, 'p', 25, 'o', 6, 'p', 15, 'o', 6, 'r', 15,
		11, 'o', 2, 15, 29, 'a', 16, 't', 6, 13, 55, 't', 18, 24, 54, 's',
		16, 13, 72, 'p', 16, 9, 91, 'm', 16,
	}

	metaInput := store.NewBytesInput(metaBytes)
	dataInput := store.NewBytesInput(dataBytes)

	ctx := context.Background()
	manager := NewBoxManager[int64]()
	fst, err := NewFstV1(ctx, manager, metaInput, dataInput)
	assert.Nil(t, err)

	follow := &Arc{}
	follow, err = fst.GetFirstArc(follow)
	assert.Nil(t, err)

	reader := newReverseBytesReader(dataBytes)

	arc := &Arc{}

	follow, err = fst.ReadFirstTargetArc(ctx, reader, follow, arc)
	assert.Nil(t, err)
	assert.Equal(t, int('m'), arc.Label())
	assert.False(t, arc.IsLast())

	follow, err = fst.ReadNextArc(ctx, follow, reader)
	assert.Nil(t, err)
	assert.Equal(t, int('p'), arc.Label())
	assert.False(t, arc.IsLast())

	follow, err = fst.ReadNextArc(ctx, follow, reader)
	assert.Nil(t, err)
	assert.Equal(t, int('s'), arc.Label())
	assert.False(t, arc.IsLast())

	follow, err = fst.ReadNextArc(ctx, follow, reader)
	assert.Nil(t, err)
	assert.Equal(t, int('t'), arc.Label())
	assert.True(t, arc.IsLast())

	follow = arc
	arc, err = fst.ReadNextArc(ctx, follow, reader)
	assert.Nil(t, err)
	assert.Equal(t, int('t'), arc.Label())
	assert.True(t, arc.IsLast())
}

func TestFindTargetArc(t *testing.T) {
	metaBytes := []byte{
		63, 215, 108, 23, 3, 70, 83, 84,
		0, 0, 0, 7, 0, 0, 40, 41,
	}

	dataBytes := []byte{
		0, 'h', 15, 't', 6, 9, 'p', 25, 'o', 6, 'p', 15, 'o', 6, 'r', 15,
		11, 'o', 2, 15, 29, 'a', 16, 't', 6, 13, 55, 't', 18, 24, 54, 's',
		16, 13, 72, 'p', 16, 9, 91, 'm', 16,
	}

	metaInput := store.NewBytesInput(metaBytes)
	dataInput := store.NewBytesInput(dataBytes)

	ctx := context.Background()
	manager := NewBoxManager[int64]()
	fst, err := NewFstV1(ctx, manager, metaInput, dataInput)
	assert.Nil(t, err)

	follow := &Arc{}
	follow, err = fst.GetFirstArc(follow)
	assert.Nil(t, err)

	reader := newReverseBytesReader(dataBytes)

	arc := &Arc{}
	targetArc, found, err := fst.FindTargetArc(ctx, int('p'), reader, follow, arc)
	assert.Nil(t, err)
	assert.True(t, found)

	targetArc, found, err = fst.FindTargetArc(ctx, int('o'), reader, targetArc, arc)
	assert.Nil(t, err)
	assert.True(t, found)

	targetArc, found, err = fst.FindTargetArc(ctx, int('p'), reader, targetArc, arc)
	assert.Nil(t, err)
	assert.True(t, found)
	assert.True(t, targetArc.IsLast())
}
