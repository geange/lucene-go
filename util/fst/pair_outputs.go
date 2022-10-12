package fst

import "github.com/geange/lucene-go/core/store"

var _ Outputs[*Pair[any, any]] = &PairOutputs[any, any]{}

type PairOutputs[A, B any] struct {
	NoOutput *Box[*Pair[A, B]]
	outputs1 Outputs[A]
	outputs2 Outputs[B]
}

func (p PairOutputs[A, B]) Common(output1, output2 *Box[*Pair[A, B]]) *Box[*Pair[A, B]] {
	return NewBox(&Pair[A, B]{
		Output1: p.outputs1.Common(output1.V.Output1, output2.V.Output1),
		Output2: p.outputs2.Common(output1.V.Output2, output2.V.Output2),
	})
}

func (p PairOutputs[A, B]) Subtract(output, inc *Box[*Pair[A, B]]) *Box[*Pair[A, B]] {
	return NewBox(&Pair[A, B]{
		Output1: p.outputs1.Subtract(output.V.Output1, inc.V.Output1),
		Output2: p.outputs2.Subtract(output.V.Output2, inc.V.Output2),
	})
}

func (p PairOutputs[A, B]) Add(prefix, output *Box[*Pair[A, B]]) *Box[*Pair[A, B]] {
	return NewBox(&Pair[A, B]{
		Output1: p.outputs1.Add(prefix.V.Output1, output.V.Output1),
		Output2: p.outputs2.Add(prefix.V.Output2, output.V.Output2),
	})
}

func (p PairOutputs[A, B]) Write(output *Box[*Pair[A, B]], out store.DataOutput) error {
	err := p.outputs1.Write(output.V.Output1, out)
	if err != nil {
		return err
	}
	return p.outputs2.Write(output.V.Output2, out)
}

func (p PairOutputs[A, B]) WriteFinalOutput(output *Box[*Pair[A, B]], out store.DataOutput) error {
	return p.Write(output, out)
}

func (p PairOutputs[A, B]) Read(in store.DataInput) (*Box[*Pair[A, B]], error) {
	output1, err := p.outputs1.Read(in)
	if err != nil {
		return nil, err
	}
	output2, err := p.outputs2.Read(in)
	if err != nil {
		return nil, err
	}

	return NewBox(&Pair[A, B]{
		Output1: output1,
		Output2: output2,
	}), nil
}

func (p PairOutputs[A, B]) ReadFinalOutput(in store.DataInput) (*Box[*Pair[A, B]], error) {
	return p.Read(in)
}

func (p PairOutputs[A, B]) SkipOutput(in store.DataInput) error {
	err := p.outputs1.SkipOutput(in)
	if err != nil {
		return err
	}
	return p.outputs2.SkipOutput(in)
}

func (p PairOutputs[A, B]) SkipFinalOutput(in store.DataInput) error {
	return p.SkipOutput(in)
}

func (p PairOutputs[A, B]) GetNoOutput() *Box[*Pair[A, B]] {
	return p.NoOutput
}

func (p PairOutputs[A, B]) OutputToString(output *Box[*Pair[A, B]]) string {
	return ""
}

func (p PairOutputs[A, B]) Merge(first, second *Box[*Pair[A, B]]) *Box[*Pair[A, B]] {
	return nil
}

type Pair[A, B any] struct {
	Output1 *Box[A]
	Output2 *Box[B]
}

func NewPair[A any, B any](output1 A, output2 B) *Pair[A, B] {
	return &Pair[A, B]{
		Output1: NewBox(output1),
		Output2: NewBox(output2),
	}
}
