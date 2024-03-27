package index

import (
	"errors"
)

var (
	EMPTY_TERMSTATE TermState
)

// TermStates Maintains a IndexReader TermState view over IndexReader instances containing a single term.
// The TermStates doesn't track if the given TermState objects are valid, neither if the TermState instances
// refer to the same terms in the associated readers.
type TermStates struct {
	topReaderContextIdentity string
	states                   []TermState
	term                     *Term
	docFreq                  int
	totalTermFreq            int64
}

func NewTermStates(term *Term, context IndexReaderContext) *TermStates {
	size := 1
	if leaves, err := context.Leaves(); err == nil {
		if len(leaves) > 0 {
			size = len(leaves)
		}
	}

	return &TermStates{
		topReaderContextIdentity: context.Identity(),
		states:                   make([]TermState, size),
		term:                     term,
		docFreq:                  0,
		totalTermFreq:            0,
	}
}

func (r *TermStates) WasBuiltFor(context IndexReaderContext) bool {
	return r.topReaderContextIdentity == context.Identity()
}

// BuildTermStates Creates a TermStates from a top-level IndexReaderContext and the given Term.
// This method will lookup the given term in all context's leaf readers and register each of the
// readers containing the term in the returned TermStates using the leaf reader's ordinal.
// Note: the given context must be a top-level context.
// Params: 	needsStats – if true then all leaf contexts will be visited up-front to collect term statistics.
//
//	Otherwise, the TermState objects will be built only when requested
func BuildTermStates(context IndexReaderContext, term *Term, needsStats bool) (*TermStates, error) {
	var perReaderTermState *TermStates
	if needsStats {
		perReaderTermState = NewTermStates(nil, context)

		leaves, err := context.Leaves()
		if err != nil {
			return nil, err
		}

		for _, ctx := range leaves {
			termsEnum, err := loadTermsEnum(ctx, term)
			if err != nil {
				return nil, err
			}
			if termsEnum != nil {
				termState, err := termsEnum.TermState()
				if err != nil {
					return nil, err
				}

				docFreq, err := termsEnum.DocFreq()
				if err != nil {
					return nil, err
				}

				totalTermFreq, err := termsEnum.TotalTermFreq()
				if err != nil {
					return nil, err
				}
				perReaderTermState.Register(termState, ctx.Ord(), docFreq, totalTermFreq)
			}
		}

	} else {
		perReaderTermState = NewTermStates(term, context)
	}

	return perReaderTermState, nil
}

func (r *TermStates) Register(state TermState, ord, docFreq int, totalTermFreq int64) {
	r.Register2(state, ord)
	r.AccumulateStatistics(docFreq, totalTermFreq)
}

// Register2 Expert: Registers and associates a TermState with an leaf ordinal. The leaf ordinal should be
// derived from a IndexReaderContext's leaf ord. On the contrary to register(TermState, int, int, long)
// this method does NOT update term statistics.
func (r *TermStates) Register2(state TermState, ord int) {
	r.states[ord] = state
}

// AccumulateStatistics Expert: Accumulate term statistics.
func (r *TermStates) AccumulateStatistics(docFreq int, totalTermFreq int64) {
	//    assert docFreq >= 0;
	//    assert totalTermFreq >= 0;
	//    assert docFreq <= totalTermFreq;
	r.docFreq += docFreq
	r.totalTermFreq += totalTermFreq
}

// DocFreq Returns the accumulated document frequency of all TermState instances passed to register(TermState, int, int, long).
// Returns:
// the accumulated document frequency of all TermState instances passed to register(TermState, int, int, long).
func (r *TermStates) DocFreq() (int, error) {
	if r.term != nil {
		return 0, errors.New("cannot call docFreq() when needsStats=false")
	}
	return r.docFreq, nil
}

// TotalTermFreq Returns the accumulated term frequency of all TermState instances passed to register(TermState, int, int, long).
// Returns:
// the accumulated term frequency of all TermState instances passed to register(TermState, int, int, long).
func (r *TermStates) TotalTermFreq() (int64, error) {
	if r.term != nil {
		return 0, errors.New("cannot call totalTermFreq() when needsStats=false")
	}
	return r.totalTermFreq, nil
}

func loadTermsEnum(ctx LeafReaderContext, term *Term) (TermsEnum, error) {
	terms, err := ctx.LeafReader().Terms(term.Field())
	if err != nil {
		return nil, err
	}

	if terms != nil {
		termsEnum, err := terms.Iterator()
		if err != nil {
			return nil, err
		}
		ok, err := termsEnum.SeekExact(nil, term.Bytes())
		if err != nil {
			return nil, err
		}

		if ok {
			return termsEnum, nil
		}
	}
	return nil, nil
}

// Get Returns the TermState for a leaf reader context or null if no TermState for the context was registered.
// Params: 	ctx – the LeafReaderContextImpl to get the TermState for.
// Returns: the TermState for the given readers ord or null if no TermState for the reader was
func (r *TermStates) Get(ctx LeafReaderContext) (TermState, error) {
	if r.term == nil {
		return r.states[ctx.Ord()], nil
	}

	if r.states[ctx.Ord()] == nil {
		te, err := loadTermsEnum(ctx, r.term)
		if err != nil {
			return nil, err
		}
		if te == nil {
			r.states[ctx.Ord()] = EMPTY_TERMSTATE
		} else {
			r.states[ctx.Ord()], err = te.TermState()
			if err != nil {
				return nil, err
			}
		}
	}

	if r.states[ctx.Ord()] == EMPTY_TERMSTATE {
		return nil, nil
	}
	return r.states[ctx.Ord()], nil
}
