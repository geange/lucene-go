package fst

import (
	"github.com/geange/lucene-go/core/store"
)

var _ Outputs[[]rune] = &CharSequenceOutputs[[]rune]{}

type CharSequenceOutputs[T []rune] struct {
}

func (c *CharSequenceOutputs[T]) Common(output1, output2 T) T {
	//TODO implement me
	panic("implement me")
}

func (c *CharSequenceOutputs[T]) Subtract(output, inc T) T {
	//TODO implement me
	panic("implement me")
}

func (c *CharSequenceOutputs[T]) Add(prefix, output T) T {
	//TODO implement me
	panic("implement me")
}

func (c *CharSequenceOutputs[T]) Write(output T, out store.DataOutput) error {
	//TODO implement me
	panic("implement me")
}

func (c *CharSequenceOutputs[T]) Read(in store.DataInput) (T, error) {
	//TODO implement me
	panic("implement me")
}

func (c *CharSequenceOutputs[T]) SkipOutput(in store.DataInput) error {
	//TODO implement me
	panic("implement me")
}

func (c *CharSequenceOutputs[T]) GetNoOutput() T {
	//TODO implement me
	panic("implement me")
}

func (c *CharSequenceOutputs[T]) OutputToString(output T) string {
	//TODO implement me
	panic("implement me")
}
