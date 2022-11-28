package fst

import (
	"github.com/geange/lucene-go/core/store"
	"github.com/pkg/errors"
)

var _ Outputs = &PositiveIntOutputs{}

var (
	ErrTypeNotInt64 = errors.New("type is not int64")
)

type PositiveIntOutputs struct {
}

func NewPositiveIntOutputs() *PositiveIntOutputs {
	return &PositiveIntOutputs{}
}

func (p *PositiveIntOutputs) Common(output1, output2 any) (any, error) {
	n1, _ := output1.(int64)
	n2, _ := output2.(int64)

	if n1 == 0 || n2 == 0 {
		return nil, nil
	}

	if n1 < n2 {
		return n1, nil
	}
	return n2, nil
}

func (p *PositiveIntOutputs) Subtract(output1, inc any) (any, error) {
	n1, _ := output1.(int64)
	n2, _ := inc.(int64)

	if n1 == n2 {
		return nil, nil
	}

	return n1 - n2, nil
}

func (p *PositiveIntOutputs) Add(prefix, output any) (any, error) {
	n1, _ := prefix.(int64)
	n2, _ := output.(int64)

	return n1 + n2, nil
}

func (p *PositiveIntOutputs) Write(output any, out store.DataOutput) error {
	n1, ok := output.(int64)
	if !ok {
		return errors.Wrap(ErrTypeNotInt64, "output's type not match")
	}
	return out.WriteUvarint(uint64(n1))
}

func (p *PositiveIntOutputs) WriteFinalOutput(output any, out store.DataOutput) error {
	return p.Write(output, out)
}

func (p *PositiveIntOutputs) Read(in store.DataInput) (any, error) {
	num, err := in.ReadUvarint()
	if err != nil {
		return nil, err
	}

	if num == 0 {
		return nil, nil
	}

	return int64(num), nil
}

func (p *PositiveIntOutputs) SkipOutput(in store.DataInput) error {
	//TODO implement me
	panic("implement me")
}

func (p *PositiveIntOutputs) ReadFinalOutput(in store.DataInput) (any, error) {
	return p.Read(in)
}

func (p *PositiveIntOutputs) SkipFinalOutput(in store.DataInput) error {
	//TODO implement me
	panic("implement me")
}

func (p *PositiveIntOutputs) IsNoOutput(v any) bool {
	if v == nil {
		return true
	}

	n, ok := v.(int64)
	if ok {
		return n == 0
	}
	return false
}

var (
	positiveIntOutputsNoOutput = new(any)
)

func (p *PositiveIntOutputs) GetNoOutput() any {
	return positiveIntOutputsNoOutput
}

func (p *PositiveIntOutputs) Merge(first, second any) (any, error) {
	//TODO implement me
	panic("implement me")
}
