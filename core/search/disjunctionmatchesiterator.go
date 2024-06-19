package search

import (
	"errors"
	index2 "github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/interface/search"
	"io"

	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/util/bytesref"
	"github.com/geange/lucene-go/core/util/structure"
)

var _ search.MatchesIterator = &DisjunctionMatchesIterator{}

type DisjunctionMatchesIterator struct {
	queue   *structure.PriorityQueue[search.MatchesIterator]
	started bool
}

func (d *DisjunctionMatchesIterator) Next() (bool, error) {
	if d.started == false {
		d.started = true
		return d.started, nil
	}
	next, err := d.queue.Top().Next()
	if err != nil {
		return false, err
	}
	if next == false {
		_, err := d.queue.Pop()
		if err != nil {
			return false, err
		}
	}
	if d.queue.Size() > 0 {
		d.queue.UpdateTop()
		return true, nil
	}
	return false, nil
}

func (d *DisjunctionMatchesIterator) StartPosition() int {
	return d.queue.Top().StartPosition()
}

func (d *DisjunctionMatchesIterator) EndPosition() int {
	return d.queue.Top().EndPosition()
}

func (d *DisjunctionMatchesIterator) StartOffset() (int, error) {
	return d.queue.Top().StartOffset()
}

func (d *DisjunctionMatchesIterator) EndOffset() (int, error) {
	return d.queue.Top().EndOffset()
}

func (d *DisjunctionMatchesIterator) GetSubMatches() (search.MatchesIterator, error) {
	return d.queue.Top().GetSubMatches()
}

func (d *DisjunctionMatchesIterator) GetQuery() search.Query {
	return d.queue.Top().GetQuery()
}

func fromSubIterators(mis []search.MatchesIterator) (search.MatchesIterator, error) {
	if len(mis) == 0 {
		return nil, nil
	}
	if len(mis) == 1 {
		return mis[0], nil
	}
	return newDisjunctionMatchesIterator(mis)
}

func newDisjunctionMatchesIterator(matches []search.MatchesIterator) (search.MatchesIterator, error) {
	queue := structure.NewPriorityQueue[search.MatchesIterator](len(matches), func(a, b search.MatchesIterator) bool {
		return a.StartPosition() < b.StartPosition() ||
			(a.StartPosition() == b.StartPosition() && a.EndPosition() < b.EndPosition()) ||
			(a.StartPosition() == b.StartPosition() && a.EndPosition() == b.EndPosition())
	})

	for _, mi := range matches {
		if ok, _ := mi.Next(); ok {
			queue.Add(mi)
		}
	}
	return &DisjunctionMatchesIterator{
		queue:   queue,
		started: false,
	}, nil
}

// FromTermsEnumMatchesIterator
// Create a DisjunctionMatchesIterator over a list of terms extracted from a BytesRefIterator
// Only terms that have at least one match in the given document will be included
func FromTermsEnumMatchesIterator(context index2.LeafReaderContext, doc int, query search.Query,
	field string, terms bytesref.BytesIterator) (search.MatchesIterator, error) {

	t, err := context.Reader().(index2.LeafReader).Terms(field)
	if err != nil {
		return nil, err
	}
	te, err := t.Iterator()
	if err != nil {
		return nil, err
	}

	var reuse index2.PostingsEnum

	for {
		term, err := terms.Next(nil)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		ok, _ := te.SeekExact(nil, term)
		if ok {
			pe, err := te.Postings(reuse, index.POSTINGS_ENUM_OFFSETS)
			if err != nil {
				return nil, err
			}
			if v, _ := pe.Advance(doc); v == doc {
				iterator, err := NewTermMatchesIterator(query, pe)
				if err != nil {
					return nil, err
				}
				return newTermsEnumDisjunctionMatchesIterator(
					iterator, terms, te, doc, query), nil
			} else {
				reuse = pe
			}
		}
	}
	return nil, nil
}

var _ search.MatchesIterator = &termsEnumDisjunctionMatchesIterator{}

type termsEnumDisjunctionMatchesIterator struct {
	first search.MatchesIterator
	terms bytesref.BytesIterator
	te    index2.TermsEnum
	doc   int
	query search.Query
	it    search.MatchesIterator
}

func newTermsEnumDisjunctionMatchesIterator(first search.MatchesIterator, terms bytesref.BytesIterator,
	te index2.TermsEnum, doc int, query search.Query) *termsEnumDisjunctionMatchesIterator {
	return &termsEnumDisjunctionMatchesIterator{
		first: first,
		terms: terms,
		te:    te,
		doc:   doc,
		query: query,
	}
}

func (t *termsEnumDisjunctionMatchesIterator) init() error {
	mis := make([]search.MatchesIterator, 0)
	mis = append(mis, t.first)
	var reuse index2.PostingsEnum

	for {
		term, err := t.terms.Next(nil)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		ok, _ := t.te.SeekExact(nil, term)
		if ok {
			pe, err := t.te.Postings(reuse, index.POSTINGS_ENUM_OFFSETS)
			if err != nil {
				return err
			}
			if v, err := pe.Advance(t.doc); err != nil {
				return err
			} else if v == t.doc {
				iterator, err := NewTermMatchesIterator(t.query, pe)
				if err != nil {
					return err
				}
				mis = append(mis, iterator)
				reuse = nil
			} else {
				reuse = pe
			}
		}
	}
	var err error
	t.it, err = fromSubIterators(mis)
	return err
}

func (t *termsEnumDisjunctionMatchesIterator) Next() (bool, error) {
	if t.it == nil {
		err := t.init()
		if err != nil {
			return false, err
		}
	}
	return t.it.Next()
}

func (t *termsEnumDisjunctionMatchesIterator) StartPosition() int {
	return t.it.StartPosition()
}

func (t *termsEnumDisjunctionMatchesIterator) EndPosition() int {
	return t.it.EndPosition()
}

func (t *termsEnumDisjunctionMatchesIterator) StartOffset() (int, error) {
	return t.it.StartOffset()
}

func (t *termsEnumDisjunctionMatchesIterator) EndOffset() (int, error) {
	return t.it.EndOffset()
}

func (t *termsEnumDisjunctionMatchesIterator) GetSubMatches() (search.MatchesIterator, error) {
	return t.it.GetSubMatches()
}

func (t *termsEnumDisjunctionMatchesIterator) GetQuery() search.Query {
	return t.it.GetQuery()
}
