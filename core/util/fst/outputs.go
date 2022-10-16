package fst

// Outputs Represents the outputs for an FST, providing the basic algebra required for building and traversing the FST.
// Note that any operation that returns NO_OUTPUT must return the same singleton object from getNoOutput.
// lucene.experimental
type Outputs interface {

	// TODO: maybe change this API to allow for re-use of the
	// output instances -- this is an insane amount of garbage
	// (new object per byte/char/int) if eg used during
	// analysis

	// Common Eg common("foobar", "food") -> "foo"
	Common(output1, output2 any) any

	Subtract(output1, output2 any) any

	Add(output1, output2 any) any
}
