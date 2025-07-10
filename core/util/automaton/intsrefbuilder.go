package automaton

import "github.com/geange/lucene-go/core/util/array"

type IntsRefBuilder struct {
	ref *IntsRef
}

func NewIntsRefBuilder() *IntsRefBuilder {
	return &IntsRefBuilder{ref: NewIntsRef()}
}

type IntsRef struct {
	ints   []int
	offset int
	length int
}

func NewIntsRef() *IntsRef {
	return &IntsRef{
		ints: make([]int, 0),
	}
}

func (i *IntsRefBuilder) SetLength(length int) {
	i.ref.length = length
}

func (i *IntsRefBuilder) Length() int {
	return i.ref.length
}

func (i *IntsRefBuilder) Clear() {
	i.SetLength(0)
}

func (i *IntsRefBuilder) At(offset int) int {
	return i.ref.ints[offset]
}

func (i *IntsRefBuilder) Set(offset, value int) {
	i.ref.ints[offset] = value
}

func (i *IntsRefBuilder) Append(value int) {
	if i.ref.offset+i.ref.length >= len(i.ref.ints) {
		i.ref.ints = append(i.ref.ints, value)
	}
	i.ref.length++
}

func (i *IntsRefBuilder) Grow(depth int) {
	i.ref.ints = array.Grow(i.ref.ints, depth)
}

func (i *IntsRefBuilder) Get() []int {
	return i.ref.ints
}
