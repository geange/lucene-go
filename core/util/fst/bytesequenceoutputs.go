package fst

import (
	"bytes"
	"context"
	"errors"

	"github.com/geange/lucene-go/core/store"
)

var _ Outputs[[]byte] = &ByteSequenceOutputs{}

type ByteSequenceOutputs struct {
	noOutput []byte
}

func NewByteSequenceOutputs() *ByteSequenceOutputs {
	return &ByteSequenceOutputs{
		noOutput: make([]byte, 0),
	}
}

func (b *ByteSequenceOutputs) Common(output1, output2 []byte) ([]byte, error) {

	pos := 0
	pos2 := 0

	stopAt := min(len(output1), len(output2))
	for i := 0; i < stopAt; i++ {
		if output1[i] != output2[i] {
			break
		}
		pos++
	}

	if pos == 0 {
		// no common prefix
		return b.noOutput, nil
	}

	if pos == len(output1) {
		// output1 is a prefix of output2
		return output1, nil
	}

	if pos2 == len(output2) {
		// output2 is a prefix of output1
		return output2, nil
	}

	return output1[:pos], nil
}

func (b *ByteSequenceOutputs) Subtract(output, inc []byte) ([]byte, error) {
	if (len(inc)) == 0 {
		// no prefix removed
		return output, nil
	}

	if !bytes.HasPrefix(output, inc) {
		return nil, errors.New("output does not start with inc prefix")
	}
	if len(inc) == len(output) {
		// entire output removed
		return b.noOutput, nil
	}
	return bytes.TrimPrefix(output, inc), nil
}

func (b *ByteSequenceOutputs) Add(prefix, output []byte) ([]byte, error) {
	if len(prefix) == 0 {
		return output, nil
	}

	if len(output) == 0 {
		return prefix, nil
	}

	return bytes.Join([][]byte{prefix, output}, []byte{}), nil
}

func (b *ByteSequenceOutputs) Merge(first, second []byte) ([]byte, error) {
	return nil, errors.New("merge is not supported")
}

func (b *ByteSequenceOutputs) Write(ctx context.Context, output []byte, out store.DataOutput) error {
	if err := out.WriteUvarint(ctx, uint64(len(output))); err != nil {
		return err
	}
	if _, err := out.Write(output); err != nil {
		return err
	}
	return nil
}

func (b *ByteSequenceOutputs) WriteFinalOutput(ctx context.Context, output []byte, out store.DataOutput) error {
	return b.Write(ctx, output, out)
}

func (b *ByteSequenceOutputs) Read(ctx context.Context, in store.DataInput) ([]byte, error) {
	num, err := in.ReadUvarint(ctx)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, num)
	if _, err := in.Read(buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func (b *ByteSequenceOutputs) SkipOutput(ctx context.Context, in store.DataInput) error {
	_, err := in.ReadUvarint(ctx)
	return err
}

func (b *ByteSequenceOutputs) ReadFinalOutput(ctx context.Context, in store.DataInput) ([]byte, error) {
	return b.Read(ctx, in)
}

func (b *ByteSequenceOutputs) SkipFinalOutput(ctx context.Context, in store.DataInput) error {
	return b.SkipOutput(ctx, in)
}

func (b *ByteSequenceOutputs) IsNoOutput(v []byte) bool {
	return len(v) == 0
}

func (b *ByteSequenceOutputs) GetNoOutput() []byte {
	return b.noOutput
}

func (b *ByteSequenceOutputs) Equal(b1, b2 []byte) bool {
	return bytes.Equal(b1, b2)
}

func (b *ByteSequenceOutputs) Hash(v []byte) int64 {
	return bytesHashCode(v)
}

func bytesHashCode(bytes []byte) int64 {
	if bytes == nil {
		return 0
	}

	result := int64(1)
	for _, b := range bytes {
		result = 31*result + int64(int8(b))
	}
	return result
}
