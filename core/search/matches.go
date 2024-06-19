package search

import "github.com/geange/lucene-go/core/interface/search"

type baseMatches struct {
	strs []string
}

var _ search.Matches = &matchWithNoTerms{}

type matchWithNoTerms struct {
}

func (m *matchWithNoTerms) Strings() []string {
	return nil
}

func (m *matchWithNoTerms) GetMatches(field string) (search.MatchesIterator, error) {
	return nil, nil
}

func (m *matchWithNoTerms) GetSubMatches() []search.Matches {
	return nil
}

var _ search.Matches = &matchForField{}

type matchForField struct {
	field  string
	cached bool
}

func (m *matchForField) Strings() []string {
	//TODO implement me
	panic("implement me")
}

func (m *matchForField) GetMatches(field string) (search.MatchesIterator, error) {
	//TODO implement me
	panic("implement me")
}

func (m *matchForField) GetSubMatches() []search.Matches {
	//TODO implement me
	panic("implement me")
}

var _ search.Matches = &MatchesAnon{}

type MatchesAnon struct {
	FnStrings       func() []string
	FnGetMatches    func(field string) (search.MatchesIterator, error)
	FnGetSubMatches func() []search.Matches
}

func (m *MatchesAnon) Strings() []string {
	return m.FnStrings()
}

func (m *MatchesAnon) GetMatches(field string) (search.MatchesIterator, error) {
	return m.FnGetMatches(field)
}

func (m *MatchesAnon) GetSubMatches() []search.Matches {
	return m.FnGetSubMatches()
}
