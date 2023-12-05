package index

import "github.com/geange/lucene-go/core/tokenattr"

// BaseTermsEnum A base TermsEnum that adds default implementations for
// * attributes()
// * termState()
// * seekExact(BytesRef)
// * seekExact(BytesRef, TermState)
// In some cases, the default implementation may be slow and consume huge memory, so subclass
// SHOULD have its own implementation if possible.
type BaseTermsEnum struct {
	atts     *tokenattr.AttributeSource
	seekCeil func(text []byte) (SeekStatus, error)
}

type BaseTermsEnumConfig struct {
	SeekCeil func(text []byte) (SeekStatus, error)
}

func NewBaseTermsEnum(cfg *BaseTermsEnumConfig) *BaseTermsEnum {
	return &BaseTermsEnum{
		atts:     tokenattr.NewAttributeSource(),
		seekCeil: cfg.SeekCeil,
	}
}

func (b *BaseTermsEnum) TermState() (TermState, error) {
	return &innerTermState{}, nil
}

func (b *BaseTermsEnum) SeekExact(text []byte) (bool, error) {
	status, err := b.seekCeil(text)
	if err != nil {
		return false, err
	}
	return status == SEEK_STATUS_FOUND, nil
}

func (b *BaseTermsEnum) SeekExactExpert(term []byte, state TermState) error {
	_, err := b.SeekExact(term)
	return err
}

func (b *BaseTermsEnum) Attributes() *tokenattr.AttributeSource {
	if b.atts == nil {
		b.atts = tokenattr.NewAttributeSource()
	}
	return b.atts
}

type innerTermState struct {
}

func (i *innerTermState) CopyFrom(other TermState) {
	panic("implement me")
}
