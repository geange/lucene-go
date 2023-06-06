package fst

// BytesRefFSTEnum
// Enumerates all input (BytesRef) + output pairs in an Fst.
// lucene.experimental
type BytesRefFSTEnum[T PairAble] struct {
	*FstEnum[T]

	result *InputOutput[T]

	current []byte
	target  []byte
}

func NewBytesRefFSTEnum[T PairAble](fst *Fst[T]) *BytesRefFSTEnum[T] {
	fstEnum, err := NewFstEnum(fst)
	if err != nil {
		return nil
	}

	refEnum := &BytesRefFSTEnum[T]{
		FstEnum: fstEnum,
		result:  new(InputOutput[T]),
	}
	refEnum.result.Input = refEnum.current

	fstEnum.GetCurrentLabel = refEnum.getCurrentLabel
	fstEnum.GetTargetLabel = refEnum.getTargetLabel
	fstEnum.SetCurrentLabel = refEnum.setCurrentLabel
	fstEnum.Grow = refEnum.grow

	return refEnum
}

func (b *BytesRefFSTEnum[T]) Current() *InputOutput[T] {
	return b.result
}

func (b *BytesRefFSTEnum[T]) Next() (*InputOutput[T], error) {
	err := b.doNext()
	if err != nil {
		return nil, err
	}
	return b.setResult(), nil
}

// SeekCeil Seeks to smallest term that's >= target.
func (b *BytesRefFSTEnum[T]) SeekCeil(target []byte) (*InputOutput[T], error) {
	b.target = target
	b.targetLength = len(target)
	err := b.FstEnum.doSeekCeil()
	if err != nil {
		return nil, err
	}
	return b.setResult(), nil
}

// SeekFloor Seeks to biggest term that's <= target.
func (b *BytesRefFSTEnum[T]) SeekFloor(target []byte) (*InputOutput[T], error) {
	b.target = target
	b.targetLength = len(target)
	err := b.FstEnum.doSeekFloor()
	if err != nil {
		return nil, err
	}
	return b.setResult(), nil
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
	size := len(b.output)
	b.current = b.current[:size]
	b.result.Output = b.output[size-1]
	return b.result
}

func (b *BytesRefFSTEnum[T]) getTargetLabel() (int, error) {
	if b.upto-1 == len(b.target) {
		return END_LABEL, nil
	} else {
		return int(b.target[b.upto-1] & 0xFF), nil
	}
}

func (b *BytesRefFSTEnum[T]) getCurrentLabel() (int, error) {
	// return current.bytes[upto] & 0xFF;

	upto := len(b.arcs) - 1
	// current.offset fixed at 1
	return int(b.current[upto] & 0xFF), nil
}

func (b *BytesRefFSTEnum[T]) setCurrentLabel(label int) error {
	b.current = append(b.current, byte(label))
	return nil
}

func (b *BytesRefFSTEnum[T]) grow() error {
	size := b.upto + 1
	if len(b.current) < size {
		growSize := size - len(b.current)
		b.current = append(b.current, make([]byte, growSize)...)
	}
	return nil
}
