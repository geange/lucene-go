package index

import (
	"context"

	"github.com/geange/lucene-go/core/interface/index"
)

var _ index.SortedNumericDocValues = &SingletonSortedNumericDocValues{}

type SingletonSortedNumericDocValues struct {
	in index.NumericDocValues
}

func NewSingletonSortedNumericDocValues(in index.NumericDocValues) *SingletonSortedNumericDocValues {
	return &SingletonSortedNumericDocValues{in: in}
}

func (s *SingletonSortedNumericDocValues) DocID() int {
	//TODO implement me
	panic("implement me")
}

func (s *SingletonSortedNumericDocValues) NextDoc(ctx context.Context) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SingletonSortedNumericDocValues) Advance(ctx context.Context, target int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SingletonSortedNumericDocValues) SlowAdvance(ctx context.Context, target int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SingletonSortedNumericDocValues) Cost() int64 {
	//TODO implement me
	panic("implement me")
}

func (s *SingletonSortedNumericDocValues) AdvanceExact(target int) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SingletonSortedNumericDocValues) NextValue() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SingletonSortedNumericDocValues) DocValueCount() int {
	//TODO implement me
	panic("implement me")
}
