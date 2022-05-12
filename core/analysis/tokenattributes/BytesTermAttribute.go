package tokenattributes

// BytesTermAttribute This attribute can be used if you have the raw term bytes to be indexed.
// It can be used as replacement for CharTermAttribute, if binary terms should be indexed.
type BytesTermAttribute interface {
	TermToBytesRefAttribute

	// SetBytesRef Sets the BytesRef of the term
	SetBytesRef(bytes []byte) error
}
