package analysis

import (
	"github.com/geange/lucene-go/core/util/attribute"
	"github.com/geange/lucene-go/core/util/automaton"
)

// A TokenStream enumerates the sequence of tokens, either from Fields of a Document or from query text.
// This is an abstract class; concrete subclasses are:
// Tokenizer, a TokenStream whose input is a Reader; and
// TokenFilter, a TokenStream whose input is another TokenStream.
//
// TokenStream extends AttributeSource, which provides access to all of the token Attributes for the TokenStream.
// Note that only one instance per AttributeImpl is created and reused for every token. This approach reduces
// object creation and allows local caching of references to the AttributeImpls. See incrementToken() for further details.
// The workflow of the new TokenStream API is as follows:
//
// 1. Instantiation of TokenStream/TokenFilters which add/get attributes to/from the AttributeSource.
// 2. The consumer calls reset().
// 3. The consumer retrieves attributes from the stream and stores local references to all attributes it wants to access.
// 4. The consumer calls incrementToken() until it returns false consuming the attributes after each call.
// 5. The consumer calls end() so that any end-of-stream operations can be performed.
// 6. The consumer calls close() to release any resource when finished using the TokenStream.
//
// To make sure that filters and consumers know which attributes are available, the attributes must be added during
// instantiation. Filters and consumers are not required to check for availability of attributes in incrementToken().
// You can find some example code for the new API in the analysis package level Javadoc.
// Sometimes it is desirable to capture a current state of a TokenStream, e.g., for buffering purposes
// (see CachingTokenFilter, TeeSinkTokenFilter). For this usecase AttributeSourceV2.captureState and
// AttributeSourceV2.restoreState can be used.
// The TokenStream-API in Lucene is based on the decorator pattern. Therefore all non-abstract subclasses must
// be final or have at least a final implementation of incrementToken! This is checked when Java assertions are enabled.
type TokenStream interface {
	AttributeSource() *attribute.Source

	// IncrementToken
	// Consumers (i.e., IndexWriter) use this method to advance the stream to the next token.
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

	// End
	// This method is called by the consumer after the last token has been consumed, after incrementToken()
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

// TokenStreamToAutomaton Consumes a TokenStream and creates an Automaton where the transition labels are UTF8
// bytes (or Unicode code points if unicodeArcs is true) from the TermToBytesRefAttribute. Between tokens we
// insert POS_SEP and for holes we insert HOLE.
type TokenStreamToAutomaton struct {
	preservePositionIncrements bool
	finalOffsetGapAsHole       bool
	unicodeArcs                bool
}

func NewTokenStreamToAutomaton() *TokenStreamToAutomaton {
	return &TokenStreamToAutomaton{preservePositionIncrements: true}
}

// SetPreservePositionIncrements Whether to generate holes in the automaton for missing positions, true by default.
func (r *TokenStreamToAutomaton) SetPreservePositionIncrements(enablePositionIncrements bool) {
	r.preservePositionIncrements = enablePositionIncrements
}

// SetFinalOffsetGapAsHole f true, any final offset gaps will result in adding a position hole.
func (r *TokenStreamToAutomaton) SetFinalOffsetGapAsHole(finalOffsetGapAsHole bool) {
	r.finalOffsetGapAsHole = finalOffsetGapAsHole
}

// SetUnicodeArcs Whether to make transition labels Unicode code points instead of UTF8 bytes, false by default
func (r *TokenStreamToAutomaton) SetUnicodeArcs(unicodeArcs bool) {
	r.unicodeArcs = unicodeArcs
}

// ChangeToken Subclass and implement this if you need to change the token (such as escaping certain bytes)
// before it's turned into a graph.
func (r *TokenStreamToAutomaton) ChangeToken(in []byte) []byte {
	return in
}

const (
	// POS_SEP We create transition between two adjacent tokens.
	POS_SEP = 0x001f

	// HOLE We add this arc to represent a hole.
	HOLE = 0x001e
)

func (r *TokenStreamToAutomaton) ToAutomaton(in TokenStream) (*automaton.Automaton, error) {
	builder := automaton.NewNewBuilder()
	builder.CreateState()

	//in.GetAttributeSource().Add(tokenattr.NewPackedTokenAttributeImp())

	panic("")
}
