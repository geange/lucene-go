package index

type FlushNotifications interface {
	// DeleteUnusedFiles
	// Called when files were written to disk that are not used anymore. It's the implementation's
	// responsibility to clean these files up
	DeleteUnusedFiles(files map[string]struct{})

	// FlushFailed
	// Called when a segment failed to Flush.
	FlushFailed(info SegmentInfo)

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
