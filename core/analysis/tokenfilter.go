package analysis

import (
	"github.com/geange/lucene-go/core/util/attribute"
)

// A TokenFilter is a TokenStream whose input is another TokenStream.
// This is an abstract class; subclasses must override incrementToken().
// See Also: TokenStream
type TokenFilter interface {
	TokenStream

	End() error
	Reset() error
	Close() error
}

type BaseTokenFilter struct {
	source *attribute.Source
	input  TokenStream
}

func NewBaseTokenFilter(input TokenStream) *BaseTokenFilter {
	return &BaseTokenFilter{
		source: input.AttributeSource(),
		input:  input,
	}
}

func (t *BaseTokenFilter) AttributeSource() *attribute.Source {
	return t.input.AttributeSource()
}

func (t *BaseTokenFilter) End() error {
	return t.input.End()
}

func (t *BaseTokenFilter) Reset() error {
	return t.input.Reset()
}

func (t *BaseTokenFilter) Close() error {
	return t.input.Close()
}
