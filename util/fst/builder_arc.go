package fst

// builderArc Expert: holds a pending (seen but not yet serialized) arc.
type builderArc[T any] struct {
	label           int
	target          Node
	isFinal         bool
	output          *Box[T]
	nextFinalOutput *Box[T]
}

func newBuilderArc[T any]() *builderArc[T] {
	return &builderArc[T]{}
}
