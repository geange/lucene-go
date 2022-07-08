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

	Input TokenStream
}

func NewTokenFilterImp(input TokenStream) *TokenFilterImp {
	return &TokenFilterImp{
		source: input.AttributeSource(),
		Input:  input,
	}
}

func (t *TokenFilterImp) AttributeSource() *tokenattributes.AttributeSource {
	return t.Input.AttributeSource()
}

//func (t *TokenFilterImp) IncrementToken() (bool, error) {
//	return t.input.IncrementToken()
//}

func (t *TokenFilterImp) End() error {
	return t.Input.End()
}

func (t *TokenFilterImp) Reset() error {
	return t.Input.Reset()
}

func (t *TokenFilterImp) Close() error {
	return t.Input.Close()
}
