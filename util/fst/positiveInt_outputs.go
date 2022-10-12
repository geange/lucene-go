package fst

import "github.com/geange/lucene-go/core/store"

var _ Outputs[int64] = &PositiveIntOutputs[int64]{}

var (
	NoOutputPositiveIntOutputs = NewBox(int64(0))
)

type PositiveIntOutputs[T int64] struct {
}

func (p *PositiveIntOutputs[T]) Common(output1, output2 *Box[int64]) *Box[int64] {
	if output1 == NoOutputPositiveIntOutputs || output2 == NoOutputPositiveIntOutputs {
		return NoOutputPositiveIntOutputs
	}

	if output1.V < output2.V {
		return output1
	}
	return output2
}

func (p *PositiveIntOutputs[T]) Subtract(output, inc *Box[int64]) *Box[int64] {
	if inc == NoOutputPositiveIntOutputs {
		return output
	} else if output.V == inc.V {
		return NoOutputPositiveIntOutputs
	} else {
		return NewBox(output.V - inc.V)
	}
}

func (p *PositiveIntOutputs[T]) Add(prefix, output *Box[int64]) *Box[int64] {
	if prefix == NoOutputPositiveIntOutputs {
		return output
	} else if output == NoOutputPositiveIntOutputs {
		return prefix
	} else {
		return NewBox(prefix.V + output.V)
	}
}

func (p *PositiveIntOutputs[T]) Write(output *Box[int64], out store.DataOutput) error {
	return out.WriteUvarint(uint64(output.V))
}

func (p *PositiveIntOutputs[T]) WriteFinalOutput(output *Box[int64], out store.DataOutput) error {
	return p.Write(output, out)
}

func (p *PositiveIntOutputs[T]) Read(in store.DataInput) (*Box[int64], error) {
	v, err := in.ReadUvarint()
	if err != nil {
		return nil, err
	}
	if v == 0 {
		return NoOutputPositiveIntOutputs, nil
	}
	return NewBox(int64(v)), nil
}

func (p *PositiveIntOutputs[T]) ReadFinalOutput(in store.DataInput) (*Box[int64], error) {
	return p.Read(in)
}

func (p *PositiveIntOutputs[T]) SkipOutput(in store.DataInput) error {
	return nil
}

func (p *PositiveIntOutputs[T]) SkipFinalOutput(in store.DataInput) error {
	return p.SkipOutput(in)
}

func (p *PositiveIntOutputs[T]) GetNoOutput() *Box[int64] {
	return NoOutputPositiveIntOutputs
}

func (p *PositiveIntOutputs[T]) OutputToString(output *Box[int64]) string {
	return ""
}

func (p *PositiveIntOutputs[T]) Merge(first, second *Box[int64]) *Box[int64] {
	return nil
}
