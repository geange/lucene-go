package core

// A TokenStream enumerates the sequence of tokens, either from Fields of a Document or from query text.
// This is an abstract class; concrete subclasses are:
// * Tokenizer, a TokenStream whose input is a Reader; and
// * TokenFilter, a TokenStream whose input is another TokenStream.
type TokenStream interface {
	GetAttributeSource() *AttributeSource

	// IncrementToken Consumers (i.e., IndexWriter) use this method to advance the stream to the next token.
	// Implementing classes must implement this method and update the appropriate AttributeImpls with the
	// attributes of the next token.
	//
	// The producer must make no assumptions about the attributes after the method has been returned: the caller
	// may arbitrarily change it. If the producer needs to preserve the state for subsequent calls, it can use
	// captureState to create a copy of the current attribute state.
	//
	// This method is called for every token of a document, so an efficient implementation is crucial for good
	// performance. To avoid calls to addAttribute(Class) and getAttribute(Class), references to all AttributeImpls
	// that this stream uses should be retrieved during instantiation.
	//
	// To ensure that filters and consumers know which attributes are available, the attributes must be added
	// during instantiation. Filters and consumers are not required to check for availability of
	// attributes in incrementToken().
	//
	// Returns: false for end of stream; true otherwise
	IncrementToken() (bool, error)

	// End This method is called by the consumer after the last token has been consumed, after incrementToken()
	// returned false (using the new TokenStream API). Streams implementing the old API should upgrade to use
	// this feature.
	//
	// This method can be used to perform any end-of-stream operations, such as setting the final offset of a
	// stream. The final offset of a stream might differ from the offset of the last token eg in case one or
	// more whitespaces followed after the last token, but a WhitespaceTokenizer was used.
	//
	// Additionally any skipped positions (such as those removed by a stopfilter) can be applied to the
	// position increment, or any adjustment of other attributes where the end-of-stream value may be important.
	// If you override this method, always call super.end().
	// Throws: IOException – If an I/O error occurs
	// 当 TokenStream 被消费完，调用该方法
	End() error

	// Reset This method is called by a consumer before it begins consumption using incrementToken().
	// Resets this stream to a clean state. Stateful implementations must implement this method so that
	// they can be reused, just as if they had been created fresh.
	//
	// If you override this method, always call super.reset(), otherwise some internal state will not be
	// correctly reset (e.g., Tokenizer will throw IllegalStateException on further usage).
	// 在调用 incrementToken() 前调用，清除状态，保证对象可以被复用。
	Reset() error

	// Close Releases resources associated with this stream.
	// If you override this method, always call super.close(), otherwise some internal state will not
	// be correctly reset (e.g., Tokenizer will throw IllegalStateException on reuse).
	// 关闭并释放资源
	Close() error
}
