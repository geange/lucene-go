package fst

import "github.com/geange/lucene-go/core/store"

var _ Outputs[*Pair[any, any]] = &PairOutputs[any, any]{}

type PairOutputs[A, B any] struct {
	NO_OUTPUT *Pair[A, B]
	outputs1  Outputs[A]
	outputs2  Outputs[B]
}

func (p *PairOutputs[A, B]) Common(pair1, pair2 *Pair[A, B]) *Pair[A, B] {
	return NewPair(
		p.outputs1.Common(pair1.Output1, pair2.Output1),
		p.outputs2.Common(pair1.Output2, pair2.Output2),
	)
}

func (p *PairOutputs[A, B]) Subtract(output, inc *Pair[A, B]) *Pair[A, B] {
	return NewPair(
		p.outputs1.Subtract(output.Output1, inc.Output1),
		p.outputs2.Subtract(output.Output2, inc.Output2),
	)
}

func (p *PairOutputs[A, B]) Add(prefix, output *Pair[A, B]) *Pair[A, B] {
	return NewPair(
		p.outputs1.Add(prefix.Output1, output.Output1),
		p.outputs2.Add(prefix.Output2, output.Output2),
	)
}

func (p *PairOutputs[A, B]) Write(output *Pair[A, B], writer store.DataOutput) error {
	err := p.outputs1.Write(output.Output1, writer)
	if err != nil {
		return err
	}
	return p.outputs2.Write(output.Output2, writer)
}

func (p *PairOutputs[A, B]) Read(in store.DataInput) (*Pair[A, B], error) {
	output1, err := p.outputs1.Read(in)
	if err != nil {
		return nil, err
	}
	output2, err := p.outputs2.Read(in)
	if err != nil {
		return nil, err
	}
	return NewPair(output1, output2), nil
}

func (p *PairOutputs[A, B]) SkipOutput(in store.DataInput) error {
	err := p.outputs1.SkipOutput(in)
	if err != nil {
		return err
	}
	return p.outputs2.SkipOutput(in)
}

func (p *PairOutputs[A, B]) GetNoOutput() *Pair[A, B] {
	return p.NO_OUTPUT
}

func (p *PairOutputs[A, B]) OutputToString(output *Pair[A, B]) string {
	//TODO implement me
	panic("implement me")
}

type Pair[A, B any] struct {
	Output1 A
	Output2 B
}

func NewPair[A any, B any](output1 A, output2 B) *Pair[A, B] {
	return &Pair[A, B]{
		Output1: output1,
		Output2: output2,
	}
}
