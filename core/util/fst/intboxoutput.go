package fst

import (
	"context"
	"errors"

	"github.com/geange/lucene-go/core/store"
)

type IntBox[T Int] struct {
	value T
}

func NewIntBox[T Int](v T) *IntBox[T] {
	return &IntBox[T]{value: v}
}

func (b *IntBox[T]) Value() T {
	return b.value
}

func (b *IntBox[T]) check(v Output) (*IntBox[T], error) {
	if box, ok := v.(*IntBox[T]); ok {
		return box, nil
	}
	return nil, errors.New("not *Box[T]")
}

func (b *IntBox[T]) Common(v Output) (Output, error) {
	in, err := b.check(v)
	if err != nil {
		return nil, err
	}

	return &IntBox[T]{
		value: min(b.value, in.value),
	}, nil
}

func (b *IntBox[T]) Sub(v Output) (Output, error) {
	in, err := b.check(v)
	if err != nil {
		return nil, err
	}

	return &IntBox[T]{
		value: b.value - in.value,
	}, nil
}

func (b *IntBox[T]) Add(v Output) (Output, error) {
	in, err := b.check(v)
	if err != nil {
		return nil, err
	}
	return &IntBox[T]{
		value: b.value + in.value,
	}, nil
}

func (b *IntBox[T]) Merge(v Output) (Output, error) {
	//TODO implement me
	panic("implement me")
}

func (b *IntBox[T]) IsNoOutput() bool {
	return b.value == T(0)
}

func (b *IntBox[T]) Equal(v Output) bool {
	check, err := b.check(v)
	if err != nil {
		return false
	}
	return b.value == check.value
}

func (b *IntBox[T]) Hash() int64 {
	return int64(b.value)
}

type BoxManager[T Int] struct {
	empty Output
}

func NewBoxManager[T Int]() *BoxManager[T] {
	return &BoxManager[T]{}
}

func (b *BoxManager[T]) EmptyOutput() Output {
	if b.empty == nil {
		b.empty = b.New()
	}
	return b.empty
}

func (b *BoxManager[T]) New() Output {
	return &IntBox[T]{}
}

func (b *BoxManager[T]) check(v any) (*IntBox[T], error) {
	box, ok := v.(*IntBox[T])
	if ok {
		return box, nil
	}
	return nil, errors.New("not *Box[T]")
}

func (b *BoxManager[T]) Read(ctx context.Context, in store.DataInput, v any) error {
	box, err := b.check(v)
	if err != nil {
		return err
	}

	n, err := in.ReadUvarint(ctx)
	if err != nil {
		return err
	}
	box.value = T(n)
	return nil
}

func (b *BoxManager[T]) SkipOutput(ctx context.Context, in store.DataInput) error {
	_, err := in.ReadUvarint(nil)
	return err
}

func (b *BoxManager[T]) ReadFinalOutput(ctx context.Context, in store.DataInput, v any) error {
	return b.Read(ctx, in, v)
}

func (b *BoxManager[T]) SkipFinalOutput(ctx context.Context, in store.DataInput) error {
	return b.SkipOutput(ctx, in)
}

func (b *BoxManager[T]) Write(ctx context.Context, out store.DataOutput, v any) error {
	box, err := b.check(v)
	if err != nil {
		return err
	}
	return out.WriteUvarint(ctx, uint64(box.value))
}

func (b *BoxManager[T]) WriteFinalOutput(ctx context.Context, out store.DataOutput, v any) error {
	return b.Write(ctx, out, v)
}
