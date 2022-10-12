package fst

import (
	"github.com/geange/lucene-go/core/store"
	. "github.com/geange/lucene-go/math"
)

var _ Outputs[[]rune] = &CharSequenceOutputs[[]rune]{}

var (
	NoOutputCharSequence = NewBox([]rune{})
)

type CharSequenceOutputs[T []rune] struct {
}

func (c *CharSequenceOutputs[T]) Common(output1, output2 *Box[[]rune]) *Box[[]rune] {
	pos1, pos2 := 0, 0
	stopAt1 := pos1 + Min(len(output1.V), len(output2.V))
	for pos1 < stopAt1 {
		if output1.V[pos1] != output2.V[pos2] {
			break
		}
		pos1++
		pos2++
	}

	if pos1 == 0 {
		return NoOutputCharSequence
	} else if pos1 == len(output1.V) {
		return output1
	} else if pos2 == len(output2.V) {
		return output2
	} else {
		return NewBox(output1.V[:pos1])
	}
}

func (c *CharSequenceOutputs[T]) Subtract(output, inc *Box[[]rune]) *Box[[]rune] {
	if inc == NoOutputCharSequence {
		return NoOutputCharSequence
	} else if len(inc.V) == len(output.V) {
		return NoOutputCharSequence
	} else {
		return NewBox(output.V[len(inc.V):])
	}
}

func (c *CharSequenceOutputs[T]) Add(prefix, output *Box[[]rune]) *Box[[]rune] {
	if prefix == NoOutputCharSequence {
		return output
	} else if output == NoOutputCharSequence {
		return prefix
	} else {
		result := NewBox(make([]rune, len(prefix.V)+len(output.V)))
		copy(result.V, prefix.V)
		copy(result.V[len(prefix.V):], output.V)
		return result
	}
}

func (c *CharSequenceOutputs[T]) Write(prefix *Box[[]rune], out store.DataOutput) error {
	err := out.WriteUvarint(uint64(len(prefix.V)))
	if err != nil {
		return err
	}
	for _, char := range prefix.V {
		err := out.WriteUvarint(uint64(char))
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *CharSequenceOutputs[T]) WriteFinalOutput(output *Box[[]rune], out store.DataOutput) error {
	return c.Write(output, out)
}

func (c *CharSequenceOutputs[T]) Read(in store.DataInput) (*Box[[]rune], error) {
	size, err := in.ReadUvarint()
	if err != nil {
		return nil, err
	}
	if size == 0 {
		return NoOutputCharSequence, nil
	}

	output := NewBox(make([]rune, int(size)))
	for i := range output.V {
		v, err := in.ReadUvarint()
		if err != nil {
			return nil, err
		}
		output.V[i] = rune(v)
	}
	return output, nil
}

func (c *CharSequenceOutputs[T]) ReadFinalOutput(in store.DataInput) (*Box[[]rune], error) {
	return c.Read(in)
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

func (c *CharSequenceOutputs[T]) SkipFinalOutput(in store.DataInput) error {
	return c.SkipOutput(in)
}

func (c *CharSequenceOutputs[T]) GetNoOutput() *Box[[]rune] {
	return NoOutputCharSequence
}

func (c *CharSequenceOutputs[T]) OutputToString(output *Box[[]rune]) string {
	return ""
}

func (c *CharSequenceOutputs[T]) Merge(first, second *Box[[]rune]) *Box[[]rune] {
	return nil
}
