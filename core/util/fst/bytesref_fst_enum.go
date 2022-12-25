package fst

type BytesRefFSTEnum[T PairAble] struct {
	*FstEnum[T]

	result *InputOutput[T]

	current []byte
	target  []byte
}

func NewBytesRefFSTEnum[T PairAble](fst *FST[T]) *BytesRefFSTEnum[T] {
	fstEnum, err := NewFstEnum(fst)
	if err != nil {
		return nil
	}

	refEnum := &BytesRefFSTEnum[T]{
		FstEnum: fstEnum,
	}
	refEnum.result.Input = refEnum.current

	fstEnum.GetCurrentLabel = refEnum.getCurrentLabel
	fstEnum.GetTargetLabel = refEnum.getTargetLabel
	fstEnum.SetCurrentLabel = refEnum.setCurrentLabel

	return refEnum
}

func (b *BytesRefFSTEnum[T]) Current() *InputOutput[T] {
	return b.result
}

func (b *BytesRefFSTEnum[T]) Next() (*InputOutput[T], error) {
	panic("")
}

// SeekCeil Seeks to smallest term that's >= target.
func (b *BytesRefFSTEnum[T]) SeekCeil(target []byte) (*InputOutput[T], error) {
	panic("")
}

// SeekFloor Seeks to biggest term that's <= target.
func (b *BytesRefFSTEnum[T]) SeekFloor(target []byte) (*InputOutput[T], error) {
	panic("")
}

// SeekExact Seeks to exactly this term, returning null if the term doesn't exist.
// This is faster than using seekFloor or seekCeil because it short-circuits as soon the match is not found.
func (b *BytesRefFSTEnum[T]) SeekExact(target []byte) (*InputOutput[T], error) {
	b.target = target
	b.targetLength = len(b.target)

	ok, err := b.DoSeekExact()
	if err != nil {
		return nil, err
	}
	if ok {
		// assert upto == 1+target.length;
		return b.setResult(), nil
	}
	return nil, nil
}

func (b *BytesRefFSTEnum[T]) setResult() *InputOutput[T] {
	upto := len(b.arcs) - 1
	if upto == 0 {
		return nil
	}
	b.current = b.current[:upto-1]
	b.result.Output = b.output[upto]
	return b.result
}

func (b *BytesRefFSTEnum[T]) getTargetLabel() (int, error) {
	upto := len(b.arcs) - 1
	if upto-1 == len(b.target) {
		return END_LABEL, nil
	}
	return int(b.target[upto-1] & 0xFF), nil
}

func (b *BytesRefFSTEnum[T]) getCurrentLabel() (int, error) {
	upto := len(b.arcs) - 1
	// current.offset fixed at 1
	return int(b.current[upto] & 0xFF), nil
}

func (b *BytesRefFSTEnum[T]) setCurrentLabel(label int) error {
	b.current = append(b.current, byte(label))
	return nil
}

func (b *BytesRefFSTEnum[T]) grow() error {
	panic("")
}
