package index

import (
	"context"
	"errors"
	"fmt"
	"github.com/geange/lucene-go/core/analysis"
	"github.com/geange/lucene-go/core/analysis/standard"
	"github.com/geange/lucene-go/core/interface/index"
	"maps"
	"math"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util"
	"github.com/geange/lucene-go/core/util/version"
)

var (
	// Use package-private instance var to enforce the limit so testing
	// can use less electricity:
	actualMaxDocs = MAX_POSITION
)

const (
	// MAX_DOCS
	// Hard limit on maximum number of documents that may be added to the index.
	// If you try to add more than this you'll hit IllegalArgumentException
	MAX_DOCS = math.MaxInt32 - 128

	// MAX_POSITION
	// Maximum item of the token position in an indexed field.
	MAX_POSITION = math.MaxInt32 - 128

	UNBOUNDED_MAX_MERGE_SEGMENTS = -1

	// WRITE_LOCK_NAME
	// Name of the write lock in the index.
	WRITE_LOCK_NAME = "write.lock"

	// SOURCE
	// key for the source of a segment in the diagnostics.
	SOURCE = "source"

	// SOURCE_MERGE
	// Source of a segment which results from a merge of other segments.
	SOURCE_MERGE = "merge"

	// SOURCE_FLUSH
	// Source of a segment which results from a Flush.
	SOURCE_FLUSH = "Flush"

	// SOURCE_ADDINDEXES_READERS
	// Source of a segment which results from a call to addIndexes(CodecReader...).
	SOURCE_ADDINDEXES_READERS = "addIndexes(CodecReader...)"

	BYTE_BLOCK_SHIFT = 15
	BYTE_BLOCK_SIZE  = 1 << BYTE_BLOCK_SHIFT

	// MAX_TERM_LENGTH
	// Absolute hard maximum length for a term, in bytes once encoded as UTF8.
	// If a term arrives from the analyzer longer than this length, an IllegalArgumentException
	// is thrown and a message is printed to infoStream, if set (see IndexWriterConfig.setInfoStream(InfoStream)).
	MAX_TERM_LENGTH = BYTE_BLOCK_SIZE - 2

	MAX_STORED_STRING_LENGTH = math.MaxInt
)

type IndexWriter struct {
	enableTestPoints         bool
	directoryOrig            store.Directory           // original user directory
	directory                store.Directory           // wrapped with additional checks
	changeCount              *atomic.Int64             // increments every time a change is completed
	lastCommitChangeCount    *atomic.Int64             // last changeCount that was committed
	rollbackSegments         []index.SegmentCommitInfo // list of segmentInfo we will fallback to if the commit fails
	pendingCommit            *SegmentInfos             // set when a commit is pending (after prepareCommit() & before commit())
	pendingSeqNo             int64
	pendingCommitChangeCount int64
	filesToCommit            map[string]struct{}
	segmentInfos             *SegmentInfos
	globalFieldNumberMap     *FieldNumbers
	docWriter                *DocumentsWriter
	eventQueue               *EventQueue
	mergeSource              MergeSource
	writeDocValuesLock       sync.RWMutex
	deleter                  *IndexFileDeleter
	//segmentsToMerge          *hashmap.Map
	mergeMaxNumSegments int
	writeLock           store.Lock
	closed              bool
	closing             bool
	atomMaybeMerge      *atomic.Bool
	commitUserData      map[string]string
	//mergingSegments          *hashset.Set
	mergeScheduler MergeScheduler
	//runningAddIndexesMerges  *hashset.Set
	pendingMerges []*OneMerge
	//runningMerges            *hashset.Set
	mergeExceptions       []*OneMerge
	mergeGen              int64
	merges                *Merges
	didMessageState       bool
	flushCount            *atomic.Int64
	flushDeletesCount     *atomic.Int64
	readerPool            *ReaderPool
	mergeFinishedGen      *atomic.Int64
	bufferedUpdatesStream *BufferedUpdatesStream
	config                *IndexWriterConfig
	startCommitTime       int64
	pendingNumDocs        *atomic.Int64
	softDeletesEnabled    bool
	flushNotifications    index.FlushNotifications
	boolMaybeMerge        *atomic.Bool
}

