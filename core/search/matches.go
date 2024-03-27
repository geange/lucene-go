package search

// Matches
// Reports the positions and optionally offsets of all matching terms in a query for a single document
// To obtain a MatchesIterator for a particular field, call GetMatches(String). Note that you can call
// GetMatches(String) multiple times to retrieve new iterators, but it is not thread-safe.
// 报告单个文档的查询中所有匹配项的位置和可选偏移量，以获取特定字段的匹配迭代器，称为getMatches（String）。
// 注意，可以多次调用getMatches（String）来检索新的迭代器，但它不是线程安全的。
type Matches interface {
	Strings() []string

	// GetMatches
	// Returns a MatchesIterator over the matches for a single field, or null if there are no matches
	// in that field.
	GetMatches(field string) (MatchesIterator, error)

	// GetSubMatches
	// Returns a collection of Matches that make up this instance; if it is not a composite,
	// then this returns an empty list
	GetSubMatches() []Matches
}

type baseMatches struct {
	strs []string
}

var _ Matches = &matchWithNoTerms{}

type matchWithNoTerms struct {
}

func (m *matchWithNoTerms) Strings() []string {
	return nil
}

func (m *matchWithNoTerms) GetMatches(field string) (MatchesIterator, error) {
	return nil, nil
}

func (m *matchWithNoTerms) GetSubMatches() []Matches {
	return nil
}

var _ Matches = &matchForField{}

type matchForField struct {
	field  string
	cached bool
}

func (m *matchForField) Strings() []string {
	//TODO implement me
	panic("implement me")
}

func (m *matchForField) GetMatches(field string) (MatchesIterator, error) {
	//TODO implement me
	panic("implement me")
}

func (m *matchForField) GetSubMatches() []Matches {
	//TODO implement me
	panic("implement me")
}

var _ Matches = &MatchesAnon{}

type MatchesAnon struct {
	FnStrings       func() []string
	FnGetMatches    func(field string) (MatchesIterator, error)
	FnGetSubMatches func() []Matches
}

func (m *MatchesAnon) Strings() []string {
	return m.FnStrings()
}

func (m *MatchesAnon) GetMatches(field string) (MatchesIterator, error) {
	return m.FnGetMatches(field)
}

func (m *MatchesAnon) GetSubMatches() []Matches {
	return m.FnGetSubMatches()
}
