package tokenattributes

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
