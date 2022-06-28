package core

import "github.com/geange/lucene-go/core/util"

type Terms interface {
	// Iterator Returns an iterator that will step through all terms. This method will not return null.
	Iterator() (TermsEnum, error)

	// Intersect Returns a TermsEnum that iterates over all terms and documents that are accepted by the
	// provided CompiledAutomaton. If the startTerm is provided then the returned enum will only return
	// terms > startTerm, but you still must call next() first to get to the first term. Note that the provided
	// startTerm must be accepted by the automaton.
	// This is an expert low-level API and will only work for NORMAL compiled automata. To handle any compiled
	// automata you should instead use CompiledAutomaton.getTermsEnum instead.
	// NOTE: the returned TermsEnum cannot seek
	Intersect(compiled *CompiledAutomaton, startTerm []byte) (TermsEnum, error)

	// Size Returns the number of terms for this field, or -1 if this measure isn't stored by the codec.
	// Note that, just like other term measures, this measure does not take deleted documents into account.
	Size() (int64, error)

	// GetSumTotalTermFreq Returns the sum of TermsEnum.totalTermFreq for all terms in this field. Note that,
	// just like other term measures, this measure does not take deleted documents into account.
	GetSumTotalTermFreq() (int64, error)

	// GetSumDocFreq Returns the sum of TermsEnum.docFreq() for all terms in this field. Note that,
	// just like other term measures, this measure does not take deleted documents into account.
	GetSumDocFreq() (int64, error)

	// GetDocCount Returns the number of documents that have at least one term for this field. Note that,
	// just like other term measures, this measure does not take deleted documents into account.
	GetDocCount() (int, error)

	// HasFreqs Returns true if documents in this field store per-document term frequency (PostingsEnum.freq).
	HasFreqs() bool

	// HasOffsets Returns true if documents in this field store offsets.
	HasOffsets() bool

	// HasPositions Returns true if documents in this field store positions.
	HasPositions() bool

	// HasPayloads Returns true if documents in this field store payloads.
	HasPayloads() bool

	// GetMin Returns the smallest term (in lexicographic order) in the field. Note that, just like other
	// term measures, this measure does not take deleted documents into account. This returns null when
	// there are no terms.
	GetMin() (*util.BytesRef, error)

	// GetMax Returns the largest term (in lexicographic order) in the field. Note that, just like other term
	// measures, this measure does not take deleted documents into account. This returns null when there are no terms.
	GetMax() (*util.BytesRef, error)
}
