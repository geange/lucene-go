package automaton

import "iter"

type LimitedFiniteStringsIterator struct {
	*FiniteStringsIterator

	limit int
	count int
}

func NewLimitedFiniteStringsIterator(a *Automaton, limit int) *LimitedFiniteStringsIterator {
	return &LimitedFiniteStringsIterator{
		FiniteStringsIterator: NewFiniteStringsIteratorBuilder(a).New(),
		limit:                 limit,
		count:                 0,
	}
}

func (l *LimitedFiniteStringsIterator) Iterator() iter.Seq[[]int] {
	return func(yield func([]int) bool) {

	}
}
