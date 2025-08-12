package fst

import (
	"context"
	"errors"

	"github.com/geange/lucene-go/core/store"
)

var (
	_ Outputs[int64] = &PositiveIntOutputs{}

	positiveIntOutputs = newPositiveIntOutputs()
)

func newPositiveIntOutputs() *PositiveIntOutputs {
	return &PositiveIntOutputs{
		noOutput: new(int64),
	}
}

func NewPositiveIntOutputs() *PositiveIntOutputs {
	return positiveIntOutputs
}

type PositiveIntOutputs struct {
	noOutput *int64
}

func (p *PositiveIntOutputs) Common(a, b int64) (int64, error) {
	return min(a, b), nil
}

func (p *PositiveIntOutputs) Subtract(a, b int64) (int64, error) {
	return a - b, nil
}

func (p *PositiveIntOutputs) Add(prefix, output int64) (int64, error) {
	return prefix + output, nil
}

func (p *PositiveIntOutputs) Write(ctx context.Context, output int64, out store.DataOutput) error {
	return out.WriteUvarint(ctx, uint64(output))
}

func (p *PositiveIntOutputs) WriteFinalOutput(ctx context.Context, output int64, out store.DataOutput) error {
	return p.Write(ctx, output, out)
}

func (p *PositiveIntOutputs) Read(ctx context.Context, in store.DataInput) (int64, error) {
	n, err := in.ReadUvarint(ctx)
	if err != nil {
		return 0, err
	}
	return int64(n), nil
}

func (p *PositiveIntOutputs) SkipOutput(ctx context.Context, in store.DataInput) error {
	if _, err := in.ReadUvarint(ctx); err != nil {
		return err
	}
	return nil
}

func (p *PositiveIntOutputs) ReadFinalOutput(ctx context.Context, in store.DataInput) (int64, error) {
	return p.Read(ctx, in)
}

func (p *PositiveIntOutputs) SkipFinalOutput(ctx context.Context, in store.DataInput) error {
	return p.SkipOutput(ctx, in)
}

func (p *PositiveIntOutputs) GetNoOutput() int64 {
	return *p.noOutput
}

func (p *PositiveIntOutputs) IsNoOutput(v int64) bool {
	return v == 0
}

func (p *PositiveIntOutputs) Merge(first, second int64) (int64, error) {
	return 0, errors.New("implement me")
}

func (p *PositiveIntOutputs) Equal(a, b int64) bool {
	return a == b
}

func (p *PositiveIntOutputs) Hash(v int64) int64 {
	return v
}
