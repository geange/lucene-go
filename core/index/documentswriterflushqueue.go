package index

import (
	"sync"
	"sync/atomic"

	linked "github.com/geange/gods-generic/lists/singlylinkedlist"
)

// DocumentsWriterFlushQueue
// lucene.internal
type DocumentsWriterFlushQueue struct {
	purgeLock   sync.Mutex
	queue       linked.List[*FlushTicket]
	ticketCount *atomic.Int32
}

func (q *DocumentsWriterFlushQueue) hasTickets() bool {
	// TODO: fix it
	// return ticketCount.get() != 0
	return false
}

func (q *DocumentsWriterFlushQueue) AddFlushTicket(dwpt *DocumentsWriterPerThread) (*FlushTicket, error) {
	// Each flush is assigned a ticket in the order they acquire the ticketQueue lock
	q.ticketCount.Add(1)

	success := false

	defer func(flag *bool) {
		if !success {
			q.ticketCount.Add(-1)
		}
	}(&success)

	// prepare flush freezes the global deletes - do in synced block!
	frozenBufferedUpdates, err := dwpt.prepareFlush()
	if err != nil {
		return nil, err
	}

	ticket := NewFlushTicket(frozenBufferedUpdates, true)
	q.queue.Add(ticket)

	return ticket, nil
}

func (q *DocumentsWriterFlushQueue) AddSegment(ticket *FlushTicket, segment *FlushedSegment) {
	// the actual flush is done asynchronously and once done the FlushedSegment
	// is passed to the flush ticket
	ticket.setSegment(segment)
}

func (q *DocumentsWriterFlushQueue) forcePurge(consumer func(*FlushTicket) error) error {
	q.purgeLock.Lock()
	defer q.purgeLock.Unlock()
	return q.innerPurge(consumer)
}

func (q *DocumentsWriterFlushQueue) tryPurge(consumer func(*FlushTicket) error) error {
	if !q.purgeLock.TryLock() {
		return nil
	}

	defer q.purgeLock.Unlock()

	return q.innerPurge(consumer)
}

func (q *DocumentsWriterFlushQueue) innerPurge(consumer func(ticket *FlushTicket) error) error {
	for {
		head, ok := q.queue.Get(0)
		if !ok {
			break
		}
		canPublish := head != nil && head.canPublish() // do this synced

		if canPublish {
			err := consumer(head)
			if err != nil {
				return err
			}
		}
		q.queue.Remove(0)
	}
	return nil
}

func (q *DocumentsWriterFlushQueue) decTickets() {
	q.ticketCount.Add(-1)
}

func (q *DocumentsWriterFlushQueue) AddDeletes(queue *DocumentsWriterDeleteQueue) (bool, error) {
	panic("")
}

type FlushTicket struct {
	sync.Mutex

	frozenUpdates *FrozenBufferedUpdates
	hasSegment    bool
	segment       *FlushedSegment
	failed        bool
	published     bool
}

func NewFlushTicket(frozenUpdates *FrozenBufferedUpdates, hasSegment bool) *FlushTicket {
	return &FlushTicket{
		frozenUpdates: frozenUpdates,
		hasSegment:    hasSegment,
	}
}

func (t *FlushTicket) canPublish() bool {
	return t.hasSegment == false || t.segment != nil || t.failed
}

func (t *FlushTicket) markPublished() {
	t.Lock()
	defer t.Unlock()
	t.published = true
}

func (t *FlushTicket) setSegment(segment *FlushedSegment) {
	t.segment = segment
}

func (t *FlushTicket) setFailed() {
	t.failed = true
}

// Returns the flushed segment or null if this flush ticket doesn't have a segment.
// This can be the case if this ticket represents a flushed global frozen updates package.
func (t *FlushTicket) getFlushedSegment() *FlushedSegment {
	return t.segment
}

// Returns a frozen global deletes package.
func (t *FlushTicket) getFrozenUpdates() *FrozenBufferedUpdates {
	return t.frozenUpdates
}
