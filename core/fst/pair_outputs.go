package fst

import (
	"github.com/geange/lucene-go/core/store"
	"reflect"
)

var _ Outputs[Pair[any, any]] = &PairOutputs[any, any]{}

type PairOutputs[A any, B any] struct {
	OutputsImp[Pair[A, B]]

	NO_OUTPUT Pair[A, B]

	outputs1 Outputs[A]
	outputs2 Outputs[B]
}

func (p *PairOutputs[A, B]) Common(pair1, pair2 Pair[A, B]) Pair[A, B] {
	return p.NewPair(
		p.outputs1.Common(pair1.Output1, pair2.Output1),
		p.outputs2.Common(pair1.Output2, pair2.Output2),
	)
}

func (p *PairOutputs[A, B]) Subtract(pair1, pair2 Pair[A, B]) Pair[A, B] {
	return p.NewPair(
		p.outputs1.Subtract(pair1.Output1, pair2.Output1),
		p.outputs2.Subtract(pair1.Output2, pair2.Output2),
	)
}

func (p *PairOutputs[A, B]) Add(pair1, pair2 Pair[A, B]) Pair[A, B] {
	return p.NewPair(
		p.outputs1.Add(pair1.Output1, pair2.Output1),
		p.outputs2.Add(pair1.Output2, pair2.Output2),
	)
}

func (p *PairOutputs[A, B]) Write(output Pair[A, B], writer store.DataOutput) error {
	if err := p.outputs1.Write(output.Output1, writer); err != nil {
		return err
	}
	return p.outputs2.Write(output.Output2, writer)
}

func (p *PairOutputs[A, B]) Read(in store.DataInput) (Pair[A, B], error) {
	output1, err := p.outputs1.Read(in)
	if err != nil {
		return Pair[A, B]{}, err
	}
	output2, err := p.outputs2.Read(in)
	if err != nil {
		return Pair[A, B]{}, err
	}
	return p.NewPair(output1, output2), nil
}

func (p *PairOutputs[A, B]) GetNoOutput() Pair[A, B] {
	return p.NO_OUTPUT
}

func (p *PairOutputs[A, B]) SkipOutput(in store.DataInput) error {
	if err := p.outputs1.SkipOutput(in); err != nil {
		return err
	}
	return p.outputs2.SkipOutput(in)
}

type Pair[A any, B any] struct {
	Output1 A
	Output2 B
}

func (p *PairOutputs[A, B]) NewPair(a A, b B) Pair[A, B] {
	if reflect.DeepEqual(a, p.outputs1.GetNoOutput()) {
		a = p.outputs1.GetNoOutput()
	}

	if reflect.DeepEqual(b, p.outputs2.GetNoOutput()) {
		b = p.outputs2.GetNoOutput()
	}

	if reflect.DeepEqual(a, p.outputs1.GetNoOutput()) && reflect.DeepEqual(b, p.outputs2.GetNoOutput()) {
		return p.NO_OUTPUT
	}

	return Pair[A, B]{a, b}

}
