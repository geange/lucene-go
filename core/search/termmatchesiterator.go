package search

import (
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/interface/search"
)

var _ search.MatchesIterator = &TermMatchesIterator{}

// TermMatchesIterator
// A MatchesIterator over a single term's postings list
type TermMatchesIterator struct {
	upto  int
	pos   int
	pe    index.PostingsEnum
	query search.Query
}

func NewTermMatchesIterator(query search.Query, pe index.PostingsEnum) (*TermMatchesIterator, error) {
	freq, err := pe.Freq()
	if err != nil {
		return nil, err
	}

	return &TermMatchesIterator{
		pe:    pe,
		query: query,
		upto:  freq,
	}, nil
}

func (t *TermMatchesIterator) Next() (bool, error) {
	upto := t.upto
	t.upto--
	if upto > 0 {
		pos, err := t.pe.NextPosition()
		if err != nil {
			return false, err
		}
		t.pos = pos
		return true, nil
	}
	return false, nil
}

func (t *TermMatchesIterator) StartPosition() int {
	return t.pos
}

func (t *TermMatchesIterator) EndPosition() int {
	return t.pos
}

func (t *TermMatchesIterator) StartOffset() (int, error) {
	return t.pe.StartOffset()
}

func (t *TermMatchesIterator) EndOffset() (int, error) {
	return t.pe.EndOffset()
}

func (t *TermMatchesIterator) GetSubMatches() (search.MatchesIterator, error) {
	return nil, nil
}

func (t *TermMatchesIterator) GetQuery() search.Query {
	return t.query
}
