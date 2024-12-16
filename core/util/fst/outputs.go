package fst

import (
	"context"
	"errors"
	"slices"

	"github.com/geange/lucene-go/core/store"
)

// Outputs Represents the output for an FST, providing the basic algebra required for building and traversing the FST.
// Note that any operation that returns noOutput must return the same singleton object from getNoOutput.
// lucene.experimental
type Outputs[T any] interface {

	// Common Eg common("foobar", "food") -> "foo"
	Common(output1, output2 T) (T, error)

	// Subtract Eg sub("foobar", "foo") -> "bar"
	Subtract(output1, inc T) (T, error)

	// Add Eg add("foo", "bar") -> "foobar"
	Add(prefix, output T) (T, error)

	// Write Encode an output value into a DataOutput.
	Write(output T, out store.DataOutput) error

	// WriteFinalOutput Encode an final node output value into a DataOutput.
	// By default this just calls write(Object, DataOutput).
	WriteFinalOutput(output T, out store.DataOutput) error

	// Read Decode an output value previously written with write(Object, DataOutput).
	Read(in store.DataInput) (T, error)

	// SkipOutput Skip the output; defaults to just calling read and discarding the result.
	SkipOutput(in store.DataInput) error

	// ReadFinalOutput Decode an output value previously written with writeFinalOutput(Object, DataOutput).
	// By default this just calls read(DataInput).
	ReadFinalOutput(in store.DataInput) (T, error)

	// SkipFinalOutput Skip the output previously written with writeFinalOutput;
	// defaults to just calling readFinalOutput and discarding the result.
	SkipFinalOutput(in store.DataInput) error

	IsNoOutput(v T) bool

	GetNoOutput() T

	Merge(first, second T) (T, error)
}

type Output interface {
	Common(v Output) (Output, error)
	Sub(v Output) (Output, error)
	Add(v Output) (Output, error)
	Merge(v Output) (Output, error)
	IsNoOutput() bool
	Equal(v Output) bool // 检查output是否一致
	Hash() int64
}

type OutputManager interface {
	OutputBuilder
	OutputReader
	OutputWriter
}

type OutputBuilder interface {
	EmptyOutput() Output
	New() Output
}

type OutputReader interface {
	// Read Decode an output value previously written with write(Object, DataOutput).
	Read(ctx context.Context, in store.DataInput, v any) error

	// SkipOutput Skip the output; defaults to just calling read and discarding the result.
	SkipOutput(ctx context.Context, in store.DataInput) error

	// ReadFinalOutput Decode an output value previously written with writeFinalOutput(Object, DataOutput).
	// By default this just calls read(DataInput).
	ReadFinalOutput(ctx context.Context, in store.DataInput, v any) error

	// SkipFinalOutput Skip the output previously written with writeFinalOutput;
	// defaults to just calling readFinalOutput and discarding the result.
	SkipFinalOutput(ctx context.Context, in store.DataInput) error
}

type OutputWriter interface {
	// Write Encode an output value into a DataOutput.
	Write(ctx context.Context, out store.DataOutput, v any) error

	// WriteFinalOutput Encode an final node output value into a DataOutput.
	// By default this just calls write(Object, DataOutput).
	WriteFinalOutput(ctx context.Context, out store.DataOutput, v any) error
}

type Int interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

type Ints[T Int] []T

func (r Ints[T]) Common(v Output) (Output, error) {
	items, err := r.check(v)
	if err != nil {
		return nil, err
	}

	res := make(Ints[T], len(r))
	for i := 0; i < len(r); i++ {
		res[i] = min(r[i], items[i])
	}
	return res, nil
}

func (r Ints[T]) Sub(v Output) (Output, error) {
	items, err := r.check(v)
	if err != nil {
		return nil, err
	}

	res := make(Ints[T], len(r))
	for i := 0; i < len(r); i++ {
		res[i] = r[i] - items[i]
	}
	return res, nil
}

func (r Ints[T]) Add(v Output) (Output, error) {
	items, err := r.check(v)
	if err != nil {
		return nil, err
	}

	res := make(Ints[T], len(r))
	for i := 0; i < len(r); i++ {
		res[i] = r[i] + items[i]
	}
	return res, nil
}

func (r Ints[T]) Merge(_ Output) (Output, error) {
	return nil, errors.New("unsupported operation")
}

