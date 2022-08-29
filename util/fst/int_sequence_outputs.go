package fst

import (
	"github.com/geange/lucene-go/core/store"
	. "github.com/geange/lucene-go/math"
	. "github.com/geange/lucene-go/util/structure"
)

var (
	NoOutputIntSequence IntsRef
)

var _ Outputs[*IntsRef] = &IntSequenceOutputs[*IntsRef]{}

type IntSequenceOutputs[T []int] struct {
}

func (i *IntSequenceOutputs[T]) Common(output1, output2 *IntsRef) *IntsRef {
	pos1, pos2 := 0, 0
	stopAt1 := pos1 + Min(output1.Len(), output2.Len())
	for pos1 < stopAt1 {
		if output1.Ints[pos1] != output2.Ints[pos2] {
			break
		}
		pos1++
		pos2++
	}

	if pos1 == 0 {
		return &NoOutputIntSequence
	} else if pos1 == output1.Len() {
		return output1
	} else if pos2 == output2.Len() {
		return output2
	} else {
		return NewIntsRef(output1.Ints[:pos1])
	}
}

func (i *IntSequenceOutputs[T]) Subtract(output, inc *IntsRef) *IntsRef {
	if inc == &NoOutputIntSequence {
		return &NoOutputIntSequence
	} else if inc.Len() == output.Len() {
		return &NoOutputIntSequence
	} else {
		return NewIntsRef(output.Ints[inc.Len():])
	}
}

func (i *IntSequenceOutputs[T]) Add(prefix, output *IntsRef) *IntsRef {
	if prefix == &NoOutputIntSequence {
		return output
	} else if output == &NoOutputIntSequence {
		return prefix
	} else {
		result := NewIntsRef(make([]int, prefix.Len()+output.Len()))
		copy(result.Ints, prefix.Ints)
		copy(result.Ints[prefix.Len():], output.Ints)
		return result
	}
}

func (i *IntSequenceOutputs[T]) Write(prefix *IntsRef, out store.DataOutput) error {
	err := out.WriteUvarint(uint64(prefix.Len()))
	if err != nil {
		return err
	}
	for _, char := range prefix.Ints {
		err := out.WriteUvarint(uint64(char))
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *IntSequenceOutputs[T]) Read(in store.DataInput) (*IntsRef, error) {
	size, err := in.ReadUvarint()
	if err != nil {
		return nil, err
	}
	if size == 0 {
		return &NoOutputIntSequence, nil
	}

	output := NewIntsRef(make([]int, int(size)))
	for i := range output.Ints {
		v, err := in.ReadUvarint()
		if err != nil {
			return nil, err
		}
		output.Ints[i] = int(v)
	}
	return output, nil
}

func (i *IntSequenceOutputs[T]) SkipOutput(in store.DataInput) error {
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

func (i *IntSequenceOutputs[T]) GetNoOutput() *IntsRef {
	return &NoOutputIntSequence
}

func (i *IntSequenceOutputs[T]) OutputToString(output *IntsRef) string {
	//TODO implement me
	panic("implement me")
}
