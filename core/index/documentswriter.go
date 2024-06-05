package index

import (
	"context"
	"sync/atomic"

	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/store"
)

// DocumentsWriter
// This class accepts multiple added documents and directly writes segment files.
// Each added document is passed to the indexing chain, which in turn processes the document into
// the different codec formats. Some formats write bytes to files immediately, e.g. stored fields
// and term vectors, while others are buffered by the indexing chain and written only on Flush.
// Once we have used our allowed RAM buffer, or the number of added docs is large enough (in the
// case we are flushing by doc count instead of RAM usage), we create a real segment and Flush it
// to the Directory. Threads: Multiple threads are allowed into addDocument at once. There is an
// initial synchronized call to DocumentsWriterFlushControl.ObtainAndLock() which allocates a DWPT
// for this indexing thread. The same thread will not necessarily get the same DWPT over time.
// Then updateDocuments is called on that DWPT without synchronization (most of the "heavy lifting"
// is in this call). Once a DWPT fills up enough RAM or hold enough documents in memory the DWPT
// is checked out for Flush and all changes are written to the directory. Each DWPT corresponds to
// one segment being written. When Flush is called by IndexWriter we check out all DWPTs that are
// associated with the current DocumentsWriterDeleteQueue out of the DocumentsWriterPerThreadPool
// and write them to disk. The Flush process can piggy-back on incoming indexing threads or even
// block them from adding documents if flushing can't keep up with new documents being added.
// Unless the stall control kicks in to block indexing threads flushes are happening concurrently
// to actual index requests. Exceptions: Because this class directly updates in-memory posting lists,
// and flushes stored fields and term vectors directly to files in the directory, there are certain
// limited times when an exception can corrupt this state. For example, a disk full while flushing
// stored fields leaves this file in a corrupt state. Or, an OOM exception while appending to the
// in-memory posting lists can corrupt that posting list. We call such exceptions "aborting exceptions".
// In these cases we must call abort() to discard all docs added since the last Flush. All other
// exceptions ("non-aborting exceptions") can still partially update the index structures.
// These updates are consistent, but, they represent only a part of the document seen up until the
// exception was hit. When this happens, we immediately mark the document as deleted so that the
// document is always atomically ("all or none") added to the index.
type DocumentsWriter struct {
	pendingNumDocs     *atomic.Int64
	flushNotifications FlushNotifications
	closed             bool
	config             *liveIndexWriterConfig
	numDocsInRAM       *atomic.Int64

	// TODO: cut over to BytesHash in BufferedDeletes
	deleteQueue *DocumentsWriterDeleteQueue
	ticketQueue *DocumentsWriterFlushQueue

	// we preserve changes during a full Flush since IW might not checkout before
	// we release all changes. NRT Readers otherwise suddenly return true from
	// isCurrent while there are actually changes currently committed. See also
	// #anyChanges() & #flushAllThreads
	pendingChangesInCurrentFullFlush bool

	perThreadPool *DocumentsWriterPerThreadPool
	flushControl  *DocumentsWriterFlushControl
}

func NewDocumentsWriter(flushNotifications FlushNotifications, indexCreatedVersionMajor int, pendingNumDocs *atomic.Int64, enableTestPoints bool,
	segmentName string, config *liveIndexWriterConfig, directoryOrig, directory store.Directory,
	globalFieldNumberMap *FieldNumbers) *DocumentsWriter {

	infos := NewFieldInfosBuilder(globalFieldNumberMap)

	deleteQueue := NewDocumentsWriterDeleteQueue()

	docWriter := &DocumentsWriter{
		pendingNumDocs:                   pendingNumDocs,
		flushNotifications:               flushNotifications,
		closed:                           false,
		config:                           config,
		numDocsInRAM:                     new(atomic.Int64),
		deleteQueue:                      deleteQueue,
		ticketQueue:                      nil,
		pendingChangesInCurrentFullFlush: false,
		perThreadPool:                    nil,
		flushControl: &DocumentsWriterFlushControl{
			perThread: NewDocumentsWriterPerThread(indexCreatedVersionMajor,
				segmentName, directoryOrig,
				directory, config, deleteQueue, infos,
				pendingNumDocs, enableTestPoints)},
	}
	return docWriter
}

func (d *DocumentsWriter) purgeFlushTickets(forced bool, consumer func(*FlushTicket) error) error {
	if forced {
		return d.ticketQueue.forcePurge(consumer)
	}
	return nil
}

func (d *DocumentsWriter) preUpdate() (bool, error) {
	panic("")
}

// TODO: fix it
func (d *DocumentsWriter) updateDocuments(ctx context.Context, docs []*document.Document, delNode *Node) (int64, error) {
	dwpt := d.flushControl.ObtainAndLock()
	dwptNumDocs := dwpt.GetNumDocsInRAM()
	seqNo, err := dwpt.updateDocuments(ctx, docs, delNode)
	if err != nil {
		return 0, err
	}
	d.numDocsInRAM.Add(int64(dwpt.GetNumDocsInRAM() - dwptNumDocs))

	return seqNo, nil
}

func (d *DocumentsWriter) Flush(ctx context.Context) error {
	dwpt := d.flushControl.ObtainAndLock()
	_, err := d.doFlush(ctx, dwpt)
	return err
}

