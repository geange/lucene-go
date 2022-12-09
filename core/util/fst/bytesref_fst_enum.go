package fst

var _ enumSPI = &BytesRefFSTEnum[PairAble]{}

type BytesRefFSTEnum[T PairAble] struct {
	Enum[T]

	current []byte
	target  []byte
}

func (b *BytesRefFSTEnum[T]) getTargetLabel() (int, error) {
	if b.upto-1 == len(b.target) {
		return END_LABEL, nil
	}
	return int(b.target[b.upto-1] & 0xFF), nil
}

func (b *BytesRefFSTEnum[T]) getCurrentLabel() (int, error) {
	// current.offset fixed at 1
	return int(b.current[b.upto] & 0xFF), nil
}

func (b *BytesRefFSTEnum[T]) setCurrentLabel(label int) error {
	b.current[b.upto] = byte(label)
	return nil
}

func (b *BytesRefFSTEnum[T]) grow() error {
	panic("")
}
