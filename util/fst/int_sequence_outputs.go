package fst

import (
	"github.com/geange/lucene-go/core/store"
)

var _ Outputs[[]int] = &IntSequenceOutputs[[]int]{}

type IntSequenceOutputs[T []int] struct {
}

func (i *IntSequenceOutputs[T]) Common(output1, output2 T) T {
	//TODO implement me
	panic("implement me")
}

func (i *IntSequenceOutputs[T]) Subtract(output, inc T) T {
	//TODO implement me
	panic("implement me")
}

func (i *IntSequenceOutputs[T]) Add(prefix, output T) T {
	//TODO implement me
	panic("implement me")
}

func (i *IntSequenceOutputs[T]) Write(output T, out store.DataOutput) error {
	//TODO implement me
	panic("implement me")
}

func (i *IntSequenceOutputs[T]) Read(in store.DataInput) (T, error) {
	//TODO implement me
	panic("implement me")
}

func (i *IntSequenceOutputs[T]) SkipOutput(in store.DataInput) error {
	//TODO implement me
	panic("implement me")
}

func (i *IntSequenceOutputs[T]) GetNoOutput() T {
	//TODO implement me
	panic("implement me")
}

func (i *IntSequenceOutputs[T]) OutputToString(output T) string {
	//TODO implement me
	panic("implement me")
}