func NewIndexWriter(ctx context.Context, dir store.Directory, conf *IndexWriterConfig) (*IndexWriter, error) {
	writer := &IndexWriter{
		changeCount:           new(atomic.Int64),
		lastCommitChangeCount: new(atomic.Int64),
		pendingNumDocs:        new(atomic.Int64),
		flushCount:            new(atomic.Int64),
		atomMaybeMerge:        new(atomic.Bool),
	}
	conf.setIndexWriter(writer)
	writer.config = conf
	writer.softDeletesEnabled = conf.getSoftDeletesField() != ""

	writer.directoryOrig = dir
	writer.directory = dir
	writer.mergeScheduler = writer.config.GetMergeScheduler()
	writer.mergeScheduler.Initialize(writer.directoryOrig)

	mode := conf.GetOpenMode()
	var err error
	var indexExists, create bool
	switch mode {
	case CREATE:
		indexExists, err = IsIndexExists(writer.directory)
		create = true
	case APPEND:
		indexExists = true
		create = false
	default:
		// CREATE_OR_APPEND - create only if an index does not exist
		indexExists, err = IsIndexExists(writer.directory)
		create = !indexExists
	}
	if err != nil {
		return nil, err
	}

	// If index is too old, reading the segments will throw IndexFormatTooOldException.
	files, err := writer.directory.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	// Set up our initial SegmentInfos:
	commit := writer.config.GetIndexCommit()

	// Set up our initial SegmentInfos:
	var reader *StandardDirectoryReader
	if commit != nil {
		reader = commit.GetReader()
	}

	if create {
		if writer.config.GetIndexCommit() != nil {
			// We cannot both open from a commit point and create:
			if mode == CREATE {
				return nil, errors.New("cannot use IndexWriterConfig.setIndexCommit() with OpenMode.CREATE")
			} else {
				return nil, errors.New("cannot use IndexWriterConfig.setIndexCommit() when index has no commit")
			}
		}

		// Try to read first.  This is to allow create
		// against an index that's currently open for
		// searching.  In this case we write the next
		// segments_N file with no segments:
		segmentInfos := NewSegmentInfos(writer.config.GetIndexCreatedVersionMajor())
		if indexExists {
			previous, err := ReadLatestCommit(ctx, writer.directory)
			if err != nil {
				return nil, err
			}
			segmentInfos.UpdateGenerationVersionAndCounter(previous)
		}
		writer.segmentInfos = segmentInfos
		writer.rollbackSegments = writer.segmentInfos.CreateBackupSegmentInfos()

		// Record that we have a change (zero out all segments) pending:
		writer.Changed()
	} else if reader != nil {
		// Init from an existing already opened NRT or non-NRT reader:
		if reader.Directory() != commit.GetDirectory() {
			return nil, errors.New("IndexCommit's reader must have the same directory as the IndexCommit")
		}

		if reader.Directory() != writer.directoryOrig {
			return nil, errors.New("IndexCommit's reader must have the same directory passed to IndexWriter")
		}

		if reader.segmentInfos.GetLastGeneration() == 0 {
			// TODO: maybe we could allow this?  It's tricky...
			return nil, errors.New("index must already have an initial commit to open from reader")
		}

		// Must clone because we don't want the incoming NRT reader to "see" any changes this writer now makes:
		reader.segmentInfos.Clone()

		var lastCommit *SegmentInfos
		lastCommit, err = ReadCommit(ctx, writer.directoryOrig, writer.segmentInfos.GetSegmentsFileName())
		if err != nil {
			return nil, err
		}

		if reader.writer != nil {
			// assert reader.writer.closed;
			if !reader.writer.closed {
				return nil, errors.New("reader.writer not closed")
			}

			// In case the old writer wrote further segments (which we are now dropping),
			// update SIS metadata so we remain write-once:
			writer.segmentInfos.UpdateGenerationVersionAndCounter(reader.writer.segmentInfos)
			lastCommit.UpdateGenerationVersionAndCounter(reader.writer.segmentInfos)
		}

		writer.rollbackSegments = lastCommit.CreateBackupSegmentInfos()
	} else {
		// Init from either the latest commit point, or an explicit prior commit point:

		lastSegmentsFile, err := GetLastCommitSegmentsFileName(files)
		if err != nil {
			return nil, err
		}
		if lastSegmentsFile == "" {
			return nil, errors.New("no segments* file found")
		}

		// Do not use SegmentInfos.read(Directory) since the spooky
		// retrying it does is not necessary here (we hold the write lock):
		segmentInfos, err := ReadCommit(ctx, writer.directoryOrig, lastSegmentsFile)
		if err != nil {
			return nil, err
		}
		writer.segmentInfos = segmentInfos

		if commit != nil {
			// Swap out all segments, but, keep metadata in
			// SegmentInfos, like version & generation, to
			// preserve write-once.  This is important if
			// readers are open against the future commit
			// points.

			if commit.GetDirectory() != writer.directoryOrig {
				return nil, errors.New("IndexCommit's directory doesn't match my directory")
			}

			oldInfos, err := ReadCommit(ctx, writer.directoryOrig, commit.GetSegmentsFileName())
			if err != nil {
				return nil, err
			}
			if err := writer.segmentInfos.Replace(oldInfos); err != nil {
				return nil, err
			}
			writer.Changed()
		}
		writer.rollbackSegments = writer.segmentInfos.CreateBackupSegmentInfos()
	}

	writer.commitUserData = writer.segmentInfos.GetUserData()
	writer.pendingNumDocs.Swap(writer.segmentInfos.TotalMaxDoc())

	// start with previous field numbers, but new FieldInfos
	// NOTE: this is correct even for an NRT reader because we'll pull FieldInfos even for the un-committed segments:
	globalFieldNumberMap, err := writer.getFieldNumberMap()
	if err != nil {
		return nil, err
	}
	writer.globalFieldNumberMap = globalFieldNumberMap

	if err := writer.validateIndexSort(); err != nil {
		return nil, err
	}

	// TODO: liveIndexWriterConfig 转换成interface
	//writer.config.GetFlushPolicy().Init(writer.config.)

	//writer.segmentInfos = NewSegmentInfos(conf.GetIndexCreatedVersionMajor())
	//
	//writer.globalFieldNumberMap = writer.getFieldNumberMap()
	writer.bufferedUpdatesStream = NewBufferedUpdatesStream()

	writer.eventQueue = NewEventQueue(writer)
	writer.flushNotifications = writer.newFlushNotifications()

	writer.docWriter = NewDocumentsWriter(writer.flushNotifications, writer.segmentInfos.getIndexCreatedVersionMajor(), writer.pendingNumDocs,
		writer.enableTestPoints, writer.newSegmentName(),
		writer.config.liveIndexWriterConfig, writer.directoryOrig, writer.directory, writer.globalFieldNumberMap)

	writer.bufferedUpdatesStream.GetCompletedDelGen()
	writer.readerPool, err = NewReaderPool(writer.directory, writer.directoryOrig, writer.segmentInfos,
		writer.globalFieldNumberMap, writer.bufferedUpdatesStream.GetCompletedDelGen,
		conf.getSoftDeletesField(), reader)
	if err != nil {
		return nil, err
	}

	deleter, err := NewIndexFileDeleter(ctx, files, writer.directoryOrig, writer.directory,
		writer.config.GetIndexDeletionPolicy(),
		writer.segmentInfos, writer,
		indexExists, reader != nil)
	if err != nil {
		return nil, err
	}
	writer.deleter = deleter

	writer.flushDeletesCount = new(atomic.Int64)
	writer.boolMaybeMerge = new(atomic.Bool)

	return writer, nil
}

// Confirms that the incoming index sort (if any) matches the existing index sort (if any).
func (w *IndexWriter) validateIndexSort() error {
	indexSort := w.config.GetIndexSort()
	if indexSort != nil {
		for _, info := range w.segmentInfos.segments {
			segmentIndexSort := info.Info().GetIndexSort()
			if segmentIndexSort == nil || isCongruentSort(indexSort, segmentIndexSort) == false {
				return errors.New("cannot change previous indexSort")
			}
		}
	}
	return nil
}

func isCongruentSort(indexSort, otherSort index.Sort) bool {
	fields1 := indexSort.GetSort()
	fields2 := otherSort.GetSort()
	if len(fields1) > len(fields2) {
		return false
	}

	for i := range fields1 {
		if !fields1[i].Equals(fields2[i]) {
			return false
		}
	}
	return true
}

type Merges struct {
	mergesEnabled bool
}

func (m *Merges) areEnabled() bool {
	return m.mergesEnabled
}

func (m *Merges) disable() {
	m.mergesEnabled = false
}

func (m *Merges) enable() {
	m.mergesEnabled = true
}

// AddDocument
// Adds a document to this index.
// Note that if an Exception is hit (for example disk full) then the index will be consistent, but this
// document may not have been added. Furthermore, it's possible the index will have one segment in
// non-compound format even when using compound files (when a merge has partially succeeded).
//
// This method periodically flushes pending documents to the Directory (see above), and also periodically
// triggers segment merges in the index according to the MergePolicy in use.
//
// Merges temporarily consume space in the directory. The amount of space required is up to 1X the size of
// all segments being merged, when no readers/searchers are open against the index, and up to 2X the size
// of all segments being merged when readers/searchers are open against the index (see forceMerge(int) for details).
// The sequence of primitive merge operations performed is governed by the merge policy.
//
// Note that each term in the document can be no longer than MAX_TERM_LENGTH in bytes, otherwise an
// IllegalArgumentException will be thrown.
//
// Note that it's possible to create an invalid Unicode string in java if a UTF16 surrogate pair is
// malformed. In this case, the invalid characters are silently replaced with the Unicode replacement character U+FFFD.
//
// Returns: The sequence number for this operation
// Throws:  CorruptIndexException – if the index is corrupt
//
//	IOException – if there is a low-level IO error
func (w *IndexWriter) AddDocument(ctx context.Context, doc *document.Document) (int64, error) {
	return w.UpdateDocument(ctx, nil, doc)
}

