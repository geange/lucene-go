package fst

import (
	"context"

	"github.com/geange/lucene-go/core/util/array"
)

type IntsFSTEnum[T any] struct {
	enum *FSTEnum[T]

	current []int
	result  *IntsInputOutput[T]
	target  []int
}

func NewIntsFSTEnum[T any](fst *FST[T]) (*IntsFSTEnum[T], error) {
	fstEnum, err := newFSTEnum(fst)
	if err != nil {
		return nil, err
	}
	return &IntsFSTEnum[T]{
		enum:    fstEnum,
		current: make([]int, 10),
		result:  &IntsInputOutput[T]{},
	}, nil
}

type IntsInputOutput[T any] struct {
	input  []int
	output T
}

func (i *IntsInputOutput[T]) GetInput() []int {
	return i.input
}

func (i *IntsInputOutput[T]) GetOutput() T {
	return i.output
}

func (r *IntsFSTEnum[T]) Current() *IntsInputOutput[T] {
	return r.result
}

func (r *IntsFSTEnum[T]) Next(ctx context.Context) (*IntsInputOutput[T], error) {
	if err := r.enum.DoNext(ctx, r); err != nil {
		return nil, err
	}

	return r.setResult(), nil
}

func (r *IntsFSTEnum[T]) SeekCeil(ctx context.Context, target []int) (*IntsInputOutput[T], bool, error) {
	r.target = target
	r.enum.SetTargetLength(len(target))

	if err := r.enum.DoSeekCeil(ctx, r); err != nil {
		return nil, false, err
	}

	output := r.setResult()
	if output == nil {
		return nil, false, nil
	}
	return output, true, nil
}

func (r *IntsFSTEnum[T]) SeekFloor(ctx context.Context, target []int) (*IntsInputOutput[T], bool, error) {
	r.target = target
	r.enum.SetTargetLength(len(target))
	if err := r.enum.DoSeekFloor(ctx, r); err != nil {
		return nil, false, err
	}

	output := r.setResult()
	if output == nil {
		return nil, false, nil
	}
	return output, true, nil
}

func (r *IntsFSTEnum[T]) SeekExact(ctx context.Context, target []int) (*IntsInputOutput[T], bool, error) {
	r.target = target
	r.enum.SetTargetLength(len(r.target))

	ok, err := r.enum.DoSeekExact(ctx, r)
	if err != nil {
		return nil, false, err
	}
	if ok {
		return r.setResult(), true, nil
	}
	return nil, false, nil
}

func (r *IntsFSTEnum[T]) GetTargetLabel(upto int) int {
	if upto-1 == len(r.target) {
		return END_LABEL
	} else {
		return r.target[upto-1]
	}
}
func (r *IntsFSTEnum[T]) GetCurrentLabel(upto int) int {
	return r.current[upto]
}
func (r *IntsFSTEnum[T]) SetCurrentLabel(upto, label int) {
	r.current[upto] = label
}
func (r *IntsFSTEnum[T]) Grow(upto int) {
	r.current = array.Grow(r.current, upto+1)
}

func (r *IntsFSTEnum[T]) setResult() *IntsInputOutput[T] {
	upto := r.enum.upto

	if upto == 0 {
		return nil
	}
	r.result.input = r.current[1:upto]
	r.result.output = r.enum.output[upto]
	return r.result
}