func (r Ints[T]) IsNoOutput() bool {
	for _, t := range r {
		if t != 0 {
			return false
		}
	}
	return true
}

func (r Ints[T]) Equal(v Output) bool {
	items, err := r.check(v)
	if err != nil {
		return false
	}
	return slices.Equal(r, items)
}

func (r Ints[T]) Hash() int64 {
	prime := 31
	result := 0
	for i := 0; i < len(r); i++ {
		result = prime*result + int(r[i])
	}
	return int64(result)
}

func (r Ints[T]) check(v Output) (Ints[T], error) {
	items, ok := v.(Ints[T])
	if !ok {
		return nil, errors.New("input type not fit")
	}

	if len(items) != len(r) {
		return nil, errors.New("input size not equal")
	}
	return items, nil
}

type IntsManager[T Int] struct {
	size        int
	emptyOutput Output
}

func NewIntsManager[T Int](size int) *IntsManager[T] {
	return &IntsManager[T]{size: size}
}

func (r *IntsManager[T]) EmptyOutput() Output {
	if r.emptyOutput == nil {
		r.emptyOutput = r.New()
	}
	return r.emptyOutput
}

func (r *IntsManager[T]) check(v any) (Ints[T], error) {
	items, ok := v.(Ints[T])
	if !ok {
		return nil, errors.New("input type not fit")
	}
	if len(items) != r.size {
		return nil, errors.New("input size not equal")
	}
	return items, nil
}

func (r *IntsManager[T]) New() Output {
	return make(Ints[T], r.size)
}

func (r *IntsManager[T]) Read(ctx context.Context, in store.DataInput, v any) error {
	items, err := r.check(v)
	if err != nil {
		return err
	}

	for i := 0; i < r.size; i++ {
		num, err := in.ReadUvarint(ctx)
		if err != nil {
			return err
		}
		items[i] = T(num)
	}
	return nil
}

func (r *IntsManager[T]) ReadFinalOutput(ctx context.Context, in store.DataInput, v any) error {
	return r.Read(ctx, in, v)
}