// UpdateDocument
// Updates a document by first deleting the document(s) containing term and then adding
// the new document. The delete and then add are atomic as seen by a reader on the same index
// (Flush may happen only after the add).
//
// term: the term to identify the document(s) to be deleted
// doc: the document to be added
//
// Returns: The sequence number for this operation
// Throws:
//
//	CorruptIndexException – if the index is corrupt
//	IOException – if there is a low-level IO error
func (w *IndexWriter) UpdateDocument(ctx context.Context, term index.Term, doc *document.Document) (int64, error) {
	var delNode *Node
	if term != nil {
		delNode = deleteQueueNewNode(term)
	}
	return w.updateDocuments(ctx, delNode, []*document.Document{doc})
}

// SoftUpdateDocument
// Expert: Updates a document by first updating the document(s) containing term with the given doc-values
// fields and then adding the new document. The doc-values update and then add are atomic as seen by a
// reader on the same index (flush may happen only after the add). One use of this API is to retain older
// versions of documents instead of replacing them. The existing documents can be updated to reflect they
// are no longer current while atomically adding new documents at the same time. In contrast to
// updateDocument(Term, Iterable) this method will not delete documents in the index matching the given term
// but instead update them with the given doc-values fields which can be used as a soft-delete mechanism.
// See addDocuments(Iterable) and updateDocuments(Term, Iterable).
//
// Returns: The sequence number for this operation
// Throws: CorruptIndexException: if the index is corrupt
// IOException: if there is a low-level IO error
func (w *IndexWriter) SoftUpdateDocument(ctx context.Context, term index.Term, doc *document.Document, softDeletes ...document.IndexableField) (int64, error) {
	updates, err := w.buildDocValuesUpdate(term, softDeletes)
	if err != nil {
		return 0, err
	}
	var delNode *Node
	if term != nil {
		delNode = deleteQueueNewNodeDocValuesUpdates(updates)
	}
	return w.updateDocuments(ctx, delNode, []*document.Document{doc})
}

func (w *IndexWriter) buildDocValuesUpdate(term index.Term, updates []document.IndexableField) ([]index.DocValuesUpdate, error) {
	dvUpdates := make([]index.DocValuesUpdate, 0, len(updates))

	for _, field := range updates {
		dvType := field.FieldType().DocValuesType()

		if w.globalFieldNumberMap.contains(field.Name(), dvType) == false {
			// if this field doesn't exists we try to add it. if it exists and the DV type doesn't match we
			// get a consistent error message as if you try to do that during an indexing operation.
			if _, err := w.globalFieldNumberMap.AddOrGet(field.Name(), -1,
				document.INDEX_OPTIONS_NONE, dvType, 0, 0, 0,
				field.Name() == w.config.softDeletesField); err != nil {
				return nil, err
			}
		}

		if _, ok := w.config.GetIndexSortFields()[field.Name()]; ok {
			return nil, errors.New("cannot update docvalues field involved in the index sort")
		}

		switch dvType {
		case document.DOC_VALUES_TYPE_NUMERIC:
			switch v := field.Get().(type) {
			case int32:
				dvUpdates = append(dvUpdates, index.NewNumericDocValuesUpdate(term, field.Name(), int64(v)))
			case int64:
				dvUpdates = append(dvUpdates, index.NewNumericDocValuesUpdate(term, field.Name(), v))
			default:
				panic("TODO")
			}

		case document.DOC_VALUES_TYPE_BINARY:
			value, err := document.Bytes(field.Get())
			if err != nil {
				return nil, err
			}
			dvUpdates = append(dvUpdates, index.NewBinaryDocValuesUpdate(term, field.Name(), value))
		default:
			return nil, errors.New("can only update NUMERIC or BINARY fields")
		}
	}
	return dvUpdates, nil
}

func (w *IndexWriter) Commit(ctx context.Context) error {
	//return w.docWriter.Flush(ctx)
	_, err := w.commitInternal(ctx, w.config.GetMergePolicy())
	return err
}

// Close
// Closes all open resources and releases the write lock. If IndexWriterConfig. commitOnClose is true,
// this will attempt to gracefully shut down by writing any changes, waiting for any running merges,
// committing, and closing. In this case, note that:
//
// If you called prepareCommit but failed to call commit, this method will throw IllegalStateException
// and the IndexWriter will not be closed.
//
// If this method throws any other exception, the IndexWriter will be closed, but changes may have been lost.
// Note that this may be a costly operation, so, try to re-use a single writer instead of closing and opening
// a new one. See commit() for caveats about write caching done by some IO devices.
//
// NOTE: You must ensure no other threads are still making changes at the same time that this method is invoked.
func (w *IndexWriter) Close() error {
	if w.config.GetCommitOnClose() {
		return w.shutdown(context.Background())
	}
	// TODO: rollback
	return w.shutdown(context.Background())
}

func (w *IndexWriter) updateDocuments(ctx context.Context, delNode *Node, docs []*document.Document) (int64, error) {
	seqNo, err := w.docWriter.updateDocuments(ctx, docs, delNode)
	if err != nil {
		return 0, err
	}

	seqNo, err = w.maybeProcessEvents(seqNo)
	if err != nil {
		return 0, err
	}
	return seqNo, nil
}

func (w *IndexWriter) maybeProcessEvents(seqNo int64) (int64, error) {
	if seqNo < 0 {
		seqNo = -seqNo
		if err := w.processEvents(true); err != nil {
			return 0, err
		}
	}
	return seqNo, nil
}

func (w *IndexWriter) processEvents(triggerMerge bool) error {
	if err := w.eventQueue.processEvents(); err != nil {
		return err
	}

	if triggerMerge {
		return w.maybeMerge(w.config.GetMergePolicy(), MERGE_TRIGGER_EXPLICIT, UNBOUNDED_MAX_MERGE_SEGMENTS)
	}
	return nil
}

func (w *IndexWriter) MaybeMerge() error {
	return w.maybeMerge(w.config.GetMergePolicy(), MERGE_TRIGGER_EXPLICIT, UNBOUNDED_MAX_MERGE_SEGMENTS)
}

func (w *IndexWriter) maybeMerge(mergePolicy MergePolicy, trigger MergeTrigger, maxNumSegments int) error {
	err := w.ensureOpenV1(false)
	if err != nil {
		return err
	}

	if _, err := w.updatePendingMerges(mergePolicy, trigger, maxNumSegments); err != nil {
		return w.executeMerge(trigger)
	}
	return nil
}

func (w *IndexWriter) ensureOpen() error {
	// TODO: fix it
	return nil
}

