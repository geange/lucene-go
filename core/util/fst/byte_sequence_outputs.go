package fst

import (
	"bytes"
	"errors"
	"github.com/geange/lucene-go/core/store"
	. "github.com/geange/lucene-go/math"
)

var (
	ErrDataFormat = errors.New("data format")
	ErrNoOutput   = errors.New("no output")

	_ Outputs = &ByteSequenceOutputs{}
)

// ByteSequenceOutputs An FST Outputs implementation where each output is a sequence of bytes.
// lucene.experimental
type ByteSequenceOutputs struct {
}

func NewByteSequenceOutputs() *ByteSequenceOutputs {
	return &ByteSequenceOutputs{}
}

func (b *ByteSequenceOutputs) Common(output1, output2 any) (any, error) {
	bs1, ok := output1.([]byte)
	if !ok {
		return nil, ErrDataFormat
	}
	bs2, ok := output2.([]byte)
	if !ok {
		return nil, ErrDataFormat
	}

	pos := 0
	stopAt1 := Min(len(bs1), len(bs2))
	for pos < stopAt1 {
		if bs1[pos] != bs2[pos] {
			break
		}
		pos++
	}

	switch pos {
	case 0:
		return nil, ErrNoOutput
	case len(bs1):
		return bs1, nil
	case len(bs2):
		return bs2, nil
	default:
		return bs1[:pos], nil
	}
}

func (b *ByteSequenceOutputs) Subtract(output1, output2 any) (any, error) {
	bs1, ok := output1.([]byte)
	if !ok {
		return nil, ErrDataFormat
	}
	bs2, ok := output2.([]byte)
	if !ok {
		return nil, ErrDataFormat
	}

	if len(bs2) == 0 {
		return bs1, nil
	}

	if bytes.HasPrefix(bs1, bs2) {
		if len(bs1) == len(bs2) {
			return nil, ErrNoOutput
		}
		return bs1[len(bs2):], nil
	}

	return nil, errors.New("subtract error")
}

func (b *ByteSequenceOutputs) Add(output1, output2 any) (any, error) {
	bs1, ok := output1.([]byte)
	if !ok {
		return nil, ErrDataFormat
	}
	bs2, ok := output2.([]byte)
	if !ok {
		return nil, ErrDataFormat
	}

	if len(bs1) == 0 {
		return bs2, nil
	}

	if len(bs2) == 0 {
		return bs1, nil
	}
	return append(bs1, bs2...), nil
}

func (b *ByteSequenceOutputs) Write(output any, out store.DataOutput) error {
	prefix, ok := output.([]byte)
	if !ok {
		return errors.New("output is not []byte")
	}

	err := assert(prefix != nil)
	if err != nil {
		return err
	}
	err = out.WriteUvarint(uint64(len(prefix)))
	if err != nil {
		return err
	}
	return out.WriteBytes(prefix)
}

func (b *ByteSequenceOutputs) WriteFinalOutput(output any, out store.DataOutput) error {
	return b.Write(output, out)
}

func (b *ByteSequenceOutputs) Read(in store.DataInput) (any, error) {
	size, err := in.ReadUvarint()
	if err != nil {
		return nil, err
	}

	if size == 0 {
		return nil, ErrNoOutput
	}

	output := make([]byte, size)
	err = in.ReadBytes(output)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (b *ByteSequenceOutputs) SkipOutput(in store.DataInput) error {
	size, err := in.ReadUvarint()
	if err != nil {
		return err
	}
	if size != 0 {
		return in.SkipBytes(int(size))
	}
	return nil
}

func (b *ByteSequenceOutputs) ReadFinalOutput(in store.DataInput) (any, error) {
	return b.Read(in)
}

func (b *ByteSequenceOutputs) SkipFinalOutput(in store.DataInput) error {
	return b.SkipOutput(in)
}

func (b *ByteSequenceOutputs) GetNoOutput() any {
	return ErrNoOutput
}

func (b *ByteSequenceOutputs) Merge(first, second any) (any, error) {
	//TODO implement me
	panic("implement me")
}

var ByteSequenceOutputsSingleton = new(ByteSequenceOutputs)

func GetSingleton() *ByteSequenceOutputs {
	return ByteSequenceOutputsSingleton
}
