package index

import "io"

type FlushPolicy struct {
	indexWriterConfig *liveIndexWriterConfig
	infoStream        io.Writer

	// Called for each delete term. If this is a delete triggered due to an update the
	// given DocumentsWriterPerThread is non-null.
	// Note: This method is called synchronized on the given DocumentsWriterFlushControl and
	// it is guaranteed that the calling thread holds the lock on the given DocumentsWriterPerThread
	onDelete func(control *DocumentsWriterFlushControl, perThread *DocumentsWriterPerThread)

	// Called for each document addition on the given DocumentsWriterPerThreads DocumentsWriterPerThread.
	// Note: This method is synchronized by the given DocumentsWriterFlushControl and it is guaranteed that the calling thread holds the lock on the given DocumentsWriterPerThread
	onInsert func(control *DocumentsWriterFlushControl, perThread *DocumentsWriterPerThread)
}

// OnUpdate Called for each document update on the given DocumentsWriterPerThread's DocumentsWriterPerThread.
// Note: This method is called synchronized on the given DocumentsWriterFlushControl and it is guaranteed that
// the calling thread holds the lock on the given DocumentsWriterPerThread
func (f *FlushPolicy) OnUpdate(control *DocumentsWriterFlushControl, perThread *DocumentsWriterPerThread) {
	f.onInsert(control, perThread)
	f.onDelete(control, perThread)
}

// Init Called by DocumentsWriter to initialize the FlushPolicy
func (f *FlushPolicy) Init(indexWriterConfig *liveIndexWriterConfig) {
	f.indexWriterConfig = indexWriterConfig
	f.infoStream = indexWriterConfig.infoStream
}

// Returns the current most RAM consuming non-pending DocumentsWriterPerThread with at least one indexed document.
// This method will never return null
func (f *FlushPolicy) findLargestNonPendingWriter(
	control *DocumentsWriterFlushControl, perThread *DocumentsWriterPerThread) *DocumentsWriterPerThread {

	panic("")
}