func (w *IndexWriter) ensureOpenV1(failIfClosing bool) error {
	// TODO: fix it
	return nil
}

func (w *IndexWriter) executeMerge(trigger MergeTrigger) error {
	return w.mergeScheduler.Merge(w.mergeSource, trigger)
}

func (w *IndexWriter) updatePendingMerges(mergePolicy MergePolicy, trigger MergeTrigger, maxNumSegments int) (*MergeSpecification, error) {
	panic("")
	/*
		// In case infoStream was disabled on init, but then enabled at some
		// point, try again to log the config here:
		//if err := w.messageState(); err != nil {
		//	return nil, err
		//}

		//assert maxNumSegments == UNBOUNDED_MAX_MERGE_SEGMENTS || maxNumSegments > 0;
		//assert trigger != null;
		if w.merges.areEnabled() == false {
			return nil, errors.New("merges is disable")
		}

		// Do not start new merges if disaster struck
		//if w.tragedy != null {
		//	return null
		//}

		var spec *MergeSpecification
		var err error
		if maxNumSegments != UNBOUNDED_MAX_MERGE_SEGMENTS {
			// TODO:
		} else {
			switch trigger {
			case GET_READER:
			case COMMIT:
				spec, err = mergePolicy.FindFullFlushMerges(trigger, w.segmentInfos, w)
				if err != nil {
					return nil, err
				}
				break
			default:
				spec, err = mergePolicy.FindMerges(trigger, w.segmentInfos, w)
			}
		}

		return nil
	*/
}

func (w *IndexWriter) newSegmentName() string {
	w.changeCount.Add(1)
	w.segmentInfos.Changed()
	v := w.segmentInfos.counter
	w.segmentInfos.counter++
	return fmt.Sprintf("_%s", strconv.FormatInt(v, 36))
}

func (w *IndexWriter) getFieldNumberMap() (*FieldNumbers, error) {
	mp := NewFieldNumbers(w.config.softDeletesField)

	for _, info := range w.segmentInfos.segments {
		fis, err := readFieldInfos(info)
		if err != nil {
			return nil, err
		}
		for _, fi := range fis.List() {
			if _, err := mp.AddOrGet(fi.Name(), fi.Number(), fi.GetIndexOptions(), fi.GetDocValuesType(),
				fi.GetPointDimensionCount(), fi.GetPointIndexDimensionCount(),
				fi.GetPointNumBytes(), fi.IsSoftDeletesField()); err != nil {
				return nil, err
			}
		}
	}
	return mp, nil
}

// Gracefully closes (commits, waits for merges), but calls rollback if there's an exc so
// the IndexWriter is always closed. This is called from close when IndexWriterConfig.commitOnClose is true.
func (w *IndexWriter) shutdown(ctx context.Context) error {
	if w.pendingCommit != nil {
		return errors.New("cannot close: prepareCommit was already called with no corresponding call to commit")
	}

	if w.shouldClose(true) {
		if err := w.flush(true, true); err != nil {
			return err
		}
		if err := w.waitForMerges(); err != nil {
			return err
		}
		if _, err := w.commitInternal(ctx, w.config.GetMergePolicy()); err != nil {
			return err
		}
	}

	// Ensure that only one thread actually gets to do the closing

	// TODO:
	//err := w.docWriter.Flush(ctx)
	//if err != nil {
	//	return err
	//}
	return w.rollbackInternal(ctx)
}

func (w *IndexWriter) rollbackInternal(ctx context.Context) error {
	return w.rollbackInternalNoCommit(ctx)
}

func (w *IndexWriter) rollbackInternalNoCommit(ctx context.Context) error {
	// Must pre-close in case it increments changeCount so that we can then
	// set it to false before calling rollbackInternal
	err := w.mergeScheduler.Close()
	if err != nil {
		return err
	}

	w.docWriter.Close() // mark it as closed first to prevent subsequent indexing actions/flushes
	w.docWriter.Abort() // don't sync on IW here
	//w.docWriter.flushControl.waitForFlush(); // wait for all concurrently running flushes
	if err := w.publishFlushedSegments(true); err != nil {
		return err
	}

	if w.pendingCommit != nil {
		if err := w.pendingCommit.RollbackCommit(w.directory); err != nil {
			return err
		}
		//w.deleter.decRef(pendingCommit);
		//try {
		//
		//} finally {
		//	pendingCommit = null;
		//	notifyAll();
		//}
	}

	totalMaxDoc := w.segmentInfos.TotalMaxDoc()
	// Keep the same segmentInfos instance but replace all
	// of its SegmentInfo instances so IFD below will remove
	// any segments we flushed since the last commit:
	if err := w.segmentInfos.rollbackSegmentInfos(w.rollbackSegments); err != nil {
		return err
	}
	rollbackMaxDoc := w.segmentInfos.TotalMaxDoc()
	// now we need to adjust this back to the rolled back SI but don't set it to the absolute value
	// otherwise we might hide internal bugsf
	w.adjustPendingNumDocs(-(totalMaxDoc - rollbackMaxDoc))

	return nil
}

// Returns true if this thread should attempt to close, or
// false if IndexWriter is now closed; else,
// waits until another thread finishes closing
func (w *IndexWriter) shouldClose(waitForClose bool) bool {
	for {
		if w.closed == false {
			if w.closing == false {
				// We get to close
				w.closing = true
				return true
			} else if waitForClose == false {
				return false
			} else {
				// Another thread is presently trying to close;
				// wait until it finishes one way (closes
				// successfully) or another (fails to close)
				w.doWait()
			}
		} else {
			return false
		}
	}
}

// Wait for any currently outstanding merges to finish.
// It is guaranteed that any merges started prior to calling this method will have completed once this method completes.
func (w *IndexWriter) waitForMerges() error {
	// Give merge scheduler last chance to run, in case
	// any pending merges are waiting. We can't hold IW's lock
	// when going into merge because it can lead to deadlock.
	if err := w.mergeScheduler.Merge(w.mergeSource, MERGE_TRIGGER_CLOSING); err != nil {
		return err
	}

	// FIXME: w.runningMerges.size() > 0
	for len(w.pendingMerges) > 0 {
		w.doWait()
	}

	return nil
}

// FIXME: how to
func (w *IndexWriter) doWait() {
	// NOTE: the callers of this method should in theory
	// be able to do simply wait(), but, as a defense
	// against thread timing hazards where notifyAll()
	// fails to be called, we wait for at most 1 second
	// and then return so caller can check if wait
	// conditions are satisfied:

}

func (w *IndexWriter) commitInternal(ctx context.Context, mergePolicy MergePolicy) (int64, error) {
	var seqNo int64
	var err error
	if w.pendingCommit == nil {
		seqNo, err = w.prepareCommitInternal()
		if err != nil {
			return 0, err
		}
	} else {
		seqNo = w.pendingSeqNo
	}

	if err := w.finishCommit(ctx); err != nil {
		return 0, err
	}

	if w.atomMaybeMerge.Swap(false) {
		err := w.maybeMerge(mergePolicy, MERGE_TRIGGER_FULL_FLUSH, UNBOUNDED_MAX_MERGE_SEGMENTS)
		if err != nil {
			return 0, err
		}
	}

	return seqNo, nil
}

