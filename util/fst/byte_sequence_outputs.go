package fst

import (
	"github.com/geange/lucene-go/core/store"
	. "github.com/geange/lucene-go/math"
)

var _ Outputs[[]byte] = &ByteSequenceOutputs[[]byte]{}

var (
	NoOutputByteSequence = NewBox([]byte{})
)

type ByteSequenceOutputs[T []byte] struct {
}

func NewByteSequenceOutputs[T []byte]() *ByteSequenceOutputs[T] {
	return &ByteSequenceOutputs[T]{}
}

func (b *ByteSequenceOutputs[T]) Common(output1, output2 *Box[[]byte]) *Box[[]byte] {
	pos1, pos2 := 0, 0
	stopAt1 := Min(len(output1.V), len(output2.V))
	for pos1 < stopAt1 {
		if output1.V[pos1] != output2.V[pos2] {
			break
		}
		pos1++
		pos2++
	}

	if pos1 == 0 {
		return NoOutputByteSequence
	} else if pos1 == len(output1.V) {
		return output1
	} else if pos2 == len(output2.V) {
		return output2
	} else {
		return NewBox(output1.V[:pos1])
	}
}

func (b *ByteSequenceOutputs[T]) Subtract(output, inc *Box[[]byte]) *Box[[]byte] {
	if inc == NoOutputByteSequence {
		return NoOutputByteSequence
	}

	if len(inc.V) == len(output.V) {
		return NoOutputByteSequence
	}

	return NewBox(output.V[len(inc.V):])
}

func (b *ByteSequenceOutputs[T]) Add(prefix, output *Box[[]byte]) *Box[[]byte] {
	if prefix == NoOutputByteSequence {
		return NoOutputByteSequence
	} else if output == NoOutputByteSequence {
		return prefix
	}

	buff := make([]byte, len(prefix.V)+len(output.V))
	copy(buff, prefix.V)
	copy(buff[len(prefix.V):], output.V)
	return NewBox(buff)
}

func (b *ByteSequenceOutputs[T]) Write(prefix *Box[[]byte], out store.DataOutput) error {
	err := out.WriteUvarint(uint64(len(prefix.V)))
	if err != nil {
		return err
	}
	return out.WriteBytes(prefix.V)
}

func (b *ByteSequenceOutputs[T]) WriteFinalOutput(output *Box[[]byte], out store.DataOutput) error {
	return b.Write(output, out)
}

func (b *ByteSequenceOutputs[T]) Read(in store.DataInput) (*Box[[]byte], error) {
	size, err := in.ReadUvarint()
	if err != nil {
		return nil, err
	}
	if size == 0 {
		return NoOutputByteSequence, nil
	}

	buf := make([]byte, int(size))
	err = in.ReadBytes(buf)
	output := NewBox(buf)

	if err != nil {
		return nil, err
	}
	return output, nil
}

func (b *ByteSequenceOutputs[T]) ReadFinalOutput(in store.DataInput) (*Box[[]byte], error) {
	return b.Read(in)
}

func (b *ByteSequenceOutputs[T]) SkipOutput(in store.DataInput) error {
	size, err := in.ReadUvarint()
	if err != nil {
		return err
	}
	return in.SkipBytes(int(size))
}

func (b *ByteSequenceOutputs[T]) SkipFinalOutput(in store.DataInput) error {
	return b.SkipOutput(in)
}

func (b *ByteSequenceOutputs[T]) GetNoOutput() *Box[[]byte] {
	return NoOutputByteSequence
}

func (b *ByteSequenceOutputs[T]) OutputToString(output *Box[[]byte]) string {
	return ""
}

func (b *ByteSequenceOutputs[T]) Merge(first, second *Box[[]byte]) *Box[[]byte] {
	return nil
}
