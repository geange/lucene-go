package core

// TermState Encapsulates all required internal state to position the associated TermsEnum without re-seeking.
// See Also: TermsEnum.seekExact(org.apache.lucene.util.BytesRef, TermState), TermsEnum.termState()
type TermState interface {
}
