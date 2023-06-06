package search

import "reflect"

var MATCH_WITH_NO_TERMS Matches

// FromSubMatches Amalgamate a collection of Matches into a single object
func FromSubMatches(subMatches []Matches) (Matches, error) {
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
			strs := make([]string, 0)
			for _, matches := range sm {
				strs = append(strs, matches.Strings()...)
			}
			return strs
		},
		FnGetMatches: func(field string) (MatchesIterator, error) {
			subIterators := make([]MatchesIterator, 0)
			for _, matches := range sm {
				iterator, err := matches.GetMatches(field)
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
