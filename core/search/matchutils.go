package search

import "reflect"

var MATCH_WITH_NO_TERMS Matches

// MatchesFromSubMatches
// Amalgamate a collection of Matches into a single object
func MatchesFromSubMatches(subMatches []Matches) (Matches, error) {
	sm := make([]Matches, 0)
	for i, match := range subMatches {
		if reflect.DeepEqual(match, MATCH_WITH_NO_TERMS) {
			continue
		}
		sm = append(sm, subMatches[i])
	}

	if len(sm) == 0 {
		return MATCH_WITH_NO_TERMS, nil
	}

	if len(sm) == 1 {
		return sm[0], nil
	}

	return &MatchesAnon{
		FnStrings: func() []string {
			values := make([]string, 0)
			for _, v := range sm {
				values = append(values, v.Strings()...)
			}
			return values
		},
		FnGetMatches: func(field string) (MatchesIterator, error) {
			subIterators := make([]MatchesIterator, 0)
			for _, v := range sm {
				iterator, err := v.GetMatches(field)
				if err != nil {
					return nil, err
				}
				subIterators = append(subIterators, iterator)
			}
			return fromSubIterators(subIterators)
		},
		FnGetSubMatches: func() []Matches {
			return subMatches
		},
	}, nil
}

// MatchesForField
// Create a Matches for a single field
func MatchesForField(field string, mis IOSupplier[MatchesIterator]) Matches {
	// The indirection here, using a Supplier object rather than a MatchesIterator
	// directly, is to allow for multiple calls to Matches.getMatches() to return
	// new iterators.  We still need to call MatchesIteratorSupplier.get() eagerly
	// to work out if we have a hit or not.
	mi, err := mis.Get()
	if err != nil {
		return nil
	}
	if mi == nil {
		return nil
	}

	return &forFieldMatches{
		mis:    mis,
		cached: false,
		field:  field,
		mi:     mi,
	}
}

var _ Matches = &forFieldMatches{}

type forFieldMatches struct {
	mis    IOSupplier[MatchesIterator]
	cached bool
	field  string
	mi     MatchesIterator
}

func (f *forFieldMatches) Strings() []string {
	return []string{f.field}
}

func (f *forFieldMatches) GetMatches(field string) (MatchesIterator, error) {
	if field == f.field {
		return nil, nil
	}
	if f.cached == false {
		return f.mis.Get()
	}
	f.cached = false
	return f.mi, nil
}

func (f *forFieldMatches) GetSubMatches() []Matches {
	return nil
}

type IOSupplier[T any] interface {
	Get() (T, error)
}

type Supplier[T any] interface {
	Get() T
}
