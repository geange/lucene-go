package fst

import (
	"github.com/geange/lucene-go/core/store"
	"reflect"
)

type PairAble interface {
	any | int64 | *Pair[int64, int64]
}

type Pair[A, B PairAble] struct {
	Output1 A
	Output2 B
}

func NewPair[A, B PairAble](output1 A, output2 B) *Pair[A, B] {
	return &Pair[A, B]{Output1: output1, Output2: output2}
}

type PairOutputs[A, B PairAble] struct {
	outputs1 Outputs[A]
	outputs2 Outputs[B]
	noOutput *Pair[A, B]
}

func NewPairOutputs[A, B PairAble](outputs1 Outputs[A], outputs2 Outputs[B]) *PairOutputs[A, B] {
	v1 := outputs1.GetNoOutput()
	v2 := outputs2.GetNoOutput()

	return &PairOutputs[A, B]{
		outputs1: outputs1,
		outputs2: outputs2,
		noOutput: NewPair(v1, v2),
	}
}

func (p *PairOutputs[A, B]) Common(output1, output2 *Pair[A, B]) (*Pair[A, B], error) {
	common1, err := p.outputs1.Common(output1.Output1, output2.Output1)
	if err != nil {
		return nil, err
	}
	common2, err := p.outputs2.Common(output1.Output2, output2.Output2)
	if err != nil {
		return nil, err
	}
	return NewPair(common1, common2), nil
}

func (p *PairOutputs[A, B]) Subtract(output1, inc *Pair[A, B]) (*Pair[A, B], error) {
	if inc == nil {
		return output1, nil
	}

	v1, err := p.outputs1.Subtract(output1.Output1, inc.Output1)
	if err != nil {
		return nil, err
	}
	v2, err := p.outputs2.Subtract(output1.Output2, inc.Output2)
	if err != nil {
		return nil, err
	}
	return NewPair(v1, v2), nil
}

func (p *PairOutputs[A, B]) Add(prefix, output *Pair[A, B]) (*Pair[A, B], error) {
	v1, err := p.outputs1.Add(prefix.Output1, output.Output1)
	if err != nil {
		return nil, err
	}
	v2, err := p.outputs2.Add(prefix.Output2, output.Output2)
	if err != nil {
		return nil, err
	}
	return NewPair(v1, v2), nil
}

func (p *PairOutputs[A, B]) Write(output *Pair[A, B], out store.DataOutput) error {
	if err := p.outputs1.Write(output.Output1, out); err != nil {
		return err
	}
	if err := p.outputs2.Write(output.Output2, out); err != nil {
		return err
	}
	return nil
}

func (p *PairOutputs[A, B]) WriteFinalOutput(output *Pair[A, B], out store.DataOutput) error {
	return p.Write(output, out)
}

func (p *PairOutputs[A, B]) Read(in store.DataInput) (*Pair[A, B], error) {
	v1, err := p.outputs1.Read(in)
	if err != nil {
		return nil, err
	}
	v2, err := p.outputs2.Read(in)
	if err != nil {
		return nil, err
	}
	return NewPair(v1, v2), nil
}

func (p *PairOutputs[A, B]) SkipOutput(in store.DataInput) error {
	_, err := p.Read(in)
	return err
}

func (p *PairOutputs[A, B]) ReadFinalOutput(in store.DataInput) (*Pair[A, B], error) {
	return p.Read(in)
}

func (p *PairOutputs[A, B]) SkipFinalOutput(in store.DataInput) error {
	return p.SkipOutput(in)
}

func (p *PairOutputs[A, B]) IsNoOutput(v *Pair[A, B]) bool {
	if v == nil {
		return true
	}

	return reflect.DeepEqual(p.noOutput, v)
}

func (p *PairOutputs[A, B]) GetNoOutput() *Pair[A, B] {
	return p.noOutput
}

func (p *PairOutputs[A, B]) Merge(first, second *Pair[A, B]) (*Pair[A, B], error) {
	//TODO implement me
	panic("implement me")
}
