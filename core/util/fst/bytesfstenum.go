package fst

import (
	"context"

	"github.com/geange/lucene-go/core/util/array"
)

type BytesFSTEnum[T any] struct {
	enum    *FSTEnum[T]
	current []byte
	result  *BytesInputOutput[T]
	target  []byte
}

func NewBytesFSTEnum[T any](fst *FST[T]) (*BytesFSTEnum[T], error) {
	fstEnum, err := newFSTEnum(fst)
	if err != nil {
		return nil, err
	}
	return &BytesFSTEnum[T]{
		enum:    fstEnum,
		current: make([]byte, 10),
		result:  &BytesInputOutput[T]{},
	}, nil
}

type BytesInputOutput[T any] struct {
	input  []byte
	output T
}

func (b *BytesInputOutput[T]) GetInput() []byte {
	return b.input
}

func (b *BytesInputOutput[T]) GetOutput() T {
	return b.output
}

func (r *BytesFSTEnum[T]) Current() *BytesInputOutput[T] {
	return r.result
}

func (r *BytesFSTEnum[T]) Next(ctx context.Context) (*BytesInputOutput[T], error) {
	if err := r.enum.DoNext(ctx, r); err != nil {
		return nil, err
	}

	return r.setResult(), nil
}

func (r *BytesFSTEnum[T]) SeekCeil(ctx context.Context, target []byte) (*BytesInputOutput[T], bool, error) {
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

func (r *BytesFSTEnum[T]) SeekFloor(ctx context.Context, target []byte) (*BytesInputOutput[T], bool, error) {
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

func (r *BytesFSTEnum[T]) SeekExact(ctx context.Context, target []byte) (*BytesInputOutput[T], bool, error) {
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

func (r *BytesFSTEnum[T]) GetTargetLabel(upto int) int {
	if upto-1 == len(r.target) {
		return END_LABEL
	} else {
		return int(r.target[upto-1])
	}
}
func (r *BytesFSTEnum[T]) GetCurrentLabel(upto int) int {
	return int(r.current[upto])
}
func (r *BytesFSTEnum[T]) SetCurrentLabel(upto, label int) {
	r.current[upto] = byte(label)
}
func (r *BytesFSTEnum[T]) Grow(upto int) {
	r.current = array.Grow(r.current, upto+1)
}

func (r *BytesFSTEnum[T]) setResult() *BytesInputOutput[T] {
	upto := r.enum.upto
	if upto == 0 {
		return nil
	}

	r.result.input = r.current[1:upto]
	r.result.output = r.enum.output[upto]
	return r.result
}
