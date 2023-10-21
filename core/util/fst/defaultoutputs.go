package fst

import (
	"errors"

	"github.com/geange/lucene-go/core/store"
)

var _ Outputs[OutputsValue] = &DefOutputs[OutputsValue]{}

type DefOutputs[T OutputsValue] struct {
	noOutput OutputsValue
}

func NewDefOutputs[T OutputsValue]() DefOutputs[T] {
	return DefOutputs[T]{noOutput: OutputsValue{}}
}

func (d *DefOutputs[T]) Common(output1, output2 OutputsValue) (OutputsValue, error) {
	return OutputsValue{
		LastDocsStart: min(output1.LastDocsStart, output2.LastDocsStart),
		SkipPointer:   min(output1.SkipPointer, output2.SkipPointer),
		DocFreq:       min(output1.DocFreq, output2.DocFreq),
		TotalTermFreq: min(output1.TotalTermFreq, output2.TotalTermFreq),
	}, nil
}

func (d *DefOutputs[T]) Subtract(output1, inc OutputsValue) (OutputsValue, error) {
	return OutputsValue{
		LastDocsStart: output1.LastDocsStart - inc.LastDocsStart,
		SkipPointer:   output1.SkipPointer - inc.SkipPointer,
		DocFreq:       output1.DocFreq - inc.DocFreq,
		TotalTermFreq: output1.TotalTermFreq - inc.TotalTermFreq,
	}, nil
}

func (d *DefOutputs[T]) Add(prefix, output OutputsValue) (OutputsValue, error) {
	return OutputsValue{
		LastDocsStart: prefix.LastDocsStart + output.LastDocsStart,
		SkipPointer:   prefix.SkipPointer + output.SkipPointer,
		DocFreq:       prefix.DocFreq + output.DocFreq,
		TotalTermFreq: prefix.TotalTermFreq + output.TotalTermFreq,
	}, nil
}

func (d *DefOutputs[T]) Write(output OutputsValue, out store.DataOutput) error {
	if err := out.WriteUvarint(uint64(output.LastDocsStart)); err != nil {
		return err
	}
	if err := out.WriteUvarint(uint64(output.SkipPointer)); err != nil {
		return err
	}
	if err := out.WriteUvarint(uint64(output.DocFreq)); err != nil {
		return err
	}
	if err := out.WriteUvarint(uint64(output.TotalTermFreq)); err != nil {
		return err
	}
	return nil
}

func (d *DefOutputs[T]) WriteFinalOutput(output OutputsValue, out store.DataOutput) error {
	return d.Write(output, out)
}

func (d *DefOutputs[T]) Read(in store.DataInput) (OutputsValue, error) {
	value := OutputsValue{}

	if num, err := in.ReadUvarint(); err != nil {
		return d.noOutput, err
	} else {
		value.LastDocsStart = int64(num)
	}

	if num, err := in.ReadUvarint(); err != nil {
		return d.noOutput, err
	} else {
		value.SkipPointer = int64(num)
	}

	if num, err := in.ReadUvarint(); err != nil {
		return d.noOutput, err
	} else {
		value.DocFreq = int64(num)
	}

	if num, err := in.ReadUvarint(); err != nil {
		return d.noOutput, err
	} else {
		value.TotalTermFreq = int64(num)
	}

	return value, nil
}

func (d *DefOutputs[T]) SkipOutput(in store.DataInput) error {
	_, err := d.Read(in)
	return err
}

func (d *DefOutputs[T]) ReadFinalOutput(in store.DataInput) (OutputsValue, error) {
	return d.Read(in)
}

func (d *DefOutputs[T]) SkipFinalOutput(in store.DataInput) error {
	_, err := d.Read(in)
	return err
}

func (d *DefOutputs[T]) IsNoOutput(v OutputsValue) bool {
	return v.LastDocsStart == 0 && v.SkipPointer == 0 && v.DocFreq == 0 && v.TotalTermFreq == 0
}

func (d *DefOutputs[T]) GetNoOutput() OutputsValue {
	return d.noOutput
}

func (d *DefOutputs[T]) Merge(first, second OutputsValue) (OutputsValue, error) {
	return d.noOutput, errors.New("not supported")
}

type OutputsValue struct {
	LastDocsStart int64
	SkipPointer   int64
	DocFreq       int64
	TotalTermFreq int64
}

func NewOutputsValue(lastDocsStart, skipPointer, docFreq, totalTermFreq int64) OutputsValue {
	return OutputsValue{LastDocsStart: lastDocsStart, SkipPointer: skipPointer, DocFreq: docFreq, TotalTermFreq: totalTermFreq}
}
