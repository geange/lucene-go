package core

const (
	ClassBytesTerm         = "BytesTerm"
	ClassCharTerm          = "CharTerm"
	ClassOffset            = "Offset"
	ClassPositionIncrement = "PositionIncrement"
	ClassPayload           = "Payload"
	ClassPositionLength    = "PositionLength"
	ClassTermFrequency     = "TermFrequency"
	ClassTermToBytesRef    = "TermToBytesRef"
)

// TypeAttribute A Token's lexical type. The Default value is "word".
type TypeAttribute interface {

	// Type Returns this Token's lexical type. Defaults to "word".
	//See Also: setType(String)
	Type() string

	// SetType Set the lexical type.
	// See Also: type()
	SetType(_type string)
}

const (
	DEFAULT_TYPE = "word"
)

// BytesTermAttribute This attribute can be used if you have the raw term bytes to be indexed.
// It can be used as replacement for CharTermAttribute, if binary terms should be indexed.
type BytesTermAttribute interface {
	TermToBytesRefAttribute

	// SetBytesRef Sets the BytesRef of the term
	SetBytesRef(bytes []byte) error
}

// CharTermAttribute The term text of a Token.
type CharTermAttribute interface {
	// Buffer Returns the internal termBuffer character array which you can then directly alter. If the array is
	// too small for your token, use resizeBuffer(int) to increase it. After altering the buffer be sure to call
	// setLength to record the number of valid characters that were placed into the termBuffer.
	Buffer() []rune

	// ResizeBuffer Grows the termBuffer to at least size newSize, preserving the existing content.
	// Params: newSize – minimum size of the new termBuffer
	// Returns: newly created termBuffer with length >= newSize
	ResizeBuffer(newSize int)

	// SetLength Set number of valid characters (length of the term) in the termBuffer array.
	// Use this to truncate the termBuffer or to synchronize with external manipulation of the termBuffer.
	// Note: to grow the size of the array, use resizeBuffer(int) first.
	// Params: length – the truncated length
	//SetLength(length int)

	// Append Appends the specified String to this character sequence.
	// The characters of the String argument are appended, in order, increasing the length of this sequence by the
	// length of the argument. If argument is null, then the four characters "null" are appended.
	Append(s string)

	AppendRune(r rune)

	// SetEmpty Sets the length of the termBuffer to zero. Use this method before appending contents
	// using the Appendable interface.
	SetEmpty()
}

// TermToBytesRefAttribute This attribute is requested by TermsHashPerField to index the contents.
// This attribute can be used to customize the final byte[] encoding of terms.
// Consumers of this attribute call getBytesRef() for each term. Example:
//
//     final TermToBytesRefAttribute termAtt = tokenStream.getAttribute(TermToBytesRefAttribute.class);
//
//     while (tokenStream.incrementToken() {
//       final BytesRef bytes = termAtt.getBytesRef();
//
//       if (isInteresting(bytes)) {
//
//         // because the bytes are reused by the attribute (like CharTermAttribute's char[] buffer),
//         // you should make a copy if you need persistent access to the bytes, otherwise they will
//         // be rewritten across calls to incrementToken()
//
//         doSomethingWith(BytesRef.deepCopyOf(bytes));
//       }
//     }
//     ...
type TermToBytesRefAttribute interface {
	GetBytesRef() []byte
}

// TermFrequencyAttribute Sets the custom term frequency of a term within one document. If this attribute
// is present in your analysis chain for a given field, that field must be indexed with IndexOptions.DOCS_AND_FREQS.
type TermFrequencyAttribute interface {

	// SetTermFrequency Set the custom term frequency of the current term within one document.
	SetTermFrequency(termFrequency int) error

	// GetTermFrequency Returns the custom term frequency.
	GetTermFrequency() int
}

type PositionLengthAttribute interface {
	// SetPositionLength Set the position length of this Token.
	// The default value is one.
	// Params: positionLength – how many positions this token spans.
	// Throws: IllegalArgumentException – if positionLength is zero or negative.
	// See Also: getPositionLength()
	SetPositionLength(positionLength int) error

	// GetPositionLength Returns the position length of this Token.
	// See Also: setPositionLength
	GetPositionLength() int
}

// PositionIncrementAttribute Determines the position of this token relative to the previous Token in a
// TokenStream, used in phrase searching.
// The default value is one.
// Some common uses for this are:
// * Set it to zero to put multiple terms in the same position. This is useful if, e.g., a word has multiple
//   stems. Searches for phrases including either stem will match. In this case, all but the first stem's
//   increment should be set to zero: the increment of the first instance should be one. Repeating a token
//   with an increment of zero can also be used to boost the scores of matches on that token.
// * Set it to values greater than one to inhibit exact phrase matches. If, for example, one does not want
//   phrases to match across removed stop words, then one could build a stop word filter that removes stop
//   words and also sets the increment to the number of stop words removed before each non-stop word.
//   Then exact phrase queries will only match when the terms occur with no intervening stop words.
// See Also: org.apache.lucene.index.PostingsEnum
type PositionIncrementAttribute interface {

	// SetPositionIncrement Set the position increment. The default value is one.
	// Params: positionIncrement – the distance from the prior term
	// Throws: IllegalArgumentException – if positionIncrement is negative.
	// See Also: getPositionIncrement()
	SetPositionIncrement(positionIncrement int) error

	// GetPositionIncrement Returns the position increment of this Token.
	// See Also: setPositionIncrement(int)
	GetPositionIncrement() int
}

// PayloadAttribute The payload of a Token.
// The payload is stored in the index at each position, and can be used to influence scoring when using
// Payload-based queries.
// NOTE: because the payload will be stored at each position, it's usually best to use the minimum number of
// bytes necessary. Some codec implementations may optimize payload storage when all payloads have the same length.
// See Also: org.apache.lucene.index.PostingsEnum
type PayloadAttribute interface {
	// GetPayload Returns this Token's payload.
	// See Also: setPayload(BytesRef)
	GetPayload() []byte

	// SetPayload Sets this Token's payload.
	// See Also: getPayload()
	SetPayload(payload []byte) error
}

// OffsetAttribute The start and end character offset of a Token.
type OffsetAttribute interface {
	// StartOffset Returns this Token's starting offset, the position of the first character corresponding
	// to this token in the source text.
	// Note that the difference between endOffset() and startOffset() may not be equal to termText.length(),
	// as the term text may have been altered by a stemmer or some other filter.
	// See Also: setOffset(int, int)
	StartOffset() int

	// EndOffset Returns this Token's ending offset, one greater than the position of the last character
	// corresponding to this token in the source text. The length of the token in the source text
	// is (endOffset() - startOffset()).
	// See Also: setOffset(int, int)
	EndOffset() int

	// SetOffset Set the starting and ending offset.
	// Throws: IllegalArgumentException – If startOffset or endOffset are negative, or if startOffset is
	// greater than endOffset
	// See Also: startOffset(), endOffset()
	SetOffset(startOffset, endOffset int) error
}
