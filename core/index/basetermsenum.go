package index

import (
	"context"
	"github.com/geange/lucene-go/core/tokenattr"
)

// BaseTermsEnum A base TermsEnum that adds default implementations for
// * attributes()
// * termState()
// * seekExact(BytesRef)
// * seekExact(BytesRef, TermState)
// In some cases, the default implementation may be slow and consume huge memory, so subclass
// SHOULD have its own implementation if possible.
type BaseTermsEnum struct {
	attrs    *tokenattr.AttributeSource
	seekCeil func(ctx context.Context, text []byte) (SeekStatus, error)
}

type BaseTermsEnumConfig struct {
	SeekCeil func(ctx context.Context, text []byte) (SeekStatus, error)
}

func NewBaseTermsEnum(cfg *BaseTermsEnumConfig) *BaseTermsEnum {
	return &BaseTermsEnum{
		attrs:    tokenattr.NewAttributeSource(),
		seekCeil: cfg.SeekCeil,
	}
}

func (b *BaseTermsEnum) TermState() (TermState, error) {
	return &innerTermState{}, nil
}

func (b *BaseTermsEnum) SeekExact(ctx context.Context, text []byte) (bool, error) {
	status, err := b.seekCeil(ctx, text)
	if err != nil {
		return false, err
	}
	return status == SEEK_STATUS_FOUND, nil
}

func (b *BaseTermsEnum) SeekExactExpert(ctx context.Context, term []byte, state TermState) error {
	_, err := b.SeekExact(ctx, term)
	return err
}

func (b *BaseTermsEnum) Attributes() *tokenattr.AttributeSource {
	if b.attrs == nil {
		b.attrs = tokenattr.NewAttributeSource()
	}
	return b.attrs
}

type innerTermState struct {
}

func (i *innerTermState) CopyFrom(other TermState) {
	panic("implement me")
}
