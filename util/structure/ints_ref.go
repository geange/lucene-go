package structure

type IntsRef struct {
	Ints []int
}

func NewIntsRef(ints []int) *IntsRef {
	return &IntsRef{Ints: ints}
}

func (i *IntsRef) Len() int {
	return len(i.Ints)
}
