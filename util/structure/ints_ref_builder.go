package structure

// IntsRefBuilder A builder for IntsRef instances.
// lucene.internal
type IntsRefBuilder struct {
	buff []int
}

func (i *IntsRefBuilder) Length() int {
	return len(i.buff)
}

func (i *IntsRefBuilder) IntAt(offset int) int {
	return i.buff[offset]
}

func (i *IntsRefBuilder) CopyInts(ints []int) {
	if len(i.buff) < len(ints) {
		i.buff = append(i.buff, make([]int, len(ints)-len(i.buff))...)
	}
	copy(i.buff, ints)
}
