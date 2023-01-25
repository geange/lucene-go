package index

import (
	"github.com/geange/lucene-go/core/util"
	"github.com/geange/lucene-go/core/util/automaton"
)

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
	Intersect(compiled *automaton.CompiledAutomaton, startTerm []byte) (TermsEnum, error)

	// Size Returns the number of terms for this field, or -1 if this measure isn't stored by the codec.
	// Note that, just like other term measures, this measure does not take deleted documents into account.
	Size() (int, error)

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
	GetMin() ([]byte, error)

	// GetMax Returns the largest term (in lexicographic order) in the field. Note that, just like other term
	// measures, this measure does not take deleted documents into account. This returns null when there are no terms.
	GetMax() ([]byte, error)
}

type TermsDefaultConfig struct {
	Iterator func() (TermsEnum, error)
	Size     func() (int, error)
}

type TermsDefault struct {
	Iterator func() (TermsEnum, error)
	Size     func() (int, error)
}

func NewTermsDefault(cfg *TermsDefaultConfig) *TermsDefault {
	return &TermsDefault{
		Iterator: cfg.Iterator,
		Size:     cfg.Size,
	}
}

func (t *TermsDefault) Intersect(compiled *automaton.CompiledAutomaton, startTerm []byte) (TermsEnum, error) {
	// TODO: could we factor out a common interface b/w
	// CompiledAutomaton and FST?  Then we could pass FST there too,
	// and likely speed up resolving terms to deleted docs ... but
	// AutomatonTermsEnum makes this tricky because of its on-the-fly cycle
	// detection

	// TODO: eventually we could support seekCeil/Exact on
	// the returned enum, instead of only being able to seek
	// at the start

	//termsEnum, err := t.Iterator()
	//if err != nil {
	//	return nil, err
	//}
	//
	//if compiled.Type() != automaton.AUTOMATON_TYPE_NORMAL {
	//	return nil, errors.New("please use CompiledAutomaton.getTermsEnum instead")
	//}
	//
	//if len(startTerm) > 0 {
	//	//
	//	//return nAutomatonTermsEnum(termsEnum, compiled);
	//}
	panic("")
}

func (t *TermsDefault) GetMin() ([]byte, error) {
	iterator, err := t.Iterator()
	if err != nil {
		return nil, err
	}
	return iterator.Next()
}

func (t *TermsDefault) GetMax() ([]byte, error) {
	size, err := t.Size()
	if err != nil {
		return nil, err
	}

	if size == 0 {
		return nil, nil
	} else if size >= 0 {
		iterator, err := t.Iterator()
		if err != nil {
			return nil, err
		}
		if err := iterator.SeekExactByOrd(int64(size - 1)); err != nil {
			return nil, err
		}
		return iterator.Term()
	}

	// otherwise: binary search
	iterator, err := t.Iterator()
	if err != nil {
		return nil, err
	}
	v, err := iterator.Next()
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, nil
	}

	scratch := util.NewBytesRefBuilder()
	scratch.AppendByte(0)

	for {
		low := 0
		high := 256

		for low != high {
			mid := (low + high) >> 1
			scratch.SetByteAt(scratch.Length()-1, byte(mid))
			status, err := iterator.SeekCeil(scratch.Get())
			if err != nil {
				return nil, err
			}
			if status == SEEK_STATUS_END {
				// Scratch was too high
				if mid == 0 {
					scratch.SetLength(scratch.Length() - 1)
					return scratch.Get(), nil
				}
			} else {
				// Scratch was too low; there is at least one term
				// still after it:
				if low == mid {
					break
				}
				low = mid
			}
		}

		// Recurse to next digit:
		scratch.SetLength(scratch.Length() + 1)
		scratch.Grow(scratch.Length())
	}
}
