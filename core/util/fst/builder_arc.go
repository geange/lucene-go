package fst

// Arc Expert: holds a pending (seen but not yet serialized) arc.
type Arc struct {
	label           int
	target          Node
	isFinal         bool
	output          any
	nextFinalOutput any
}
