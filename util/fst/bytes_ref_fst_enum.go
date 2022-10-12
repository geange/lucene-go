package fst

import (
	. "github.com/geange/lucene-go/util/structure"
)

var _ FSTEnum = &BytesRefFSTEnum[any]{}

type BytesRefFSTEnum[T any] struct {
	*FSTEnumImp[T]

	current []byte
	target  []byte

	result InputOutput[T]
}

func (b *BytesRefFSTEnum[T]) Current() *InputOutput[T] {
	return &b.result
}

func (b *BytesRefFSTEnum[T]) Next() (*InputOutput[T], error) {
	// TODO
	return b.setResult(), nil
}

// SeekCeil Seeks to smallest term that's >= target.
func (b *BytesRefFSTEnum[T]) SeekCeil(target []byte) (*InputOutput[T], error) {
	b.target = target
	b.targetLength = len(target)
	b.doSeekCeil()
	return b.setResult(), nil
}

// Seeks to biggest term that's <= target.
func (b *BytesRefFSTEnum[T]) seekFloor(target []byte) (*InputOutput[T], error) {
	b.target = target
	b.targetLength = len(target)
	b.doSeekFloor()
	return b.setResult(), nil
}

func (b *BytesRefFSTEnum[T]) GetTargetLabel() int {
	if b.upto-1 == len(b.target) {
		return END_LABEL
	}
	return int(b.target[b.upto-1] & 0xFF)
}

func (b *BytesRefFSTEnum[T]) GetCurrentLabel() int {
	return int(b.current[b.upto] & 0xFF)
}

func (b *BytesRefFSTEnum[T]) SetCurrentLabel(label int) {
	b.current[b.upto] = byte(label)
}

func (b *BytesRefFSTEnum[T]) Grow() {
	b.current = append(b.current, 0)
}

func (b *BytesRefFSTEnum[T]) setResult() *InputOutput[T] {
	if b.upto == 0 {
		return nil
	}

	b.current = b.current[:0]
	b.result.output = b.output[b.upto]
	return &b.result
}

type InputOutput[T any] struct {
	Input  *BytesRef
	output *Box[T]
}
