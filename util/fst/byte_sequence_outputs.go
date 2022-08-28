package fst

import (
	"github.com/geange/lucene-go/core/store"
	. "github.com/geange/lucene-go/math"
	. "github.com/geange/lucene-go/util/structure"
)

var _ Outputs[*ByteRef] = &ByteSequenceOutputs[*ByteRef]{}

var (
	ByteSequenceOutputsNoOutput = ByteRef{}
)

type ByteSequenceOutputs[T []byte] struct {
}

func NewByteSequenceOutputs[T ByteRef]() *ByteSequenceOutputs[T] {
	return &ByteSequenceOutputs[T]{}
}

/**

  public BytesRef common(BytesRef output1, BytesRef output2) {
    assert output1 != null;
    assert output2 != null;

    int pos1 = output1.offset;
    int pos2 = output2.offset;
    int stopAt1 = pos1 + Math.min(output1.length, output2.length);
    while(pos1 < stopAt1) {
      if (output1.bytes[pos1] != output2.bytes[pos2]) {
        break;
      }
      pos1++;
      pos2++;
    }

    if (pos1 == output1.offset) {
      // no common prefix
      return NO_OUTPUT;
    } else if (pos1 == output1.offset + output1.length) {
      // output1 is a prefix of output2
      return output1;
    } else if (pos2 == output2.offset + output2.length) {
      // output2 is a prefix of output1
      return output2;
    } else {
      return new BytesRef(output1.bytes, output1.offset, pos1-output1.offset);
    }
  }


*/

func (b *ByteSequenceOutputs[T]) Common(output1, output2 *ByteRef) *ByteRef {
	pos1, pos2 := 0, 0
	stopAt1 := Min(output1.Len(), output2.Len())
	for pos1 < stopAt1 {
		if output1.Bytes[pos1] != output2.Bytes[pos2] {
			break
		}
		pos1++
		pos2++
	}

	if pos1 == 0 {
		return &ByteSequenceOutputsNoOutput
	} else if pos1 == output1.Len() {
		return output1
	} else if pos2 == output2.Len() {
		return output2
	} else {
		return NewByteRef(output1.Bytes[:pos1])
	}
}

func (b *ByteSequenceOutputs[T]) Subtract(output, inc *ByteRef) *ByteRef {
	if inc == &ByteSequenceOutputsNoOutput {
		return &ByteSequenceOutputsNoOutput
	}

	if inc.Len() == output.Len() {
		return &ByteSequenceOutputsNoOutput
	}

	return NewByteRef(output.Bytes[inc.Len():])
}

func (b *ByteSequenceOutputs[T]) Add(prefix, output *ByteRef) *ByteRef {
	if prefix == &ByteSequenceOutputsNoOutput {
		return &ByteSequenceOutputsNoOutput
	} else if output == &ByteSequenceOutputsNoOutput {
		return prefix
	}

	buff := make([]byte, prefix.Len()+output.Len())
	copy(buff, prefix.Bytes)
	copy(buff[prefix.Len():], output.Bytes)
	return NewByteRef(buff)
}

func (b *ByteSequenceOutputs[T]) Write(prefix *ByteRef, out store.DataOutput) error {
	err := out.WriteUvarint(uint64(prefix.Len()))
	if err != nil {
		return err
	}
	return out.WriteBytes(prefix.Bytes)
}

func (b *ByteSequenceOutputs[T]) Read(in store.DataInput) (*ByteRef, error) {
	size, err := in.ReadUvarint()
	if err != nil {
		return nil, err
	}
	if size == 0 {
		return &ByteSequenceOutputsNoOutput, nil
	}

	output := NewByteRef(make([]byte, int(size)))
	err = in.ReadBytes(output.Bytes)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (b *ByteSequenceOutputs[T]) SkipOutput(in store.DataInput) error {
	size, err := in.ReadUvarint()
	if err != nil {
		return err
	}
	return in.SkipBytes(int(size))
}

func (b *ByteSequenceOutputs[T]) GetNoOutput() *ByteRef {
	return &ByteSequenceOutputsNoOutput
}

func (b *ByteSequenceOutputs[T]) OutputToString(output T) string {
	//TODO implement me
	panic("implement me")
}
