package fst

import "github.com/geange/lucene-go/core/store"

// Outputs Represents the outputs for an FST, providing the basic algebra required for building and traversing the FST.
// Note that any operation that returns NO_OUTPUT must return the same singleton object from getNoOutput.
// lucene.experimental
type Outputs[T any] interface {

	// TODO: maybe change this API to allow for re-use of the
	// output instances -- this is an insane amount of garbage
	// (new object per byte/char/int) if eg used during
	// analysis

	// Common Eg common("foobar", "food") -> "foo"
	Common(output1, output2 *Box[T]) *Box[T]

	// Subtract Eg subtract("foobar", "foo") -> "bar"
	Subtract(output, inc *Box[T]) *Box[T]

	// Add Eg add("foo", "bar") -> "foobar"
	Add(prefix, output *Box[T]) *Box[T]

	// Write Encode an output value into a DataOutput.
	Write(output *Box[T], out store.DataOutput) error

	// WriteFinalOutput Encode an final node output value into a DataOutput.
	// By default this just calls write(Object, DataOutput).
	WriteFinalOutput(output *Box[T], out store.DataOutput) error

	// Read Decode an output value previously written with write(Object, DataOutput).
	Read(in store.DataInput) (*Box[T], error)

	// ReadFinalOutput Decode an output value previously written with writeFinalOutput(Object, DataOutput).
	// By default this just calls read(DataInput).
	ReadFinalOutput(in store.DataInput) (*Box[T], error)

	// SkipOutput Skip the output; defaults to just calling read and discarding the result.
	SkipOutput(in store.DataInput) error

	// SkipFinalOutput Skip the output previously written with writeFinalOutput;
	// defaults to just calling readFinalOutput and discarding the result.
	SkipFinalOutput(in store.DataInput) error

	// GetNoOutput NOTE: this output is compared with == so you must ensure that all methods return the single object if it's really no output
	GetNoOutput() *Box[T]

	OutputToString(output *Box[T]) string

	Merge(first, second *Box[T]) *Box[T]
}
