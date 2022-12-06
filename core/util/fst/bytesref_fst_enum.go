package fst

var _ EnumSPI = &BytesRefFSTEnum[PairAble]{}

type BytesRefFSTEnum[T PairAble] struct {
	Enum[T]
}

func (b *BytesRefFSTEnum[T]) GetTargetLabel() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (b *BytesRefFSTEnum[T]) GetCurrentLabel() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (b *BytesRefFSTEnum[T]) SetCurrentLabel(label int) error {
	//TODO implement me
	panic("implement me")
}

func (b *BytesRefFSTEnum[T]) Grow() error {
	//TODO implement me
	panic("implement me")
}
