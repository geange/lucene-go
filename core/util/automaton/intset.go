package automaton

type IntSet interface {
	GetArray() []int

	Size() int

	Hash() int64
}

var _ IntSet = &FrozenIntSet{}

type FrozenIntSet struct {
	values   []int
	state    int
	hashCode int64
}

func NewFrozenIntSet(values []int, state int, hashCode int64) *FrozenIntSet {
	return &FrozenIntSet{values: values, state: state, hashCode: hashCode}
}

func (f *FrozenIntSet) GetArray() []int {
	return f.values
}

func (f *FrozenIntSet) Size() int {
	return len(f.values)
}

func (f *FrozenIntSet) Hash() int64 {
	return f.hashCode
}
