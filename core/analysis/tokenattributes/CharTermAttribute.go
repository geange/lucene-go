package tokenattributes

// CharTermAttribute The term text of a Token.
type CharTermAttribute interface {
	// Buffer Returns the internal termBuffer character array which you can then directly alter. If the array is
	// too small for your token, use resizeBuffer(int) to increase it. After altering the buffer be sure to call
	// setLength to record the number of valid characters that were placed into the termBuffer.
	Buffer() []rune

	// Append Appends the specified String to this character sequence.
	// The characters of the String argument are appended, in order, increasing the length of this sequence by the
	// length of the argument. If argument is null, then the four characters "null" are appended.
	Append(s string)

	// SetEmpty Sets the length of the termBuffer to zero. Use this method before appending contents
	// using the Appendable interface.
	SetEmpty()
}
