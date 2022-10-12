package fst

import "github.com/geange/lucene-go/core/store"

var _ Outputs[any] = &NoOutputs[any]{}

var (
	singleton = NewNoOutputs[any]()

	NO_OUTPUT = NewBox[any](nil)
)

type NoOutputs[T any] struct {
}

func GetSingleton() *NoOutputs[any] {
	return singleton
}

func NewNoOutputs[T any]() *NoOutputs[T] {
	return &NoOutputs[T]{}
}

func (n *NoOutputs[T]) Common(output1, output2 *Box[any]) *Box[any] {
	return NO_OUTPUT
}

func (n *NoOutputs[T]) Subtract(output, inc *Box[any]) *Box[any] {
	return NO_OUTPUT
}

func (n *NoOutputs[T]) Add(prefix, output *Box[any]) *Box[any] {
	return NO_OUTPUT
}

func (n *NoOutputs[T]) Write(output *Box[any], out store.DataOutput) error {
	return nil
}

func (n *NoOutputs[T]) WriteFinalOutput(output *Box[any], out store.DataOutput) error {
	return nil
}

func (n *NoOutputs[T]) Read(in store.DataInput) (*Box[any], error) {
	return NO_OUTPUT, nil
}

func (n *NoOutputs[T]) ReadFinalOutput(in store.DataInput) (*Box[any], error) {
	return n.Read(in)
}

func (n *NoOutputs[T]) SkipOutput(in store.DataInput) error {
	_, err := n.Read(in)
	return err
}

func (n *NoOutputs[T]) SkipFinalOutput(in store.DataInput) error {
	return n.SkipOutput(in)
}

func (n *NoOutputs[T]) GetNoOutput() *Box[any] {
	return NO_OUTPUT
}

func (n *NoOutputs[T]) OutputToString(output *Box[any]) string {
	return ""
}

func (n *NoOutputs[T]) Merge(first, second *Box[any]) *Box[any] {
	return NO_OUTPUT
}
