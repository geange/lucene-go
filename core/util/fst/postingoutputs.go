package fst

import (
	"context"

	"github.com/geange/lucene-go/core/store"
)

var (
	//_ Outputs[PostingOutput] = &PostingOutputs{}

	postingOutputs = newPostingOutputs()
)

type PostingOutputs struct {
	noOutput *PostingOutput
}

func NewPostingOutputs() *PostingOutputs {
	return postingOutputs
}

func newPostingOutputs() *PostingOutputs {
	return &PostingOutputs{
		noOutput: new(PostingOutput),
	}
}

func (p *PostingOutputs) Common(a, b *PostingOutput) (*PostingOutput, error) {
	if a == nil || b == nil {
		return p.noOutput, nil
	}
	return &PostingOutput{
		LastDocsStart: min(a.LastDocsStart, b.LastDocsStart),
		SkipPointer:   min(a.SkipPointer, b.SkipPointer),
		DocFreq:       min(a.DocFreq, b.DocFreq),
		TotalTermFreq: min(a.TotalTermFreq, b.TotalTermFreq),
	}, nil
}

func (p *PostingOutputs) Subtract(a, b *PostingOutput) (*PostingOutput, error) {
	if b == nil {
		return a, nil
	}

	return &PostingOutput{
		LastDocsStart: a.LastDocsStart - b.LastDocsStart,
		SkipPointer:   a.SkipPointer - b.SkipPointer,
		DocFreq:       a.DocFreq - b.DocFreq,
		TotalTermFreq: a.TotalTermFreq - b.TotalTermFreq,
	}, nil
}

func (p *PostingOutputs) Add(a, b *PostingOutput) (*PostingOutput, error) {
	if a == nil {
		return b, nil
	}

	if b == nil {
		return a, nil
	}

	return &PostingOutput{
		LastDocsStart: a.LastDocsStart + b.LastDocsStart,
		SkipPointer:   a.SkipPointer + b.SkipPointer,
		DocFreq:       a.DocFreq + b.DocFreq,
		TotalTermFreq: a.TotalTermFreq + b.TotalTermFreq,
	}, nil
}

func (p *PostingOutputs) Write(ctx context.Context, output *PostingOutput, out store.DataOutput) error {
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

func (p *PostingOutputs) WriteFinalOutput(ctx context.Context, output *PostingOutput, out store.DataOutput) error {
	return p.Write(ctx, output, out)
}

func (p *PostingOutputs) Read(ctx context.Context, in store.DataInput) (*PostingOutput, error) {
	output := &PostingOutput{}
	if num, err := in.ReadUvarint(ctx); err != nil {
		return nil, err
	} else {
		output.LastDocsStart = int64(num)
	}

	if num, err := in.ReadUvarint(ctx); err != nil {
		return nil, err
	} else {
		output.SkipPointer = int64(num)
	}

	if num, err := in.ReadUvarint(ctx); err != nil {
		return nil, err
	} else {
		output.DocFreq = int64(num)
	}

	if num, err := in.ReadUvarint(ctx); err != nil {
		return nil, err
	} else {
		output.TotalTermFreq = int64(num)
	}
	return output, nil
}

func (p *PostingOutputs) SkipOutput(ctx context.Context, in store.DataInput) error {
	_, err := p.Read(ctx, in)
	return err
}

func (p *PostingOutputs) ReadFinalOutput(ctx context.Context, in store.DataInput) (*PostingOutput, error) {
	return p.Read(ctx, in)
}

func (p *PostingOutputs) SkipFinalOutput(ctx context.Context, in store.DataInput) error {
	return p.SkipOutput(ctx, in)
}

func (p *PostingOutputs) GetNoOutput() *PostingOutput {
	return p.noOutput
}

func (p *PostingOutputs) Merge(first, second *PostingOutput) (*PostingOutput, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PostingOutputs) Equal(a, b *PostingOutput) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	return b.LastDocsStart == a.LastDocsStart &&
		b.SkipPointer == a.SkipPointer &&
		b.DocFreq == a.DocFreq &&
		b.TotalTermFreq == a.TotalTermFreq
}

func (p *PostingOutputs) Hash(v *PostingOutput) int64 {
	//TODO implement me
	panic("implement me")
}

func (p *PostingOutputs) IsNoOutput(v *PostingOutput) bool {
	if v == nil {
		return true
	}
	return v.LastDocsStart == 0 &&
		v.SkipPointer == 0 &&
		v.TotalTermFreq == 0 &&
		v.DocFreq == 0
}
