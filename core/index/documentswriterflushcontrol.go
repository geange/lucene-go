package index

import (
	"slices"
	"sync/atomic"
)

// DocumentsWriterFlushControl
// This class controls DocumentsWriterPerThread flushing during indexing.
// It tracks the memory consumption per DocumentsWriterPerThread and uses a configured FlushPolicy
// to decide if a DocumentsWriterPerThread must Flush.
//
// In addition to the FlushPolicy the Flush control might set certain DocumentsWriterPerThread as
// Flush pending iff a DocumentsWriterPerThread exceeds the IndexWriterConfig.getRAMPerThreadHardLimitMB()
// to prevent address space exhaustion.
type DocumentsWriterFlushControl struct {
	hardMaxBytesPerDWPT int64
	activeBytes         int64
	flushBytes          int64
	numPending          int
	numDocsSinceStalled int
	flushDeletes        *atomic.Bool
	fullFlush           bool
	fullFlushMarkDone   bool

	// The flushQueue is used to concurrently distribute DWPTs that are ready to be flushed ie. when a full Flush is in
	// progress. This might be triggered by a commit or NRT refresh. The trigger will only walk all eligible DWPTs and
	// mark them as flushable putting them in the flushQueue ready for other threads (ie. indexing threads) to help flushing
	flushQueue []*DocumentsWriterPerThread

	// only for safety reasons if a DWPT is close to the RAM limit
	blockedFlushes []*DocumentsWriterPerThread

	// flushingWriters holds all currently flushing writers. There might be writers in this list that
	// are also in the flushQueue which means that writers in the flushingWriters list are not necessarily
	// already actively flushing. They are only in the state of flushing and might be picked up in the future by
	// polling the flushQueue
	flushingWriters []*DocumentsWriterPerThread

	maxConfiguredRamBuffer float64
	peakActiveBytes        int64
	peakFlushBytes         int64
	peakNetBytes           int64
	peakDelta              int64
	perThread              *DocumentsWriterPerThread

	flushPolicy     FlushPolicy
	closed          bool
	documentsWriter *DocumentsWriter
	config          *LiveIndexWriterConfig
}

//
//func NewDocumentsWriterFlushControl(documentsWriter *DocumentsWriter,
//	config *liveIndexWriterConfig) *DocumentsWriterFlushControl {
//
//	return &DocumentsWriterFlushControl{
//		perThread: NewDocumentsWriterPerThread(),
//	}
//
//}

func (d *DocumentsWriterFlushControl) ObtainAndLock() *DocumentsWriterPerThread {
	return d.perThread
}

func (d *DocumentsWriterFlushControl) DoAfterFlush(dwpt *DocumentsWriterPerThread) error {
	return nil
}

func (d *DocumentsWriterFlushControl) NextPendingFlush() *DocumentsWriterPerThread {
	return nil
}

func (d *DocumentsWriterFlushControl) MarkForFullFlush() int64 {
	return 0
}

// Prunes the blockedQueue by removing all DWPTs that are associated with the given flush queue.
// TODO: zero copy
func (d *DocumentsWriterFlushControl) pruneBlockedQueue(flushingQueue *DocumentsWriterDeleteQueue) error {
	newBlockedFlushes := make([]*DocumentsWriterPerThread, 0)
	for _, blockedFlush := range d.blockedFlushes {
		if blockedFlush.deleteQueue == flushingQueue {
			d.addFlushingDWPT(blockedFlush)
		} else {
			newBlockedFlushes = append(newBlockedFlushes, blockedFlush)
		}
	}
	d.blockedFlushes = newBlockedFlushes
	return nil
}

func (d *DocumentsWriterFlushControl) finishFullFlush() error {
	if len(d.blockedFlushes) > 0 {
		err := d.pruneBlockedQueue(d.documentsWriter.deleteQueue)
		if err != nil {
			return err
		}
		d.fullFlushMarkDone = false
		d.fullFlush = false
	}
	return nil
}

func (d *DocumentsWriterFlushControl) addFlushingDWPT(perThread *DocumentsWriterPerThread) {
	d.flushingWriters = append(d.flushingWriters, perThread)
}

func (d *DocumentsWriterFlushControl) abortFullFlushes() error {
	if err := d.abortPendingFlushes(); err != nil {
		return err
	}
	d.fullFlushMarkDone = false
	d.fullFlush = false
	return nil
}

func (d *DocumentsWriterFlushControl) abortPendingFlushes() error {
	for _, dwpt := range d.flushQueue {
		err := dwpt.abort()
		if err != nil {
			return err
		}
		err = d.doAfterFlush(dwpt)
		if err != nil {
			return err
		}
	}

	for _, blockedFlush := range d.blockedFlushes {
		d.addFlushingDWPT(blockedFlush) // add the blockedFlushes for correct accounting in doAfterFlush
		d.documentsWriter.subtractFlushedNumDocs(int64(blockedFlush.GetNumDocsInRAM()))
		err := blockedFlush.abort()
		if err != nil {
			return err
		}
		err = d.doAfterFlush(blockedFlush)
		if err != nil {
			return err
		}
	}

	clear(d.flushQueue)
	clear(d.blockedFlushes)
	return nil
}

func (d *DocumentsWriterFlushControl) doAfterFlush(dwpt *DocumentsWriterPerThread) error {
	slices.DeleteFunc(d.flushingWriters, func(thread *DocumentsWriterPerThread) bool {
		return thread == dwpt
	})
	return nil
}

func (d *DocumentsWriterFlushControl) isFullFlush() bool {
	return d.fullFlush
}

func (d *DocumentsWriterFlushControl) getAndResetApplyAllDeletes() bool {
	return d.flushDeletes.Swap(false)
}
