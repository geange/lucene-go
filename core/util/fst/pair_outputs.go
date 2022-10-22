package fst

import (
	"github.com/geange/lucene-go/core/store"
	"reflect"
)

var _ Outputs = &PairOutputs{}

type PairOutputs struct {
	NO_OUTPUT *Pair

	outputs1 Outputs
	outputs2 Outputs
}

func (p *PairOutputs) newPair(a, b any) (*Pair, error) {
	if a == p.outputs1.GetNoOutput() {
		a = p.outputs1.GetNoOutput()
	}
	if b == (p.outputs2.GetNoOutput()) {
		b = p.outputs2.GetNoOutput()
	}

	if a == p.outputs1.GetNoOutput() && b == p.outputs2.GetNoOutput() {
		return p.NO_OUTPUT, nil
	} else {
		pair := NewPair(a, b)
		err := assert(p.valid(pair))
		if err != nil {
			return nil, err
		}
		return pair, nil
	}
}

func (p *PairOutputs) valid(pair *Pair) bool {
	noOutput1 := reflect.DeepEqual(pair.output1, p.outputs1.GetNoOutput())
	noOutput2 := reflect.DeepEqual(pair.output2, p.outputs2.GetNoOutput())

	if noOutput1 && pair.output1 != p.outputs1.GetNoOutput() {
		return false
	}

	if noOutput2 && pair.output2 != p.outputs2.GetNoOutput() {
		return false
	}

	if noOutput1 && noOutput2 {
		if pair != p.NO_OUTPUT {
			return false
		}
		return true
	}
	return true
}

func (p *PairOutputs) Common(output1, output2 any) (any, error) {
	pair1 := output1.(*Pair)
	pair2 := output2.(*Pair)

	if err := assert(p.valid(pair1)); err != nil {
		return nil, err
	}
	if err := assert(p.valid(pair2)); err != nil {
		return nil, err
	}

	common1, err := p.outputs1.Common(pair1.output1, pair2.output1)
	if err != nil {
		return nil, err
	}

	common2, err := p.outputs2.Common(pair1.output2, pair2.output2)
	if err != nil {
		return nil, err
	}
	return p.newPair(common1, common2)
}

func (p *PairOutputs) Subtract(output1, output2 any) (any, error) {
	output := output1.(*Pair)
	inc := output2.(*Pair)

	if err := assert(p.valid(output)); err != nil {
		return nil, err
	}
	if err := assert(p.valid(inc)); err != nil {
		return nil, err
	}

	subtract1, err := p.outputs1.Subtract(output.output1, inc.output1)
	if err != nil {
		return nil, err
	}
	subtract2, err := p.outputs2.Subtract(output.output2, inc.output2)
	if err != nil {
		return nil, err
	}

	return p.newPair(subtract1, subtract2)
}

func (p *PairOutputs) Add(output1, output2 any) (any, error) {
	prefix := output1.(*Pair)
	output := output2.(*Pair)

	if err := assert(p.valid(prefix)); err != nil {
		return nil, err
	}
	if err := assert(p.valid(output)); err != nil {
		return nil, err
	}

	add1, err := p.outputs1.Add(prefix.output1, output.output1)
	if err != nil {
		return nil, err
	}
	add2, err := p.outputs2.Add(prefix.output2, output.output2)
	if err != nil {
		return nil, err
	}

	return p.newPair(add1, add2)
}

func (p *PairOutputs) Write(out any, writer store.DataOutput) error {
	output := out.(*Pair)

	if err := assert(p.valid(output)); err != nil {
		return err
	}
	err := p.outputs1.Write(output.output1, writer)
	if err != nil {
		return err
	}
	return p.outputs2.Write(output.output2, writer)
}

func (p *PairOutputs) WriteFinalOutput(output any, writer store.DataOutput) error {
	return p.Write(output, writer)
}

func (p *PairOutputs) Read(in store.DataInput) (any, error) {
	output1, err := p.outputs1.Read(in)
	if err != nil {
		return nil, err
	}
	output2, err := p.outputs2.Read(in)
	if err != nil {
		return nil, err
	}
	return p.newPair(output1, output2)
}

func (p *PairOutputs) SkipOutput(in store.DataInput) error {
	err := p.outputs1.SkipOutput(in)
	if err != nil {
		return err
	}
	return p.outputs2.SkipOutput(in)
}

func (p *PairOutputs) ReadFinalOutput(in store.DataInput) (any, error) {
	return p.Read(in)
}

func (p *PairOutputs) SkipFinalOutput(in store.DataInput) error {
	return p.SkipOutput(in)
}

func (p *PairOutputs) GetNoOutput() any {
	return p.NO_OUTPUT
}

func (p *PairOutputs) Merge(first, second any) (any, error) {
	//TODO implement me
	panic("implement me")
}

type Pair struct {
	output1 any
	output2 any
}

func NewPair(output1, output2 any) *Pair {
	return &Pair{
		output1: output1,
		output2: output2,
	}
}

func (p *Pair) Equals(other *Pair) bool {
	return reflect.DeepEqual(p, other)
}
