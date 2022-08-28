package fst

type BuilderArc[T any] struct {
	Label           int
	Target          Node
	IsFinal         bool
	Output          T
	NextFinalOutput T
}
