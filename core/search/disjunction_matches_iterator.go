package search

import "github.com/geange/lucene-go/core/util/structure"

var _ MatchesIterator = &DisjunctionMatchesIterator{}

type DisjunctionMatchesIterator struct {
	queue   *structure.PriorityQueue[MatchesIterator]
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

func (d *DisjunctionMatchesIterator) GetSubMatches() (MatchesIterator, error) {
	return d.queue.Top().GetSubMatches()
}

func (d *DisjunctionMatchesIterator) GetQuery() Query {
	return d.queue.Top().GetQuery()
}

func fromSubIterators(mis []MatchesIterator) (MatchesIterator, error) {
	if len(mis) == 0 {
		return nil, nil
	}
	if len(mis) == 1 {
		return mis[0], nil
	}
	return newDisjunctionMatchesIterator(mis)
}

func newDisjunctionMatchesIterator(matches []MatchesIterator) (MatchesIterator, error) {
	queue := structure.NewPriorityQueue[MatchesIterator](len(matches), func(a, b MatchesIterator) bool {
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
