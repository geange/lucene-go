package search

import (
	"github.com/geange/lucene-go/core/interface/index"
)

type baseMatches struct {
	strs []string
}

var _ index.Matches = &matchWithNoTerms{}

type matchWithNoTerms struct {
}

func (m *matchWithNoTerms) Strings() []string {
	return nil
}

func (m *matchWithNoTerms) GetMatches(field string) (index.MatchesIterator, error) {
	return nil, nil
}

func (m *matchWithNoTerms) GetSubMatches() []index.Matches {
	return nil
}

var _ index.Matches = &matchForField{}

type matchForField struct {
	field  string
	cached bool
}

func (m *matchForField) Strings() []string {
	//TODO implement me
	panic("implement me")
}

func (m *matchForField) GetMatches(field string) (index.MatchesIterator, error) {
	//TODO implement me
	panic("implement me")
}

func (m *matchForField) GetSubMatches() []index.Matches {
	//TODO implement me
	panic("implement me")
}

var _ index.Matches = &MatchesAnon{}

type MatchesAnon struct {
	FnStrings       func() []string
	FnGetMatches    func(field string) (index.MatchesIterator, error)
	FnGetSubMatches func() []index.Matches
}

func (m *MatchesAnon) Strings() []string {
	return m.FnStrings()
}

func (m *MatchesAnon) GetMatches(field string) (index.MatchesIterator, error) {
	return m.FnGetMatches(field)
}

func (m *MatchesAnon) GetSubMatches() []index.Matches {
	return m.FnGetSubMatches()
}
