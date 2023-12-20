package fst

import (
	"context"
	"slices"
)

type AbsEnum interface {
	GetUpTo() int
	GetOutput(idx int) Output
	SetTargetLength(size int)
	DoNext(ctx context.Context, lm LabelManager) error
	DoSeekCeil(ctx context.Context, lm LabelManager) error
	DoSeekFloor(ctx context.Context, lm LabelManager) error
	DoSeekExact(ctx context.Context, lm LabelManager) (bool, error)
}

// EnumWrap
// Enumerates all input (BytesRef) + output pairs in an FST.
// lucene.experimental
type EnumWrap[T byte | int] struct {
	enum    AbsEnum
	result  *InputOutput[T]
	current []T
	target  []T
}

// InputOutput Holds a single input (BytesRef) + output pair.
type InputOutput[T byte | int] struct {
	input  []T
	output Output
}

func (i *InputOutput[T]) GetInput() []T {
	return i.input
}

func (i *InputOutput[T]) GetOutput() Output {
	return i.output
}

func NewEnumWrap[T int | byte](fst *FST) (*EnumWrap[T], error) {
	fstEnum, err := NewEnum(fst)
	if err != nil {
		return nil, err
	}

	refEnum := &EnumWrap[T]{
		enum:    fstEnum,
		current: make([]T, 10),
		result:  new(InputOutput[T]),
	}
	return refEnum, nil
}

func (b *EnumWrap[T]) Current() *InputOutput[T] {
	return b.result
}

func (b *EnumWrap[T]) Next(ctx context.Context) (*InputOutput[T], error) {
	if err := b.enum.DoNext(ctx, b); err != nil {
		return nil, err
	}

	return b.setResult(), nil
}

// SeekCeil Seeks to smallest term that's >= target.
func (b *EnumWrap[T]) SeekCeil(ctx context.Context, target []T) (*InputOutput[T], bool, error) {
	b.target = target
	b.enum.SetTargetLength(len(target))

	if err := b.enum.DoSeekCeil(ctx, b); err != nil {
		return nil, false, err
	}

	output := b.setResult()
	if output == nil {
		return nil, false, nil
	}
	return output, true, nil
}

// SeekFloor Seeks to biggest term that's <= target.
func (b *EnumWrap[T]) SeekFloor(ctx context.Context, target []T) (*InputOutput[T], bool, error) {
	b.target = target
	b.enum.SetTargetLength(len(target))
	if err := b.enum.DoSeekFloor(ctx, b); err != nil {
		return nil, false, err
	}

	output := b.setResult()
	if output == nil {
		return nil, false, nil
	}
	return output, true, nil
}

// SeekExact Seeks to exactly this term, returning null if the term doesn't exist.
// This is faster than using seekFloor or seekCeil because it short-circuits as soon the match is not found.
func (b *EnumWrap[T]) SeekExact(ctx context.Context, target []T) (*InputOutput[T], bool, error) {
	b.target = target
	b.enum.SetTargetLength(len(b.target))

	ok, err := b.enum.DoSeekExact(ctx, b)
	if err != nil {
		return nil, false, err
	}
	if ok {
		return b.setResult(), true, nil
	}
	return nil, false, nil
}

func (b *EnumWrap[T]) setResult() *InputOutput[T] {
	if b.enum.GetUpTo() == 0 {
		return nil
	}
	b.result.input = b.current[1:b.enum.GetUpTo()]
	b.result.output = b.enum.GetOutput(b.enum.GetUpTo())
	return b.result
}

func (b *EnumWrap[T]) GetTargetLabel(upto int) int {
	if upto-1 == len(b.target) {
		return END_LABEL
	} else {
		return int(b.target[upto-1])
	}
}

func (b *EnumWrap[T]) GetCurrentLabel(upto int) int {
	return int(b.current[upto])
}

func (b *EnumWrap[T]) SetCurrentLabel(label int) error {
	b.current[b.enum.GetUpTo()] = T(label)
	return nil
}

func (b *EnumWrap[T]) Grow() {
	b.current = slices.Grow(b.current, b.enum.GetUpTo()+1)
}
