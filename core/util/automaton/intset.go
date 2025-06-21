package automaton

import "slices"

type IntSet interface {
	Hashable

	GetArray() []int

	Size() int
}

var _ IntSet = &MockIntSet{}

type MockIntSet struct {
}

func (m MockIntSet) Hash() uint64 {
	return 0
}

func (m MockIntSet) Equals(other Hashable) bool {
	return false
}

func (m MockIntSet) GetArray() []int {
	return make([]int, 0)
}

func (m MockIntSet) Size() int {
	return 0
}

var _ IntSet = &FrozenIntSet{}

type FrozenIntSet struct {
	values   []int
	state    int
	hashCode uint64
}

func (f *FrozenIntSet) Hash() uint64 {
	return f.hashCode
}

func (f *FrozenIntSet) Equals(other Hashable) bool {
	if f == nil {
		switch other.(type) {
		case *FrozenIntSet:
			ptr := other.(*FrozenIntSet)
			if ptr == nil {
				return true
			}
		case *StateSet:
			ptr := other.(*StateSet)
			if ptr == nil {
				return true
			}
		default:
			return false
		}
	}

	is, ok := other.(IntSet)
	if !ok {
		return false
	}
	return is.Hash() == f.Hash()
}

func NewFrozenIntSet(values []int, hashCode uint64, state int) *FrozenIntSet {
	return &FrozenIntSet{values: values, state: state, hashCode: hashCode}
}

func (f *FrozenIntSet) GetArray() []int {
	return f.values
}

func (f *FrozenIntSet) Size() int {
	return len(f.values)
}

var _ IntSet = &StateSet{}

type StateSet struct {
	inner       map[int]int
	hashUpdated bool
	hashCode    uint64
}

func NewStateSet() *StateSet {
	return &StateSet{
		inner: make(map[int]int),
	}
}

func (s *StateSet) Hash() uint64 {
	if s.hashUpdated {
		return s.hashCode
	}
	s.hashCode = uint64(len(s.inner))
	for key := range s.inner {
		s.hashCode += uint64(mix(key))
	}
	s.hashUpdated = true
	return s.hashCode
}

func (s *StateSet) Equals(other Hashable) bool {
	is, ok := other.(IntSet)
	if !ok {
		return false
	}
	return s.Hash() == is.Hash()
}

func (s *StateSet) GetArray() []int {
	keys := make([]int, 0, len(s.inner))

	for k := range s.inner {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

func (s *StateSet) Size() int {
	return len(s.inner)
}

func (s *StateSet) keyChanged() {
	s.hashUpdated = false
	s.hashCode = 0
}

func (s *StateSet) Incr(state int) {
	s.inner[state]++
	if s.inner[state] == 1 {
		s.keyChanged()
	}
}

func (s *StateSet) Decr(state int) {
	count, ok := s.inner[state]
	if !ok {
		return
	}
	if count == 1 {
		delete(s.inner, state)
		s.keyChanged()
	} else {
		s.inner[state]--
	}
}

func (s *StateSet) Freeze(state int) *FrozenIntSet {
	return NewFrozenIntSet(s.GetArray(), s.hashCode, state)
}