func (w *IndexWriter) Changed() {
	w.changeCount.Add(1)
	w.segmentInfos.Changed()
}

func (w *IndexWriter) IsClosed() bool {
	return w.closed
}

func (w *IndexWriter) nrtIsCurrent(infos *SegmentInfos) bool {
	return false

	// TODO: fix it
	//ensureOpen();
	//isCurrent := infos.GetVersion() == w.segmentInfos.GetVersion() &&
	//	w.docWriter.anyChanges() == false &&
	//	//i.bufferedUpdatesStream.Any() == false &&
	//	w.readerPool.anyDocValuesChanges() == false
	//return isCurrent
}

func (w *IndexWriter) GetReader(ctx context.Context, applyAllDeletes bool, writeAllDeletes bool) (index.DirectoryReader, error) {
	if writeAllDeletes && applyAllDeletes == false {
		return nil, errors.New("applyAllDeletes must be true when writeAllDeletes=true")
	}

	// this function is used to control which SR are opened in order to keep track of them
	// and to reuse them in the case we wait for merges in this getReader call.

	maxFullFlushMergeWaitMillis := w.config.GetMaxFullFlushMergeWaitMillis()
	openedReadOnlyClones := make(map[string]*SegmentReader)

	readerFactory := func(sci index.SegmentCommitInfo) (*SegmentReader, error) {
		rld, err := w.getPooledInstance(sci, true)
		if err != nil {
			return nil, err
		}

		segmentReader, err := rld.GetReader(ctx, nil)
		if err != nil {
			return nil, err
		}
		if maxFullFlushMergeWaitMillis > 0 { // only track this if we actually do fullFlush merges
			openedReadOnlyClones[sci.Info().Name()] = segmentReader
		}
		return segmentReader, nil
	}

	return OpenStandardDirectoryReader(w, readerFactory, w.segmentInfos, applyAllDeletes, writeAllDeletes)
}

func (w *IndexWriter) Release(readersAndUpdates *ReadersAndUpdates) error {
	return w.release(readersAndUpdates, true)
}

func (w *IndexWriter) release(readersAndUpdates *ReadersAndUpdates, assertLiveInfo bool) error {
	//if i.readerPool.
	panic("")
}

func (w *IndexWriter) doBeforeFlush() error {
	return nil
}

func (w *IndexWriter) getPooledInstance(info index.SegmentCommitInfo, create bool) (*ReadersAndUpdates, error) {
	return w.readerPool.Get(info, create)
}

