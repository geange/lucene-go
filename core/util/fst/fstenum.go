package fst

type FSTEnum[T any] struct {
	fst          *FST[T]
	arcs         []*Arc[T]
	output       []T
	noOutput     T
	fstReader    BytesReader
	upto         int
	targetLength int
}

type FSTEnumLabel interface {
	GetTargetLabel(upto int) int
	GetCurrentLabel(upto int) int
	SetCurrentLabel(upto, label int)
	Grow(upto int)
}