// FIXME: 处理flush
func (d *DocumentsWriter) doFlush(ctx context.Context, flushingDWPT *DocumentsWriterPerThread) (bool, error) {
	hasEvents := false

	for flushingDWPT != nil {
		hasEvents = true

		dwptSuccess := true

		//ticket, err := d.ticketQueue.AddFlushTicket(flushingDWPT)
		//if err != nil {
		//	return err
		//}
		//flushingDocsInRam := flushingDWPT.GetNumDocsInRAM()

		// flush concurrently without locking
		//newSegment, err := flushingDWPT.flush(ctx, d.flushNotifications)
		if _, err := flushingDWPT.flush(ctx, d.flushNotifications); err != nil {
			dwptSuccess = false
		}
		//d.ticketQueue.AddSegment(ticket, newSegment)
		//
		//d.subtractFlushedNumDocs(int64(flushingDocsInRam))
		if (len(flushingDWPT.PendingFilesToDelete()) == 0) == false {
			files := flushingDWPT.PendingFilesToDelete()
			d.flushNotifications.DeleteUnusedFiles(files)
			hasEvents = true
		}
		if dwptSuccess == false {
			d.flushNotifications.FlushFailed(flushingDWPT.GetSegmentInfo())
			hasEvents = true
		}

		if err := d.flushControl.DoAfterFlush(flushingDWPT); err != nil {
			return false, err
		}
		flushingDWPT = d.flushControl.NextPendingFlush()
	}

	if hasEvents {
		if err := d.flushNotifications.AfterSegmentsFlushed(); err != nil {
			return false, err
		}
	}

	return hasEvents, nil
}

func (d *DocumentsWriter) anyChanges() bool {

	// changes are either in a DWPT or in the deleteQueue.
	// yet if we currently flush deletes and / or dwpt there
	// could be a window where all changes are in the ticket queue
	// before they are published to the IW. ie we need to check if the
	// ticket queue has any tickets.
	anyChanges := d.numDocsInRAM.Load() != 0 ||
		d.anyDeletions() ||
		d.ticketQueue.hasTickets() ||
		d.pendingChangesInCurrentFullFlush
	return anyChanges
}

func (d *DocumentsWriter) anyDeletions() bool {
	return d.deleteQueue.anyChanges()
}

// FlushAllThreads is synced by IW fullFlushLock. Flushing all threads is a
// two stage operation; the caller must ensure (in try/finally) that finishFlush
// is called after this method, to release the flush lock in DWFlushControl
func (d *DocumentsWriter) flushAllThreads() int64 {
	var flushingDeleteQueue *DocumentsWriterDeleteQueue

	var seqNo int64

	d.pendingChangesInCurrentFullFlush = d.anyChanges()
	flushingDeleteQueue = d.deleteQueue

	// Cutover to a new delete queue.  This must be synced on the flush control
	// otherwise a new DWPT could sneak into the loop with an already flushing
	// delete queue
	seqNo = d.flushControl.MarkForFullFlush() // swaps this.deleteQueue synced on FlushControl

	anythingFlushed := false

	ctx := context.Background()

	var flushingDWPT *DocumentsWriterPerThread
	for {
		flushingDWPT = d.flushControl.NextPendingFlush()
		if flushingDWPT == nil {
			break
		}

		hasEvent, err := d.doFlush(ctx, flushingDWPT)
		if err != nil {
			return 0
		}

		anythingFlushed = anythingFlushed || hasEvent
	}

	// If a concurrent flush is still in flight wait for it
	//d.flushControl.WaitForFlush();
	if anythingFlushed == false && flushingDeleteQueue.anyChanges() { // apply deletes if we did not flush any document
		err := d.ticketQueue.AddDeletes(flushingDeleteQueue)
		if err != nil {
			return 0
		}
	}

	flushingDeleteQueue.Close() // all DWPT have been processed and this queue has been fully flushed to the ticket-queue

	if anythingFlushed {
		return -seqNo
	} else {
		return seqNo
	}
}

func (d *DocumentsWriter) finishFullFlush(success bool) error {
	if success {
		// Release the flush lock
		if err := d.flushControl.finishFullFlush(); err != nil {
			return err
		}
	} else {
		if err := d.flushControl.abortFullFlushes(); err != nil {
			return err
		}
	}

	d.pendingChangesInCurrentFullFlush = false
	return d.applyAllDeletes() // make sure we do execute this since we block applying deletes during full flush
}

func (d *DocumentsWriter) subtractFlushedNumDocs(numFlushed int64) {
	oldValue := d.numDocsInRAM.Load()
	for d.numDocsInRAM.CompareAndSwap(oldValue, oldValue-numFlushed) == false {
		oldValue = d.numDocsInRAM.Load()
	}
}

func (d *DocumentsWriter) applyAllDeletes() error {
	panic("")
}

type FlushNotifications interface {
	// DeleteUnusedFiles
	// Called when files were written to disk that are not used anymore. It's the implementation's
	// responsibility to clean these files up
	DeleteUnusedFiles(files map[string]struct{})

	// FlushFailed
	// Called when a segment failed to Flush.
	FlushFailed(info *SegmentInfo)

	// AfterSegmentsFlushed
	// Called after one or more segments were flushed to disk.
	AfterSegmentsFlushed() error

	// Should be called if a Flush or an indexing operation caused a tragic / unrecoverable event.
	//onTragicEvent(Throwable event, String message)

	// OnDeletesApplied
	// Called once deletes have been applied either after a Flush or on a deletes call
	OnDeletesApplied()

	// OnTicketBacklog
	// Called once the DocumentsWriter ticket queue has a backlog. This means there
	// is an inner thread that tries to publish flushed segments but can't keep up with the other
	// threads flushing new segments. This likely requires other thread to forcefully purge the buffer
	// to help publishing. This can't be done in-place since we might hold index writer locks when
	// this is called. The caller must ensure that the purge happens without an index writer lock being held.
	// See Also: purgeFlushTickets(boolean, IOUtils.IOConsumer)
	OnTicketBacklog()
}