// Publishes the flushed segment, segment-private deletes (if any) and its associated global delete
// (if present) to IndexWriter. The actual publishing operation is synced on IW -> BDS so that the
// SegmentInfo's delete generation is always GlobalPacket_deleteGeneration + 1
//
// forced: if true this call will block on the ticket queue if the lock is held by another thread.
// if false the call will try to acquire the queue lock and exits if it's held by another thread.
// FIXME: 需要完善
func (w *IndexWriter) publishFlushedSegments(forced bool) error {
	err := w.docWriter.purgeFlushTickets(forced, func(ticket *FlushTicket) error {
		newSegment := ticket.getFlushedSegment()
		bufferedUpdates := ticket.getFrozenUpdates()
		ticket.markPublished()
		if newSegment == nil { // this is a flushed global deletes package - not a segments
			if bufferedUpdates != nil && bufferedUpdates.Any() { // TODO why can this be null?
				w.publishFrozenUpdates(bufferedUpdates)
			}
		} else {
			// now publish!
			err := w.publishFlushedSegment(newSegment.segmentInfo, newSegment.fieldInfos, newSegment.segmentUpdates,
				bufferedUpdates, newSegment.sortMap)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (w *IndexWriter) applyAllDeletesAndUpdates() error {
	w.flushDeletesCount.Add(1)
	return nil
}

// Ensures that all changes in the reader-pool are written to disk.
func (w *IndexWriter) writeReaderPool(writeDeletes bool) error {
	if writeDeletes {
		ok, err := w.readerPool.commit(w.segmentInfos)
		if err != nil {
			return err
		}
		if ok {
			if err := w.checkpointNoSIS(); err != nil {
				return err
			}
		}
	} else {
		ok, err := w.readerPool.writeAllDocValuesUpdates()
		if err != nil {
			return err
		}
		if ok {
			if err := w.checkpoint(); err != nil {
				return nil
			}
		}
	}

	// now do some best effort to check if a segment is fully deleted
	toDrop := make([]index.SegmentCommitInfo, 0)
	for _, info := range w.segmentInfos.AsList() {
		readersAndUpdates, err := w.readerPool.Get(info, false)
		if err != nil {
			return err
		}

		if readersAndUpdates != nil {
			deleted, err := w.isFullyDeleted(readersAndUpdates)
			if err != nil {
				return err
			}

			if deleted {
				toDrop = append(toDrop, info)
			}
			//if (isFullyDeleted(readersAndUpdates)) {
			//	toDrop.add(info);
			//}

		}
	}

	for _, info := range toDrop {
		err := w.dropDeletedSegment(info)
		if err != nil {
			return err
		}
	}

	if len(toDrop) != 0 {
		err := w.checkpoint()
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *IndexWriter) doAfterFlush() error {
	return nil
}

// GetDirectory
// Returns the Directory used by this index.
func (w *IndexWriter) GetDirectory() store.Directory {
	// return the original directory the user supplied, unwrapped.
	return w.directoryOrig
}

func (w *IndexWriter) GetConfig() *IndexWriterConfig {
	return w.config
}

func (w *IndexWriter) IncRefDeleter(segmentInfos *SegmentInfos) error {
	return nil
	//return w.deleter.IncRef(segmentInfos, false)
}

// AddIndexesFromReaders
// Merges the provided indexes into this index.
// The provided IndexReaders are not closed.
// See addIndexes for details on transactional semantics, temporary free space required in the Directory,
// and non-CFS segments on an Exception.
// NOTE: empty segments are dropped by this method and not added to this index.
// NOTE: this merges all given LeafReaders in one merge. If you intend to merge a large number of readers, it may be better to call this method multiple times, each time with a small set of readers. In principle, if you use a merge policy with a mergeFactor or maxMergeAtOnce parameter, you should pass that many readers in one call.
// NOTE: this method does not call or make use of the MergeScheduler, so any custom bandwidth throttling is at the moment ignored.
func (w *IndexWriter) AddIndexesFromReaders(readers ...index.CodecReader) (int64, error) {

	if err := w.ensureOpen(); err != nil {
		return 0, err
	}

	// long so we can detect int overflow:
	var seqNo int64
	var numDocs int

	// FIXME:
	if err := w.flush(false, true); err != nil {
		return 0, err
	}

	mergedName := w.newSegmentName()
	numSoftDeleted := 0
	for _, leaf := range readers {
		numDocs += leaf.NumDocs()
		//w.validateMergeReader(leaf);
		if w.softDeletesEnabled {
			liveDocs := leaf.GetLiveDocs()

			docIdSetIterator, err := getDocValuesDocIdSetIterator(w.config.getSoftDeletesField(), leaf)
			if err != nil {
				return 0, err
			}

			softDeletes, err := countSoftDeletes(docIdSetIterator, liveDocs)
			if err != nil {
				return 0, err
			}

			numSoftDeleted += softDeletes
		}
	}

	// Best-effort up front check:
	if err := w.testReserveDocs(int64(numDocs)); err != nil {
		return 0, err
	}

	mergeInfo := store.NewMergeInfo(numDocs, -1, false, UNBOUNDED_MAX_MERGE_SEGMENTS)
	ioCtx := store.NewIOContext(store.WithMergeInfo(mergeInfo))

	// TODO: somehow we should fix this merge so it's
	// abortable so that IW.close(false) is able to stop it

	codec := w.config.GetCodec()
	// We set the min version to null for now, it will be set later by SegmentMerger
	info := NewSegmentInfo(w.directoryOrig, version.Last, nil, mergedName, -1,
		false, codec, map[string]string{}, util.RandomId(), map[string]string{}, w.config.GetIndexSort())

	merger, err := NewSegmentMerger(readers, info, w.directory, w.globalFieldNumberMap, ioCtx)
	if err != nil {
		return 0, err
	}

	if !merger.ShouldMerge() {
		// TODO: fix it
		//return docWriter.getNextSequenceNumber()
	}

	if err := w.MaybeMerge(); err != nil {
		return 0, err
	}
	return seqNo, nil
}

// Does a best-effort check, that the current index would accept this many additional docs, but does not actually reserve them.
func (w *IndexWriter) testReserveDocs(addedNumDocs int64) error {
	if w.pendingNumDocs.Load()+addedNumDocs > int64(actualMaxDocs) {
		return w.tooManyDocs(addedNumDocs)
	}
	return nil
}

func (w *IndexWriter) tooManyDocs(addedNumDocs int64) error {
	return fmt.Errorf("number of documents in the index cannot exceed %d (current document count is %d; added numDocs is %d)",
		actualMaxDocs, w.pendingNumDocs.Load(), addedNumDocs)
}

// Flush all in-memory buffered updates (adds and deletes) to the Directory.
func (w *IndexWriter) flush(triggerMerge, applyAllDeletes bool) error {
	// NOTE: this method cannot be sync'd because
	// maybeMerge() in turn calls mergeScheduler.merge which
	// in turn can take a long time to run and we don't want
	// to hold the lock for that.  In the case of
	// ConcurrentMergeScheduler this can lead to deadlock
	// when it stalls due to too many running merges.

	// We can be called during close, when closing==true, so we must pass false to ensureOpen:
	doFlush, err := w.doFlush(applyAllDeletes)
	if err != nil {
		return err
	}
	if doFlush && triggerMerge {
		return w.maybeMerge(w.config.mergePolicy, MERGE_TRIGGER_FULL_FLUSH, UNBOUNDED_MAX_MERGE_SEGMENTS)
	}
	return nil
}

func (w *IndexWriter) doFlush(applyAllDeletes bool) (bool, error) {
	err := w.doBeforeFlush()
	if err != nil {
		return false, err
	}

	/*
		long seqNo = docWriter.flushAllThreads() ;
		          if (seqNo < 0) {
		            seqNo = -seqNo;
		            anyChanges = true;
		          } else {
		            anyChanges = false;
		          }
		          if (!anyChanges) {
		            // flushCount is incremented in flushAllThreads
		            flushCount.incrementAndGet();
		          }
		          publishFlushedSegments(true);
		          flushSuccess = true;
	*/
	anyChanges := false
	seqNo := w.docWriter.flushAllThreads()
	if seqNo < 0 {
		seqNo = -seqNo
		anyChanges = true
	}

	if !anyChanges {
		// flushCount is incremented in flushAllThreads
		w.flushCount.Add(1)
	}
	err = w.publishFlushedSegments(true)
	if err != nil {
		return false, err
	}
	err = w.docWriter.finishFullFlush(true)
	if err != nil {
		return false, err
	}
	err = w.processEvents(false)
	if err != nil {
		return false, err
	}

	if applyAllDeletes {
		err = w.applyAllDeletesAndUpdates()
		if err != nil {
			return false, err
		}
	}

	anyChanges = anyChanges || w.boolMaybeMerge.Swap(false)
	if anyChanges {
		err := w.maybeMerge(w.config.GetMergePolicy(), MERGE_TRIGGER_FULL_FLUSH, UNBOUNDED_MAX_MERGE_SEGMENTS)
		if err != nil {
			return false, err
		}
	}

	return false, nil
}

// Tries to delete the given files if unreferenced
func (w *IndexWriter) deleteNewFiles(files map[string]struct{}) error {
	return w.deleter.deleteFiles(files)
}

func (w *IndexWriter) flushFailed(info *SegmentInfo) error {
	files := info.Files()
	return w.deleter.deleteNewFiles(files)
}

func getDocValuesDocIdSetIterator(field string, reader index.LeafReader) (types.DocIdSetIterator, error) {
	panic("")
}

func readFieldInfos(si index.SegmentCommitInfo) (index.FieldInfos, error) {
	codec := si.Info().GetCodec()
	reader := codec.FieldInfosFormat()

	if si.HasFieldUpdates() {

	} else if si.Info().GetUseCompoundFile() {
		cfs, err := codec.CompoundFormat().GetCompoundReader(nil, si.Info().Dir(), si.Info(), nil)
		if err != nil {
			return nil, err
		}
		return reader.Read(nil, cfs, si.Info(), "", nil)
	}

	return reader.Read(nil, si.Info().Dir(), si.Info(), "", nil)
}

func GetActualMaxDocs() int {
	return actualMaxDocs
}

// Walk through all files referenced by the current segmentInfos and ask the Directory to sync each file,
// if it wasn't already. If that succeeds, then we prepare a new segments_N file but do not fully commit it.
func (w *IndexWriter) startCommit(toSync *SegmentInfos) error {
	if w.lastCommitChangeCount.Load() > w.changeCount.Load() {
		return fmt.Errorf("lastCommitChangeCount=%d ,changeCount=%d",
			w.lastCommitChangeCount, w.changeCount)
	}

	if w.pendingCommitChangeCount == w.lastCommitChangeCount.Load() {
		err := w.deleter.DecRef(w.filesToCommit)
		if err != nil {
			return err
		}
		w.filesToCommit = nil
		return nil
	}

	// Exception here means nothing is prepared
	// (this method unwinds everything it did on
	// an exception)
	err := toSync.prepareCommit(context.Background(), w.directory)
	if err != nil {
		return err
	}
	w.pendingCommit = toSync

	filesToSync, err := toSync.Files(false)
	err = w.directory.Sync(filesToSync)
	if err != nil {
		return err
	}

	w.segmentInfos.UpdateGeneration(toSync)
	return nil
}

func (w *IndexWriter) finishCommit(ctx context.Context) error {
	if w.pendingCommit != nil {
		commitFiles := w.filesToCommit

		err := w.deleter.DecRef(commitFiles)
		if err != nil {
			return err
		}

		_, err = w.pendingCommit.finishCommit(ctx, w.directory)
		if err != nil {
			return err
		}

		// we committed, if anything goes wrong after this, we are screwed and it's a tragedy:
		//commitCompleted := true

		// NOTE: don't use this.checkpoint() here, because
		// we do not want to increment changeCount:
		err = w.deleter.Checkpoint(w.pendingCommit, true)
		if err != nil {
			return err
		}

		// Carry over generation to our master SegmentInfos:
		w.segmentInfos.UpdateGeneration(w.pendingCommit)

		w.lastCommitChangeCount.Store(w.pendingCommitChangeCount)
		w.rollbackSegments = w.pendingCommit.CreateBackupSegmentInfos()

		w.pendingCommit = nil
		w.filesToCommit = nil
	}
	return nil
}

func (w *IndexWriter) prepareCommitInternal() (int64, error) {
	err := w.doBeforeFlush()
	if err != nil {
		return 0, err
	}

	var anyChanges bool

	seqNo := w.docWriter.flushAllThreads()
	if seqNo < 0 {
		anyChanges = true
		seqNo = -seqNo
	}

	if anyChanges == false {
		// prevent double increment since docWriter#doFlush increments the flushcount
		// if we flushed anything.
		w.flushCount.Add(1)
	}

	err = w.publishFlushedSegments(true)
	if err != nil {
		return 0, err
	}
	// cannot pass triggerMerges=true here else it can lead to deadlock:
	err = w.processEvents(false)
	if err != nil {
		return 0, err
	}

	err = w.applyAllDeletesAndUpdates()
	if err != nil {
		return 0, err
	}

	err = w.writeReaderPool(true)
	if err != nil {
		return 0, err
	}

	if w.changeCount.Load() != w.lastCommitChangeCount.Load() {
		// There are changes to commit, so we will write a new segments_N in startCommit.
		// The act of committing is itself an NRT-visible change (an NRT reader that was
		// just opened before this should see it on reopen) so we increment changeCount
		// and segments version so a future NRT reopen will see the change:
		w.changeCount.Add(1)
		w.segmentInfos.Changed()
	}

	if w.commitUserData != nil {
		userData := maps.Clone(w.commitUserData)
		w.segmentInfos.SetUserData(userData, false)
	}

	// Must clone the segmentInfos while we still
	// hold fullFlushLock and while sync'd so that
	// no partial changes (eg a delete w/o
	// corresponding add from an updateDocument) can
	// sneak into the commit point:
	toCommit := w.segmentInfos.Clone()
	w.pendingCommitChangeCount = w.changeCount.Load()
	// This protects the segmentInfos we are now going
	// to commit.  This is important in case, eg, while
	// we are trying to sync all referenced files, a
	// merge completes which would otherwise have
	// removed the files we are now syncing.
	files, err := toCommit.Files(false)
	if err != nil {
		return 0, err
	}
	err = w.deleter.IncRefFiles(files)
	if err != nil {
		return 0, err
	}
	if anyChanges {
		// we can safely call preparePointInTimeMerge since writeReaderPool(true) above wrote all
		// necessary files to disk and checkpointed them.
		//pointInTimeMerges = w.preparePointInTimeMerge(toCommit, stopAddingMergedSegments::get, MergeTrigger.COMMIT, sci->{});
	}

	// Done: finish the full flush!
	err = w.docWriter.finishFullFlush(true)
	if err != nil {
		return 0, err
	}
	err = w.doAfterFlush()
	if err != nil {
		return 0, err
	}

	// do this after handling any pointInTimeMerges since the files will have changed if any merges
	// did complete
	filesToCommit, err := toCommit.Files(false)
	if err != nil {
		return 0, err
	}
	w.filesToCommit = filesToCommit

	if anyChanges {
		w.boolMaybeMerge.Store(true)
	}
	err = w.startCommit(toCommit)
	if err != nil {
		return 0, err
	}
	if w.pendingCommit == nil {
		return -1, nil
	} else {
		return seqNo, nil
	}
}

type KV struct {
	Key   string
	Value string
}

func (w *IndexWriter) publishFrozenUpdates(updates *FrozenBufferedUpdates) int64 {
	// TODO: fixme
	return -1
}

// Atomically adds the segment private delete packet and publishes the flushed segments SegmentInfo to the index writer.
func (w *IndexWriter) publishFlushedSegment(newSegment index.SegmentCommitInfo, fieldInfos index.FieldInfos,
	packet *FrozenBufferedUpdates, globalPacket *FrozenBufferedUpdates, sortMap index.DocMap) error {

	published := false

	if globalPacket != nil && globalPacket.Any() {
		w.publishFrozenUpdates(globalPacket)
	}

	// Publishing the segment must be sync'd on IW -> BDS to make the sure
	// that no merge prunes away the seg. private delete packet
	var nextGen int64
	if packet != nil && packet.Any() {
		nextGen = w.publishFrozenUpdates(packet)
	} else {
		// Since we don't have a delete packet to apply we can get a new
		// generation right away
		nextGen = w.bufferedUpdatesStream.GetNextGen()
		// No deletes/updates here, so marked finished immediately:
		w.bufferedUpdatesStream.FinishedSegment(nextGen)
	}

	newSegment.SetBufferedDeletesGen(nextGen)
	w.segmentInfos.Add(newSegment)
	published = true
	w.checkpoint()
	if packet != nil && packet.Any() && sortMap != nil {
		// TODO: not great we do this heavyish op while holding IW's monitor lock,
		// but it only applies if you are using sorted indices and updating doc values:
		rld, err := w.getPooledInstance(newSegment, true)
		if err != nil {
			return err
		}
		rld.sortMap = sortMap
		// DON't release this ReadersAndUpdates we need to stick with that sortMap
	}
	fieldInfo := fieldInfos.FieldInfo(w.config.softDeletesField) // will return null if no soft deletes are present
	// this is a corner case where documents delete them-self with soft deletes. This is used to
	// build delete tombstones etc. in this case we haven't seen any updates to the DV in this fresh flushed segment.
	// if we have seen updates the update code checks if the segment is fully deleted.
	hasInitialSoftDeleted := fieldInfo != nil &&
		fieldInfo.GetDocValuesGen() == -1 &&
		fieldInfo.GetDocValuesType() != document.DOC_VALUES_TYPE_NONE

	infoMaxCount, err := newSegment.Info().MaxDoc()
	if err != nil {
		return err
	}
	isFullyHardDeleted := newSegment.GetDelCount() == infoMaxCount
	// we either have a fully hard-deleted segment or one or more docs are soft-deleted. In both cases we need
	// to go and check if they are fully deleted. This has the nice side-effect that we now have accurate numbers
	// for the soft delete right after we flushed to disk.
	if hasInitialSoftDeleted || isFullyHardDeleted {
		// this operation is only really executed if needed an if soft-deletes are not configured it only be executed
		// if we deleted all docs in this newly flushed segment.
		rld, err := w.getPooledInstance(newSegment, true)
		if err != nil {
			return err
		}
		if ok, _ := w.isFullyDeleted(rld); ok {
			w.dropDeletedSegment(newSegment)
			w.checkpoint()
		}
		w.release(rld, true)
	}

	if published == false {
		maxDoc, err := newSegment.Info().MaxDoc()
		if err != nil {
			return err
		}
		w.adjustPendingNumDocs(int64(-maxDoc))
	}
	w.flushCount.Add(1)
	w.doAfterFlush()

	return nil
}

func (w *IndexWriter) checkpointNoSIS() error {
	panic("")
}

func (w *IndexWriter) checkpoint() error {
	w.changed()
	return w.deleter.Checkpoint(w.segmentInfos, false)
}

func (w *IndexWriter) dropDeletedSegment(info index.SegmentCommitInfo) error {
	panic("")
}

func (w *IndexWriter) isFullyDeleted(readersAndUpdates *ReadersAndUpdates) (bool, error) {
	isFullyDeleted, err := readersAndUpdates.IsFullyDeleted()
	if err != nil {
		return false, err
	}
	return isFullyDeleted, nil
}

func (w *IndexWriter) adjustPendingNumDocs(numDocs int64) int64 {
	count := w.pendingNumDocs.Add(numDocs)
	return count
}

func (w *IndexWriter) changed() {
	w.changeCount.Add(1)
	w.segmentInfos.Changed()
}

// ReaderWarmer
// If DirectoryReader.open(IndexWriter) has been called (ie, this writer is in near real-time mode),
// then after a merge completes, this class can be invoked to warm the reader on the newly merged segment,
// before the merge commits. This is not required for near real-time search, but will reduce search latency
// on opening a new near real-time reader after a merge completes.
//
// lucene.experimental
//
// NOTE: Warm(LeafReader) is called before any deletes have been carried over to the merged segment.
type ReaderWarmer interface {
	Warm(reader index.LeafReader) error
}

type IndexWriterConfig struct {
	*liveIndexWriterConfig

	//sync.Once

	// indicates whether this config instance is already attached to a writer.
	// not final so that it can be cloned properly.
	writer *IndexWriter

	flushPolicy FlushPolicy
}

func NewIndexWriterConfig(codec index.Codec, similarity index.Similarity) *IndexWriterConfig {
	cfg := &IndexWriterConfig{}
	analyzer := standard.NewAnalyzer(analysis.EmptySet)
	cfg.liveIndexWriterConfig = newLiveIndexWriterConfig(analyzer, codec, similarity)
	return cfg
}

func (c *IndexWriterConfig) setIndexWriter(writer *IndexWriter) {
	c.writer = writer
}

func (c *IndexWriterConfig) getSoftDeletesField() string {
	return c.softDeletesField
}

// SetIndexSort
// Set the Sort order to use for all (flushed and merged) segments.
func (c *IndexWriterConfig) SetIndexSort(sort index.Sort) error {
	fields := make(map[string]struct{})
	for _, sortField := range sort.GetSort() {
		if sortField.GetIndexSorter() == nil {
			return fmt.Errorf("cannot sort index with sort field: %s", sortField)
		}
		fields[sortField.GetField()] = struct{}{}
	}

	c.indexSort = sort
	c.indexSortFields = fields
	return nil
}

func (c *IndexWriterConfig) GetIndexCreatedVersionMajor() int {
	return c.createdVersionMajor
}

// GetCommitOnClose Returns true if IndexWriter.close() should first commit before closing.
func (c *IndexWriterConfig) GetCommitOnClose() bool {
	return c.commitOnClose
}

// GetIndexCommit Returns the IndexCommit as specified in IndexWriterConfig.setIndexCommit(IndexCommit)
// or the default, null which specifies to open the latest index commit point.
func (c *IndexWriterConfig) GetIndexCommit() IndexCommit {
	return c.commit
}

func (c *IndexWriterConfig) GetMergeScheduler() MergeScheduler {
	return c.mergeScheduler
}

func (c *IndexWriterConfig) GetOpenMode() OpenMode {
	return c.openMode
}

func (c *IndexWriterConfig) GetFlushPolicy() FlushPolicy {
	return c.flushPolicy
}

const (

	// DISABLE_AUTO_FLUSH
	// Denotes a Flush trigger is disabled.
	DISABLE_AUTO_FLUSH = -1

	// DEFAULT_MAX_BUFFERED_DELETE_TERMS
	// Disabled by default (because IndexWriter flushes by RAM usage by default).
	DEFAULT_MAX_BUFFERED_DELETE_TERMS = DISABLE_AUTO_FLUSH

	// DEFAULT_MAX_BUFFERED_DOCS
	// Disabled by default (because IndexWriter flushes by RAM usage by default).
	DEFAULT_MAX_BUFFERED_DOCS = DISABLE_AUTO_FLUSH

	// DEFAULT_RAM_BUFFER_SIZE_MB
	// Default item is 16 MB (which means Flush when buffered docs consume approximately 16 MB RAM).
	DEFAULT_RAM_BUFFER_SIZE_MB = 16.0

	// DEFAULT_READER_POOLING
	// Default setting (true) for setReaderPooling.
	// We Changed this default to true with concurrent deletes/updates (LUCENE-7868),
	// because we will otherwise need to open and close segment readers more frequently.
	// False is still supported, but will have worse performance since readers will
	// be forced to aggressively move all state to disk.
	DEFAULT_READER_POOLING = true

	// DEFAULT_RAM_PER_THREAD_HARD_LIMIT_MB
	// Default item is 1945. Change using setRAMPerThreadHardLimitMB(int)
	DEFAULT_RAM_PER_THREAD_HARD_LIMIT_MB = 1945

	// DEFAULT_USE_COMPOUND_FILE_SYSTEM
	// Default item for compound file system for newly
	// written segments (set to true). For batch indexing with very large ram buffers use false
	DEFAULT_USE_COMPOUND_FILE_SYSTEM = true

	// DEFAULT_COMMIT_ON_CLOSE
	// Default item for whether calls to IndexWriter.close() include a commit.
	DEFAULT_COMMIT_ON_CLOSE = true

	// DEFAULT_MAX_FULL_FLUSH_MERGE_WAIT_MILLIS
	// Default item for time to wait for merges
	// on commit or getReader (when using a MergePolicy that implements MergePolicy.findFullFlushMerges).
	DEFAULT_MAX_FULL_FLUSH_MERGE_WAIT_MILLIS = 0
)
