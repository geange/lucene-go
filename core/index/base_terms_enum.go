package index

import "github.com/geange/lucene-go/core/tokenattributes"

// BaseTermsEnum A base TermsEnum that adds default implementations for
// * attributes()
// * termState()
// * seekExact(BytesRef)
// * seekExact(BytesRef, TermState)
// In some cases, the default implementation may be slow and consume huge memory, so subclass
// SHOULD have its own implementation if possible.
type BaseTermsEnum interface {
	TermsEnum
}

type BaseTermsEnumImp struct {
	atts *tokenattributes.AttributeSource
}

func (b *BaseTermsEnumImp) TermState() (TermState, error) {
	panic("")
}

func (b *BaseTermsEnumImp) SeekExact(text []byte) (bool, error) {
	panic("")
}

func (b *BaseTermsEnumImp) SeekExactExpert(term []byte, state TermState) error {
	panic("")
}

func (b *BaseTermsEnumImp) Attributes() *tokenattributes.AttributeSource {
	return b.atts
}

type innerTermState struct {
}

func (i *innerTermState) CopyFrom(other TermState) {
	panic("implement me")
}
