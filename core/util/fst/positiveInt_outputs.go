package fst

import (
	"github.com/geange/lucene-go/core/store"
)

var _ Outputs[int64] = &PositiveIntOutputs[int64]{}

func NewPositiveIntOutputs[T int64]() *PositiveIntOutputs[T] {
	return &PositiveIntOutputs[T]{}
}

type PositiveIntOutputs[T int64] struct {
	noOutput T
}

func (p *PositiveIntOutputs[int64]) Common(output1, output2 int64) (int64, error) {
	return min(output1, output2), nil
}

func (p *PositiveIntOutputs[int64]) Subtract(output1, inc int64) (int64, error) {
	return output1 - inc, nil
}

func (p *PositiveIntOutputs[int64]) Add(prefix, output int64) (int64, error) {
	return prefix + output, nil
}

func (p *PositiveIntOutputs[int64]) Write(output int64, out store.DataOutput) error {
	return out.WriteUvarint(uint64(output))
}

func (p *PositiveIntOutputs[int64]) WriteFinalOutput(output int64, out store.DataOutput) error {
	return p.Write(output, out)
}

func (p *PositiveIntOutputs[int64]) Read(in store.DataInput) (int64, error) {
	num, err := in.ReadUvarint()
	if err != nil {
		return 0, err
	}
	return int64(num), nil
}

func (p *PositiveIntOutputs[int64]) ReadFinalOutput(in store.DataInput) (int64, error) {
	return p.Read(in)
}

func (p *PositiveIntOutputs[int64]) SkipOutput(in store.DataInput) error {
	_, err := p.Read(in)
	return err
}

func (p *PositiveIntOutputs[int64]) SkipFinalOutput(in store.DataInput) error {
	return p.SkipOutput(in)
}

func (p *PositiveIntOutputs[int64]) IsNoOutput(v int64) bool {
	return v == 0
}

func (p *PositiveIntOutputs[int64]) GetNoOutput() int64 {
	return 0
}

func (p *PositiveIntOutputs[int64]) Merge(first, second int64) (int64, error) {
	//TODO implement me
	panic("implement me")
}

//
//func (p *PositiveIntOutputs) Common(output1, output2 any) (any, error) {
//	n1, n2 := int64(0), int64(0)
//	if n, ok := output1.(int64); ok {
//		n1 = n
//	}
//	if n, ok := output2.(int64); ok {
//		n2 = n
//	}
//
//	if n1 == 0 || n2 == 0 {
//		return positiveIntOutputsNoOutput, nil
//	}
//
//	if n1 < n2 {
//		return n1, nil
//	}
//	return n2, nil
//}
//
//func (p *PositiveIntOutputs) Subtract(output1, inc any) (any, error) {
//	n1, n2 := int64(0), int64(0)
//	if n, ok := output1.(int64); ok {
//		n1 = n
//	}
//	if n, ok := inc.(int64); ok {
//		n2 = n
//	}
//
//	if n1 == n2 {
//		return nil, nil
//	}
//
//	return n1 - n2, nil
//}
//
//func (p *PositiveIntOutputs) Add(prefix, output any) (any, error) {
//	n1, n2 := int64(0), int64(0)
//	if n, ok := prefix.(int64); ok {
//		n1 = n
//	}
//	if n, ok := output.(int64); ok {
//		n2 = n
//	}
//
//	return n1 + n2, nil
//}
//
//func (p *PositiveIntOutputs) Write(output any, out store.DataOutput) error {
//	n := int64(0)
//
//	v, ok := output.(int64)
//	if ok {
//		n = v
//	}
//	return out.WriteUvarint(uint64(n))
//}
//
//func (p *PositiveIntOutputs) WriteFinalOutput(output any, out store.DataOutput) error {
//	return p.Write(output, out)
//}
//
//func (p *PositiveIntOutputs) Read(in store.DataInput) (any, error) {
//	num, err := in.ReadUvarint()
//	if err != nil {
//		return nil, err
//	}
//
//	if num == 0 {
//		return positiveIntOutputsNoOutput, nil
//	}
//
//	return int64(num), nil
//}
//
//func (p *PositiveIntOutputs) SkipOutput(in store.DataInput) error {
//	_, err := p.Read(in)
//	return err
//}
//
//func (p *PositiveIntOutputs) ReadFinalOutput(in store.DataInput) (any, error) {
//	return p.Read(in)
//}
//
//func (p *PositiveIntOutputs) SkipFinalOutput(in store.DataInput) error {
//	return p.SkipOutput(in)
//}
//
//func (p *PositiveIntOutputs) IsNoOutput(v any) bool {
//	if v == positiveIntOutputsNoOutput {
//		return true
//	}
//
//	if v == nil {
//		return true
//	}
//
//	n, ok := v.(int64)
//	if ok {
//		return n == 0
//	}
//	return false
//}
//
//var (
//	positiveIntOutputsNoOutput = new(any)
//)
//
//func (p *PositiveIntOutputs) GetNoOutput() any {
//	return positiveIntOutputsNoOutput
//}
//
//func (p *PositiveIntOutputs) Merge(first, second any) (any, error) {
//	//TODO implement me
//	panic("implement me")
//}
