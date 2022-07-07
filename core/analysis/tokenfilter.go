package analysis

import (
	"github.com/geange/lucene-go/core/tokenattributes"
)

// A TokenFilter is a TokenStream whose input is another TokenStream.
// This is an abstract class; subclasses must override incrementToken().
// See Also: TokenStream
type TokenFilter interface {
	TokenStream
}

type TokenFilterImp struct {
	source *tokenattributes.AttributeSource

	input TokenStream
}

func NewTokenFilterImp(input TokenStream) *TokenFilterImp {
	return &TokenFilterImp{
		source: input.AttributeSource(),
		input:  input,
	}
}

func (t *TokenFilterImp) AttributeSource() *tokenattributes.AttributeSource {
	return t.input.AttributeSource()
}

//func (t *TokenFilterImp) IncrementToken() (bool, error) {
//	return t.input.IncrementToken()
//}

func (t *TokenFilterImp) End() error {
	return t.input.End()
}

func (t *TokenFilterImp) Reset() error {
	return t.input.Reset()
}

func (t *TokenFilterImp) Close() error {
	return t.input.Close()
}
