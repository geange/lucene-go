package fst

// FSTEnum Can next() and advance() through the terms in an FST
// lucene.experimental
type FSTEnum interface {
	GetTargetLabel() int
	GetCurrentLabel() int
	SetCurrentLabel(label int)
	Grow()
}
