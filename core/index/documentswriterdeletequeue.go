package index

import (
	"math"
	"sync"
	"sync/atomic"

	"github.com/geange/lucene-go/core/interface/index"
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
// Since each DocumentsWriterDeleteQueue.DeleteSlice maintains its own head and the list is only single
// linked the garbage collector takes care of pruning the list for us. All nodes in the list that are
// still relevant should be either directly or indirectly referenced by one of the DWPT's private
// DocumentsWriterDeleteQueue.DeleteSlice or by the global BufferedUpdates slice.
// Each DWPT as well as the global delete pool maintain their private DeleteSlice instance.
// In the DWPT case updating a slice is equivalent to atomically finishing the document.
// The slice update guarantees a "happens before" relationship to all other updates in the
// same indexing session. When a DWPT updates a document it:
//
// 1. consumes a document and finishes its processing
// 2. updates its private DocumentsWriterDeleteQueue.DeleteSlice either by calling updateSlice(DocumentsWriterDeleteQueue.DeleteSlice) or add(DocumentsWriterDeleteQueue.Node, DocumentsWriterDeleteQueue.DeleteSlice) (if the document has a delTerm)
// 3. applies all deletes in the slice to its private BufferedUpdates and resets it
// 4. increments its internal document id
//
// The DWPT also doesn't apply its current documents delete term until it has updated its delete slice which ensures the consistency of the update. If the update fails before the DeleteSlice could have been updated the deleteTerm will also not be added to its private deletes neither to the global deletes.
type DocumentsWriterDeleteQueue struct {
	// the current end (latest delete operation) in the delete queue:
	tail   *Node
	closed bool

	// Used to record deletes against all prior (already written to disk) segments.
	// Whenever any segment flushes, we bundle up this set of deletes and insert into
	// the buffered updates stream before the newly flushed segment(s).
	globalSlice           *DeleteSlice
	globalBufferedUpdates *index.BufferedUpdates

	// only acquired to update the global deletes, pkg-private for access by tests:
	globalBufferLock sync.Locker

	generation int64

	// Generates the sequence number that IW returns to callers changing the index,
	// showing the effective serialization of all operations.
	nextSeqNo        *atomic.Int64
	maxSeqNo         int64
	startSeqNo       int64
	previousMaxSeqId func() int64
	advanced         bool
}

func NewDocumentsWriterDeleteQueue() *DocumentsWriterDeleteQueue {
	return newDocumentsWriterDeleteQueue(0, -1, func() int64 {
		return 0
	})
}

func newDocumentsWriterDeleteQueue(generation, startSeqNo int64,
	previousMaxSeqId func() int64) *DocumentsWriterDeleteQueue {

	tail := NewNode(nil, nil)

	nextSeqNo := new(atomic.Int64)
	nextSeqNo.Add(startSeqNo)

	return &DocumentsWriterDeleteQueue{
		tail:                  tail,
		closed:                false,
		globalSlice:           NewDeleteSlice(tail),
		globalBufferedUpdates: index.NewBufferedUpdates(index.WithSegmentName("global")),
		globalBufferLock:      &sync.Mutex{},
		generation:            generation,
		nextSeqNo:             nextSeqNo,
		maxSeqNo:              math.MaxInt64,
		startSeqNo:            startSeqNo,
		previousMaxSeqId:      previousMaxSeqId,
		advanced:              false,
	}
}

func (d *DocumentsWriterDeleteQueue) Add(deleteNode *Node, slice *DeleteSlice) int64 {
	seqNo := d.add(deleteNode)

	// this is an update request where the term is the updated documents
	// delTerm. in that case we need to guarantee that this insert is atomic
	// with regards to the given delete slice. This means if two threads try to
	// update the same document with in turn the same delTerm one of them must
	// win. By taking the node we have created for our del term as the new tail
	// it is guaranteed that if another thread adds the same right after us we
	// will apply this delete next time we update our slice and one of the two
	// competing updates wins!
	slice.sliceTail = deleteNode
	d.tryApplyGlobalSlice()
	return seqNo
}

func (d *DocumentsWriterDeleteQueue) add(newNode *Node) int64 {
	d.tail.next = newNode
	d.tail = newNode
	return d.getNextSequenceNumber()
}

// UpdateSlice Negative result means there were new deletes since we last applied
func (d *DocumentsWriterDeleteQueue) UpdateSlice(slice *DeleteSlice) int64 {
	seqNo := d.getNextSequenceNumber()
	if slice.sliceTail != d.tail {
		slice.sliceTail = d.tail
		seqNo = -seqNo
	}
	return seqNo
}

func (d *DocumentsWriterDeleteQueue) getNextSequenceNumber() int64 {
	return d.nextSeqNo.Add(1)
}

func (d *DocumentsWriterDeleteQueue) tryApplyGlobalSlice() {
	d.globalBufferLock.Lock()
	defer d.globalBufferLock.Unlock()

	if d.updateSliceNoSeqNo(d.globalSlice) {
		d.globalSlice.Apply(d.globalBufferedUpdates, math.MaxInt32)
	}
}

// Just like updateSlice, but does not assign a sequence number
func (d *DocumentsWriterDeleteQueue) updateSliceNoSeqNo(slice *DeleteSlice) bool {
	if slice.sliceTail != d.tail {
		// new deletes arrived since we last checked
		slice.sliceTail = d.tail
		return true
	}
	return false
}

func (d *DocumentsWriterDeleteQueue) newSlice() *DeleteSlice {
	return NewDeleteSlice(d.tail)
}

func (d *DocumentsWriterDeleteQueue) anyChanges() bool {
	d.globalBufferLock.Lock()
	defer d.globalBufferLock.Unlock()

	// TODO: fix it
	return false
}

func (d *DocumentsWriterDeleteQueue) freezeGlobalBuffer(callerSlice *DeleteSlice) *FrozenBufferedUpdates {
	d.globalBufferLock.Lock()
	defer d.globalBufferLock.Unlock()

	currentTail := d.tail
	if callerSlice != nil {
		// Update the callers slices so we are on the same page
		callerSlice.sliceTail = currentTail
	}
	return d.freezeGlobalBufferInternal(currentTail)
}

func (d *DocumentsWriterDeleteQueue) freezeGlobalBufferInternal(currentTail *Node) *FrozenBufferedUpdates {
	if d.globalSlice.sliceTail != currentTail {
		d.globalSlice.sliceTail = currentTail
		d.globalSlice.Apply(d.globalBufferedUpdates, math.MaxInt32)
	}

	if d.globalBufferedUpdates.Any() {
		packet := NewFrozenBufferedUpdates(d.globalBufferedUpdates, nil)
		d.globalBufferedUpdates.Clear()
		return packet
	}
	return nil
}

func (d *DocumentsWriterDeleteQueue) Close() {

}

func (d *DocumentsWriterDeleteQueue) isOpen() bool {
	return d.closed == false
}

type DeleteSlice struct {
	// No need to be volatile, slices are thread captive (only accessed by one thread)!
	sliceHead *Node
	sliceTail *Node
}

func NewDeleteSlice(currentTail *Node) *DeleteSlice {
	return &DeleteSlice{
		sliceHead: currentTail,
		sliceTail: currentTail,
	}
}

func (d *DeleteSlice) Apply(del *index.BufferedUpdates, docIDUpto int) error {

	if d.sliceHead == d.sliceTail {
		return nil
	}

	// When we apply a slice we take the head and get its next as our first
	// item to apply and continue until we applied the tail. If the head and
	// tail in this slice are not equal then there will be at least one more
	// non-null node in the slice!
	current := d.sliceHead
	for {
		current = current.next
		if err := current.Apply(del, docIDUpto); err != nil {
			return err
		}

		if current == d.sliceTail {
			break
		}
	}
	d.Reset()
	return nil
}

func (d *DeleteSlice) Reset() {
	// Reset to a 0 length slice
	d.sliceHead = d.sliceTail
}

func deleteQueueNewNode(term index.Term) *Node {
	node := NewTermNode(term)
	return NewNode(term, node)
}

func deleteQueueNewNodeDocValuesUpdates(updates []index.DocValuesUpdate) *Node {
	node := NewDocValuesUpdatesNode(updates)
	return NewNode(updates, node)
}