func (r *IntsManager[T]) SkipOutput(ctx context.Context, in store.DataInput) error {
	for i := 0; i < r.size; i++ {
		if _, err := in.ReadUvarint(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (r *IntsManager[T]) SkipFinalOutput(ctx context.Context, in store.DataInput) error {
	return r.SkipOutput(ctx, in)
}

func (r *IntsManager[T]) Write(ctx context.Context, out store.DataOutput, v any) error {
	items, err := r.check(v)
	if err != nil {
		return err
	}
	for _, item := range items {
		if err := out.WriteUvarint(ctx, uint64(item)); err != nil {
			return err
		}
	}
	return nil
}

func (r *IntsManager[T]) WriteFinalOutput(ctx context.Context, out store.DataOutput, v any) error {
	return r.Write(ctx, out, v)
}

var _ Output = &PostingOutput{}

type PostingOutput struct {
	LastDocsStart int64
	SkipPointer   int64
	DocFreq       int64
	TotalTermFreq int64
}

func NewPostingOutput(lastDocsStart, skipPointer, docFreq, totalTermFreq int64) *PostingOutput {
	return &PostingOutput{
		LastDocsStart: lastDocsStart,
		SkipPointer:   skipPointer,
		DocFreq:       docFreq,
		TotalTermFreq: totalTermFreq,
	}
}

func (r *PostingOutput) check(v Output) (*PostingOutput, error) {
	output, ok := v.(*PostingOutput)
	if !ok {
		return nil, errors.New("not *PostingOutput")
	}
	return output, nil
}

func (r *PostingOutput) Common(v Output) (Output, error) {
	output, err := r.check(v)
	if err != nil {
		return nil, err
	}

	return &PostingOutput{
		LastDocsStart: min(r.LastDocsStart, output.LastDocsStart),
		SkipPointer:   min(r.SkipPointer, output.SkipPointer),
		DocFreq:       min(r.DocFreq, output.DocFreq),
		TotalTermFreq: min(r.TotalTermFreq, output.TotalTermFreq),
	}, nil
}

func (r *PostingOutput) Sub(v Output) (Output, error) {
	if v == nil {
		return r, nil
	}

	output, err := r.check(v)
	if err != nil {
		return nil, err
	}

	return &PostingOutput{
		LastDocsStart: r.LastDocsStart - output.LastDocsStart,
		SkipPointer:   r.SkipPointer - output.SkipPointer,
		DocFreq:       r.DocFreq - output.DocFreq,
		TotalTermFreq: r.TotalTermFreq - output.TotalTermFreq,
	}, nil
}

func (r *PostingOutput) Add(v Output) (Output, error) {
	output, err := r.check(v)
	if err != nil {
		return nil, err
	}

	return &PostingOutput{
		LastDocsStart: r.LastDocsStart + output.LastDocsStart,
		SkipPointer:   r.SkipPointer + output.SkipPointer,
		DocFreq:       r.DocFreq + output.DocFreq,
		TotalTermFreq: r.TotalTermFreq + output.TotalTermFreq,
	}, nil
}

func (r *PostingOutput) Merge(v Output) (Output, error) {
	return nil, errors.New("unsupported")
}

func (r *PostingOutput) IsNoOutput() bool {
	if r == nil {
		return true
	}

	return r.LastDocsStart == 0 &&
		r.SkipPointer == 0 &&
		r.DocFreq == 0 &&
		r.TotalTermFreq == 0
}

func (r *PostingOutput) Equal(v Output) bool {
	output, err := r.check(v)
	if err != nil {
		return false
	}
	return r.LastDocsStart == output.LastDocsStart &&
		r.SkipPointer == output.SkipPointer &&
		r.DocFreq == output.DocFreq &&
		r.TotalTermFreq == output.TotalTermFreq
}

func (r *PostingOutput) Hash() int64 {
	return hashInt64(r.LastDocsStart) +
		hashInt64(r.SkipPointer) +
		hashInt64(r.DocFreq) +
		hashInt64(r.TotalTermFreq)
}

func hashInt64(value int64) int64 {
	return value ^ (value >> 32)
}

var _ OutputManager = &PostingOutputManager{}

type PostingOutputManager struct {
	emptyOutput Output
}

func NewPostingOutputManager() *PostingOutputManager {
	return &PostingOutputManager{}
}

func (p *PostingOutputManager) EmptyOutput() Output {
	if p.emptyOutput == nil {
		p.emptyOutput = p.New()
	}
	return p.emptyOutput
}

func (p *PostingOutputManager) New() Output {
	return &PostingOutput{}
}

func (p *PostingOutputManager) check(v any) (*PostingOutput, error) {
	output, ok := v.(*PostingOutput)
	if !ok {
		return nil, errors.New("not *PostingOutput")
	}
	return output, nil
}

func (p *PostingOutputManager) Read(ctx context.Context, in store.DataInput, v any) error {
	output, err := p.check(v)
	if err != nil {
		return err
	}
	if num, err := in.ReadUvarint(ctx); err != nil {
		return err
	} else {
		output.LastDocsStart = int64(num)
	}

	if num, err := in.ReadUvarint(ctx); err != nil {
		return err
	} else {
		output.SkipPointer = int64(num)
	}

	if num, err := in.ReadUvarint(ctx); err != nil {
		return err
	} else {
		output.DocFreq = int64(num)
	}

	if num, err := in.ReadUvarint(ctx); err != nil {
		return err
	} else {
		output.TotalTermFreq = int64(num)
	}
	return nil
}

func (p *PostingOutputManager) SkipOutput(ctx context.Context, in store.DataInput) error {
	return p.Read(ctx, in, &PostingOutput{})
}

func (p *PostingOutputManager) ReadFinalOutput(ctx context.Context, in store.DataInput, v any) error {
	return p.Read(ctx, in, v)
}

func (p *PostingOutputManager) SkipFinalOutput(ctx context.Context, in store.DataInput) error {
	return p.SkipOutput(ctx, in)
}

func (p *PostingOutputManager) Write(ctx context.Context, out store.DataOutput, v any) error {
	output, err := p.check(v)
	if err != nil {
		return err
	}
	if err := out.WriteUvarint(ctx, uint64(output.LastDocsStart)); err != nil {
		return err
	}
	if err := out.WriteUvarint(ctx, uint64(output.SkipPointer)); err != nil {
		return err
	}
	if err := out.WriteUvarint(ctx, uint64(output.DocFreq)); err != nil {
		return err
	}
	if err := out.WriteUvarint(ctx, uint64(output.TotalTermFreq)); err != nil {
		return err
	}
	return nil
}

func (p *PostingOutputManager) WriteFinalOutput(ctx context.Context, out store.DataOutput, v any) error {
	return p.Write(ctx, out, v)
}
