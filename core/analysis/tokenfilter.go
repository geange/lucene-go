package analysis

import (
	"github.com/geange/lucene-go/core/util"
)

// A TokenFilter is a TokenStream whose input is another TokenStream.
// This is an abstract class; subclasses must override incrementToken().
// See Also: TokenStream
type TokenFilter interface {
	TokenStream
}

type TokenFilterIMP struct {
	sourceV1 *util.AttributeSourceV1

	input TokenStream
}

func NewTokenFilterIMP(input TokenStream) *TokenFilterIMP {
	return &TokenFilterIMP{
		input: input,
	}
}

func (t *TokenFilterIMP) GetAttributeSource() *util.AttributeSource {
	return t.input.GetAttributeSource()
}

func (t *TokenFilterIMP) AttributeSource() *util.AttributeSourceV1 {
	return t.input.AttributeSource()
}

//func (t *TokenFilterIMP) IncrementToken() (bool, error) {
//	return t.input.IncrementToken()
//}

func (t *TokenFilterIMP) End() error {
	return t.input.End()
}

func (t *TokenFilterIMP) Reset() error {
	return t.input.Reset()
}

func (t *TokenFilterIMP) Close() error {
	return t.input.Close()
}
