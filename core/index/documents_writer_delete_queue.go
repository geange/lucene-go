package index

import (
	"go.uber.org/atomic"
	"io"
	"sync"
)

// DocumentsWriterDeleteQueue is a non-blocking linked pending deletes queue.
// In contrast to other queue implementation we only maintain the tail of the queue.
// A delete queue is always used in a context of a set of DWPTs and a global delete pool.
// Each of the DWPT and the global pool need to maintain their 'own' head of the queue
// (as a DeleteSlice instance per DocumentsWriterPerThread).
//
// The difference between the DWPT and the global pool is that the DWPT starts maintaining a head once
// it has added its first document since for its segments private deletes only the deletes after that
// document are relevant. The global pool instead starts maintaining the head once this instance is
// created by taking the sentinel instance as its initial head.
//
// Since each DocumentsWriterDeleteQueue.DeleteSlice maintains its own head and the list is only single linked the garbage collector takes care of pruning the list for us. All nodes in the list that are still relevant should be either directly or indirectly referenced by one of the DWPT's private DocumentsWriterDeleteQueue.DeleteSlice or by the global BufferedUpdates slice.
// Each DWPT as well as the global delete pool maintain their private DeleteSlice instance. In the DWPT case updating a slice is equivalent to atomically finishing the document. The slice update guarantees a "happens before" relationship to all other updates in the same indexing session. When a DWPT updates a document it:
// 1. consumes a document and finishes its processing
// 2. updates its private DocumentsWriterDeleteQueue.DeleteSlice either by calling updateSlice(DocumentsWriterDeleteQueue.DeleteSlice) or add(DocumentsWriterDeleteQueue.Node, DocumentsWriterDeleteQueue.DeleteSlice) (if the document has a delTerm)
// 3. applies all deletes in the slice to its private BufferedUpdates and resets it
// 4. increments its internal document id
//
// The DWPT also doesn't apply its current documents delete term until it has updated its delete slice which ensures the consistency of the update. If the update fails before the DeleteSlice could have been updated the deleteTerm will also not be added to its private deletes neither to the global deletes.
type DocumentsWriterDeleteQueue struct {
	// the current end (latest delete operation) in the delete queue:
	tail   Node
	closed bool

	// Used to record deletes against all prior (already written to disk) segments. Whenever any segment flushes, we bundle up this set of deletes and insert into the buffered updates stream before the newly flushed segment(s).
	globalSlice           *DeleteSlice
	globalBufferedUpdates *BufferedUpdates

	// only acquired to update the global deletes, pkg-private for access by tests:
	globalBufferLock sync.Locker

	generation int64

	// Generates the sequence number that IW returns to callers changing the index, showing the effective serialization of all operations.
	nextSeqNo  *atomic.Int64
	infoStream io.Writer
	maxSeqNo   int64
	startSeqNo int64
	advanced   bool
}

func (d *DocumentsWriterDeleteQueue) Add(deleteNode Node, slice *DeleteSlice) int64 {

	panic("")
}

func (d *DocumentsWriterDeleteQueue) UpdateSlice(slice *DeleteSlice) int64 {
	panic("")
}

type DeleteSlice struct {
	// No need to be volatile, slices are thread captive (only accessed by one thread)!
	sliceHead Node
	sliceTail Node
}

func (d *DeleteSlice) Apply(del *BufferedUpdates, docIDUpto int) {
	if d.sliceHead == d.sliceTail {
		return
	}

	// When we apply a slice we take the head and get its next as our first
	// item to apply and continue until we applied the tail. If the head and
	// tail in this slice are not equal then there will be at least one more
	// non-null node in the slice!
	current := d.sliceHead
	for {
		current = d.sliceHead
		current.Apply(del, docIDUpto)

		if current == d.sliceTail {
			break
		}
	}
	d.Reset()
}

func (d *DeleteSlice) Reset() {
	// Reset to a 0 length slice
	d.sliceHead = d.sliceTail
}

type Node interface {
	Apply(bufferedDeletes *BufferedUpdates, docIDUpto int)
}

type TermNode struct {
	next *TermNode
	item *Term
}

func (t *TermNode) Apply(bufferedDeletes *BufferedUpdates, docIDUpto int) {
	bufferedDeletes.AddTerm(t.item, docIDUpto)
}
