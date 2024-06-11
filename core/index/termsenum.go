package index

import (
	"context"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/util/attribute"
)

// BaseTermsEnum
// A base TermsEnum that adds default implementations for
// * attributes()
// * termState()
// * seekExact(BytesRef)
// * seekExact(BytesRef, TermState)
// In some cases, the default implementation may be slow and consume huge memory, so subclass
// SHOULD have its own implementation if possible.
type BaseTermsEnum struct {
	attrs    *attribute.Source
	seekCeil func(ctx context.Context, text []byte) (index.SeekStatus, error)
}

type BaseTermsEnumConfig struct {
	SeekCeil func(ctx context.Context, text []byte) (index.SeekStatus, error)
}

func NewBaseTermsEnum(cfg *BaseTermsEnumConfig) *BaseTermsEnum {
	return &BaseTermsEnum{
		attrs:    attribute.NewSource(),
		seekCeil: cfg.SeekCeil,
	}
}

func (b *BaseTermsEnum) TermState() (index.TermState, error) {
	return &innerTermState{}, nil
}

func (b *BaseTermsEnum) SeekExact(ctx context.Context, text []byte) (bool, error) {
	status, err := b.seekCeil(ctx, text)
	if err != nil {
		return false, err
	}
	return status == index.SEEK_STATUS_FOUND, nil
}

func (b *BaseTermsEnum) SeekExactExpert(ctx context.Context, term []byte, state index.TermState) error {
	_, err := b.SeekExact(ctx, term)
	return err
}

func (b *BaseTermsEnum) Attributes() *attribute.Source {
	if b.attrs == nil {
		b.attrs = attribute.NewSource()
	}
	return b.attrs
}

type innerTermState struct {
}

func (i *innerTermState) CopyFrom(other index.TermState) {
	panic("implement me")
}

var _ index.TermsEnum = &emptyTermsEnum{}

var EmptyTermsEnum = &emptyTermsEnum{}

type emptyTermsEnum struct {
	atts *attribute.Source
}

func (e *emptyTermsEnum) Next(context.Context) ([]byte, error) {
	return []byte{}, nil
}

func (e *emptyTermsEnum) Attributes() *attribute.Source {
	if e.atts == nil {
		e.atts = attribute.NewSource()
	}
	return e.atts
}

func (e *emptyTermsEnum) SeekExact(ctx context.Context, text []byte) (bool, error) {
	return false, nil
}

func (e *emptyTermsEnum) SeekCeil(ctx context.Context, text []byte) (index.SeekStatus, error) {
	return index.SEEK_STATUS_END, nil
}

func (e *emptyTermsEnum) SeekExactByOrd(ctx context.Context, ord int64) error {
	//TODO implement me
	panic("implement me")
}

func (e *emptyTermsEnum) SeekExactExpert(ctx context.Context, term []byte, state index.TermState) error {
	//TODO implement me
	panic("implement me")
}

func (e *emptyTermsEnum) Term() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (e *emptyTermsEnum) Ord() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (e *emptyTermsEnum) DocFreq() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (e *emptyTermsEnum) TotalTermFreq() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (e *emptyTermsEnum) Postings(reuse index.PostingsEnum, flags int) (index.PostingsEnum, error) {
	//TODO implement me
	panic("implement me")
}

func (e *emptyTermsEnum) Impacts(flags int) (index.ImpactsEnum, error) {
	//TODO implement me
	panic("implement me")
}

func (e *emptyTermsEnum) TermState() (index.TermState, error) {
	//TODO implement me
	panic("implement me")
}
