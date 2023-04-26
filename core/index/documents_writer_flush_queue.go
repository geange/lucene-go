package index

// DocumentsWriterFlushQueue
// lucene.internal
type DocumentsWriterFlushQueue struct {
	queue []*FlushTicket
}

func (q *DocumentsWriterFlushQueue) hasTickets() bool {
	// TODO: fix it
	// return ticketCount.get() != 0
	return false
}

type FlushTicket struct {
	frozenUpdates FrozenBufferedUpdates
}
