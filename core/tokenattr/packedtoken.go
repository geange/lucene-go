package tokenattr

import (
	"errors"
)

type PackedTokenAttr interface {
	Attribute
	TypeAttr
	PositionIncrAttr
	PositionLengthAttr
	OffsetAttr
	TermFreqAttr
	CharTermAttr
}

// TypeAttr
// A Token's lexical types. The Default value is "word".
type TypeAttr interface {

	// Type
	// Returns this Token's lexical types. Defaults to "word".
	// See Also: setType(String)
	Type() string

	// SetType
	// Set the lexical types.
	// See Also: types()
	SetType(_type string)
}

// PositionIncrAttr
// Determines the position of this token relative to the previous Token in a
// TokenStream, used in phrase searching.
// The default value is one.
// Some common uses for this are:
//   - Set it to zero to put multiple terms in the same position. This is useful if, e.g., a word has multiple
//     stems. Searches for phrases including either stem will match. In this case, all but the first stem's
//     increment should be set to zero: the increment of the first instance should be one. Repeating a token
//     with an increment of zero can also be used to boost the scores of matches on that token.
//   - Set it to values greater than one to inhibit exact phrase matches. If, for example, one does not want
//     phrases to match across removed stop words, then one could build a stop word filter that removes stop
//     words and also sets the increment to the number of stop words removed before each non-stop word.
//     Then exact phrase queries will only match when the terms occur with no intervening stop words.
//
// See Also: org.apache.lucene.index.PostingsEnum
type PositionIncrAttr interface {

	// SetPositionIncrement
	// Set the position increment. The default value is one.
	// positionIncrement: the distance from the prior term
	SetPositionIncrement(positionIncrement int) error

	// GetPositionIncrement
	// Returns the position increment of this Token.
	GetPositionIncrement() int
}

type PositionLengthAttr interface {
	// SetPositionLength
	// Set the position length of this Token.
	// The default value is one.
	// Params: positionLength – how many positions this token spans.
	// Throws: IllegalArgumentException – if positionLength is zero or negative.
	// See Also: getPositionLength()
	SetPositionLength(positionLength int) error

	// GetPositionLength
	// Returns the position length of this Token.
	// See Also: setPositionLength
	GetPositionLength() int
}

// OffsetAttr The start and end character offset of a Token.
type OffsetAttr interface {
	// StartOffset
	// Returns this Token's starting offset, the position of the first character corresponding
	// to this token in the source text.
	// Note that the difference between endOffset() and startOffset() may not be equal to termText.length(),
	// as the term text may have been altered by a stemmer or some other filter.
	// See Also: setOffset(int, int)
	StartOffset() int

	// EndOffset
	// Returns this Token's ending offset, one greater than the position of the last character
	// corresponding to this token in the source text. The length of the token in the source text
	// is (endOffset() - startOffset()).
	// See Also: setOffset(int, int)
	EndOffset() int

	// SetOffset
	// Set the starting and ending offset.
	// Throws: IllegalArgumentException – If startOffset or endOffset are negative, or if startOffset is
	// greater than endOffset
	// See Also: startOffset(), endOffset()
	SetOffset(startOffset, endOffset int) error
}

// TermFreqAttr
// Sets the custom term frequency of a term within one document. If this attribute
// is present in your analysis chain for a given field, that field must be indexed with IndexOptions.DOCS_AND_FREQS.
type TermFreqAttr interface {

	// SetTermFrequency
	// Set the custom term frequency of the current term within one document.
	SetTermFrequency(termFrequency int) error

	// GetTermFrequency
	// Returns the custom term frequency.
	GetTermFrequency() int
}

func NewPackedTokenAttr() PackedTokenAttr {
	return &packedTokenAttr{
		bytesAttr:         newCharTermAttr(),
		startOffset:       0,
		endOffset:         0,
		_type:             DEFAULT_TYPE,
		positionIncrement: 1,
		positionLength:    1,
		termFrequency:     1,
	}
}

func newPackedTokenAttr() *packedTokenAttr {
	return &packedTokenAttr{
		bytesAttr:         newCharTermAttr(),
		startOffset:       0,
		endOffset:         0,
		_type:             DEFAULT_TYPE,
		positionIncrement: 1,
		positionLength:    1,
		termFrequency:     1,
	}
}

var _ PackedTokenAttr = &packedTokenAttr{}

type packedTokenAttr struct {
	*bytesAttr

	startOffset       int
	endOffset         int
	_type             string
	positionIncrement int
	positionLength    int
	termFrequency     int
}

func (p *packedTokenAttr) SetTermFrequency(termFrequency int) error {
	if termFrequency < 1 {
		return errors.New("term frequency must be 1 or greater")
	}
	p.termFrequency = termFrequency
	return nil
}

func (p *packedTokenAttr) GetTermFrequency() int {
	return p.termFrequency
}

func (p *packedTokenAttr) StartOffset() int {
	return p.startOffset
}

func (p *packedTokenAttr) EndOffset() int {
	return p.endOffset
}

func (p *packedTokenAttr) SetOffset(startOffset, endOffset int) error {
	if startOffset < 0 || startOffset > endOffset {
		return errors.New("startOffset must be non-negative, and endOffset must be >= startOffset")
	}
	p.startOffset = startOffset
	p.endOffset = endOffset
	return nil
}

func (p *packedTokenAttr) SetPositionLength(positionLength int) error {
	if positionLength < 1 {
		return errors.New("position length must be 1 or greater")
	}
	p.positionLength = positionLength
	return nil
}

func (p *packedTokenAttr) GetPositionLength() int {
	return p.positionLength
}

func (p *packedTokenAttr) SetPositionIncrement(positionIncrement int) error {
	if positionIncrement < 0 {
		return errors.New("increment must be zero or greater")
	}
	p.positionIncrement = positionIncrement
	return nil
}

func (p *packedTokenAttr) GetPositionIncrement() int {
	return p.positionIncrement
}

func (p *packedTokenAttr) Type() string {
	return p._type
}

func (p *packedTokenAttr) SetType(_type string) {
	p._type = _type
}

func (p *packedTokenAttr) Interfaces() []string {
	values := []string{
		"Type",
		"PositionIncrement",
		"PositionLength",
		"Offset",
		"TermFrequency",
	}
	return append(p.bytesAttr.Interfaces(), values...)
}

func (p *packedTokenAttr) Reset() error {
	p.positionIncrement, p.positionLength = 1, 1
	p.termFrequency = 1
	p.startOffset, p.endOffset = 0, 0
	p._type = "word"
	return p.bytesAttr.Reset()
}

func (p *packedTokenAttr) End() error {
	p.positionIncrement = 0
	return nil
}

func (p *packedTokenAttr) CopyTo(target Attribute) error {
	if impl, ok := target.(*packedTokenAttr); ok {
		impl.startOffset = p.startOffset
		impl.endOffset = p.endOffset
		impl._type = p._type
		impl.positionIncrement = p.positionIncrement
		impl.positionLength = p.positionLength
		impl.termFrequency = p.termFrequency
		return nil
	}
	return errors.New("target is not *packedTokenAttr")
}

func (p *packedTokenAttr) Clone() Attribute {
	return &packedTokenAttr{
		startOffset:       p.startOffset,
		endOffset:         p.endOffset,
		_type:             p._type,
		positionIncrement: p.positionIncrement,
		positionLength:    p.positionLength,
		termFrequency:     p.termFrequency,
	}
}
