package analysis

// A TokenFilter is a TokenStream whose input is another TokenStream.
// This is an abstract class; subclasses must override incrementToken().
// See Also: TokenStream
type TokenFilter interface {
}
