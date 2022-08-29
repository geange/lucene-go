package fst

import (
	"github.com/geange/lucene-go/core/store"
	. "github.com/geange/lucene-go/math"
	. "github.com/geange/lucene-go/util/structure"
)

var _ Outputs[*CharsRef] = &CharSequenceOutputs[*CharsRef]{}

var (
	NoOutputCharSequence CharsRef
)

type CharSequenceOutputs[T []rune] struct {
}

func (c *CharSequenceOutputs[T]) Common(output1, output2 *CharsRef) *CharsRef {
	pos1, pos2 := 0, 0
	stopAt1 := pos1 + Min(output1.Len(), output2.Len())
	for pos1 < stopAt1 {
		if output1.Chars[pos1] != output2.Chars[pos2] {
			break
		}
		pos1++
		pos2++
	}

	if pos1 == 0 {
		return &NoOutputCharSequence
	} else if pos1 == output1.Len() {
		return output1
	} else if pos2 == output2.Len() {
		return output2
	} else {
		return NewCharsRef(output1.Chars[:pos1])
	}
}

func (c *CharSequenceOutputs[T]) Subtract(output, inc *CharsRef) *CharsRef {
	if inc == &NoOutputCharSequence {
		return &NoOutputCharSequence
	} else if inc.Len() == output.Len() {
		return &NoOutputCharSequence
	} else {
		return NewCharsRef(output.Chars[inc.Len():])
	}
}

func (c *CharSequenceOutputs[T]) Add(prefix, output *CharsRef) *CharsRef {
	if prefix == &NoOutputCharSequence {
		return output
	} else if output == &NoOutputCharSequence {
		return prefix
	} else {
		result := NewCharsRef(make([]rune, prefix.Len()+output.Len()))
		copy(result.Chars, prefix.Chars)
		copy(result.Chars[prefix.Len():], output.Chars)
		return result
	}
}

func (c *CharSequenceOutputs[T]) Write(prefix *CharsRef, out store.DataOutput) error {
	err := out.WriteUvarint(uint64(prefix.Len()))
	if err != nil {
		return err
	}
	for _, char := range prefix.Chars {
		err := out.WriteUvarint(uint64(char))
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *CharSequenceOutputs[T]) Read(in store.DataInput) (*CharsRef, error) {
	size, err := in.ReadUvarint()
	if err != nil {
		return nil, err
	}
	if size == 0 {
		return &NoOutputCharSequence, nil
	}

	output := NewCharsRef(make([]rune, int(size)))
	for i := range output.Chars {
		v, err := in.ReadUvarint()
		if err != nil {
			return nil, err
		}
		output.Chars[i] = rune(v)
	}
	return output, nil
}

func (c *CharSequenceOutputs[T]) SkipOutput(in store.DataInput) error {
	v, err := in.ReadUvarint()
	if err != nil {
		return err
	}
	for i := 0; i < int(v); i++ {
		_, err := in.ReadUvarint()
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *CharSequenceOutputs[T]) GetNoOutput() *CharsRef {
	return &NoOutputCharSequence
}

func (c *CharSequenceOutputs[T]) OutputToString(output *CharsRef) string {
	//TODO implement me
	panic("implement me")
}
