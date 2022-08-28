package fst

import "github.com/geange/lucene-go/core/store"

var _ Outputs[any] = &NoOutputs[any]{}

var (
	singleton = NewNoOutputs[any]()

	NO_OUTPUT any
)

type NoOutputs[T any] struct {
}

func GetSingleton() *NoOutputs[any] {
	return singleton
}

func NewNoOutputs[T any]() *NoOutputs[T] {
	return &NoOutputs[T]{}
}

func (n *NoOutputs[T]) Common(output1, output2 T) T {
	return NO_OUTPUT
}

func (n *NoOutputs[T]) Subtract(output, inc T) T {
	return NO_OUTPUT
}

func (n *NoOutputs[T]) Add(prefix, output T) T {
	return NO_OUTPUT
}

func (n *NoOutputs[T]) Write(output T, out store.DataOutput) error {
	return nil
}

func (n *NoOutputs[T]) Read(in store.DataInput) (T, error) {
	return NO_OUTPUT, nil
}

func (n *NoOutputs[T]) SkipOutput(in store.DataInput) error {
	_, err := n.Read(in)
	return err
}

func (n *NoOutputs[T]) GetNoOutput() T {
	return NO_OUTPUT
}

func (n *NoOutputs[T]) OutputToString(output T) string {
	//TODO implement me
	panic("implement me")
}
