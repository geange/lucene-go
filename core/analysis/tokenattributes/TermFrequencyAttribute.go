package tokenattributes

// TermFrequencyAttribute Sets the custom term frequency of a term within one document. If this attribute
// is present in your analysis chain for a given field, that field must be indexed with IndexOptions.DOCS_AND_FREQS.
type TermFrequencyAttribute interface {

	// SetTermFrequency Set the custom term frequency of the current term within one document.
	SetTermFrequency(termFrequency int) error

	// GetTermFrequency Returns the custom term frequency.
	GetTermFrequency() int
}
