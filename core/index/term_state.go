package index

// TermState Encapsulates all required internal state to position the associated TermsEnum without re-seeking.
// See Also: TermsEnum.seekExact(org.apache.lucene.util.BytesRef, TermState), TermsEnum.termState()
type TermState interface {

	// CopyFrom Copies the content of the given TermState to this instance
	// Params: other – the TermState to copy
	CopyFrom(other TermState)
}
