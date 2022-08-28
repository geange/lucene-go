package fst

import "github.com/geange/lucene-go/core/store"

type Outputs[T any] interface {

	// TODO: maybe change this API to allow for re-use of the
	// output instances -- this is an insane amount of garbage
	// (new object per byte/char/int) if eg used during
	// analysis

	// Common Eg common("foobar", "food") -> "foo"
	Common(o1, o2 T) T

	// Subtract Eg subtract("foobar", "foo") -> "bar"
	Subtract(o1, o2 T) T

	// Add Eg add("foo", "bar") -> "foobar"
	Add(prefix, output T) T

	// Write Encode an output value into a DataOutput.
	Write(output T, out store.DataOutput) error

	// Read Decode an output value previously written with write(Object, DataOutput).
	Read(in store.DataInput) (T, error)

	// GetNoOutput NOTE: this output is compared with == so you must ensure that all methods return
	// the single object if it's really no output
	GetNoOutput() T

	//OutputToString(output T) string

	SkipOutput(in store.DataInput) error
}

type OutputsExt[T any] interface {
	// WriteFinalOutput Encode an final node output value into a DataOutput.
	// By default this just calls write(Object, DataOutput).
	WriteFinalOutput(output T, out store.DataOutput) error

	// SkipOutput Skip the output; defaults to just calling read and discarding the result.
	SkipOutput(in store.DataInput) error

	// ReadFinalOutput Decode an output value previously written with writeFinalOutput(Object, DataOutput).
	// By default this just calls read(DataInput).
	ReadFinalOutput(in store.DataInput) (T, error)

	// SkipFinalOutput Skip the output previously written with writeFinalOutput; defaults to just calling readFinalOutput and discarding the result.
	SkipFinalOutput(in store.DataInput) error

	Merge(first, second T) T
}

var _ OutputsExt[any] = &OutputsImp[any]{}

type OutputsImp[T any] struct {
	Outputs[T]
}

func (o *OutputsImp[T]) WriteFinalOutput(output T, out store.DataOutput) error {
	return o.Write(output, out)
}

func (o *OutputsImp[T]) SkipOutput(in store.DataInput) error {
	_, err := o.Read(in)
	return err
}

func (o *OutputsImp[T]) ReadFinalOutput(in store.DataInput) (T, error) {
	return o.Read(in)
}

func (o *OutputsImp[T]) SkipFinalOutput(in store.DataInput) error {
	return o.Outputs.SkipOutput(in)
}

func (o *OutputsImp[T]) Merge(first, second T) T {
	panic("UnsupportedOperationException")
}
