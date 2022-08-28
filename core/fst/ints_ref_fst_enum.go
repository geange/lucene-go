package fst

var _ FSTEnum = &IntsRefFSTEnum{}

// IntsRefFSTEnum Enumerates all input (IntsRef) + output pairs in an FST.
// lucene.experimental
type IntsRefFSTEnum struct {
}

func (i *IntsRefFSTEnum) GetTargetLabel() int {
	//TODO implement me
	panic("implement me")
}

func (i *IntsRefFSTEnum) GetCurrentLabel() int {
	//TODO implement me
	panic("implement me")
}

func (i *IntsRefFSTEnum) SetCurrentLabel(label int) {
	//TODO implement me
	panic("implement me")
}

func (i *IntsRefFSTEnum) Grow() {
	//TODO implement me
	panic("implement me")
}
