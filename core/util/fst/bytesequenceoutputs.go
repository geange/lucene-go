package fst

import (
	"context"

	"github.com/geange/lucene-go/core/store"
)

var _ Outputs[[]byte] = &ByteSequenceOutputs{}

type ByteSequenceOutputs struct{}

func NewByteSequenceOutputs() *ByteSequenceOutputs {
	return &ByteSequenceOutputs{}
}

func (b *ByteSequenceOutputs) Common(output1, output2 []byte) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (b *ByteSequenceOutputs) Subtract(output1, inc []byte) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (b *ByteSequenceOutputs) Add(prefix, output []byte) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (b *ByteSequenceOutputs) Merge(first, second []byte) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (b *ByteSequenceOutputs) Write(ctx context.Context, output []byte, out store.DataOutput) error {
	//TODO implement me
	panic("implement me")
}

func (b *ByteSequenceOutputs) WriteFinalOutput(ctx context.Context, output []byte, out store.DataOutput) error {
	//TODO implement me
	panic("implement me")
}

func (b *ByteSequenceOutputs) Read(ctx context.Context, in store.DataInput) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (b *ByteSequenceOutputs) SkipOutput(ctx context.Context, in store.DataInput) error {
	//TODO implement me
	panic("implement me")
}

func (b *ByteSequenceOutputs) ReadFinalOutput(ctx context.Context, in store.DataInput) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (b *ByteSequenceOutputs) SkipFinalOutput(ctx context.Context, in store.DataInput) error {
	//TODO implement me
	panic("implement me")
}

func (b *ByteSequenceOutputs) IsNoOutput(v []byte) bool {
	//TODO implement me
	panic("implement me")
}

func (b *ByteSequenceOutputs) GetNoOutput() []byte {
	//TODO implement me
	panic("implement me")
}

func (b *ByteSequenceOutputs) Equal(b1, b2 []byte) bool {
	//TODO implement me
	panic("implement me")
}

func (b *ByteSequenceOutputs) Hash(v []byte) int64 {
	//TODO implement me
	panic("implement me")
}
