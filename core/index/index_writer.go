package index

import (
	"context"
	"errors"
	"fmt"
	"github.com/emirpasic/gods/maps/hashmap"
	"github.com/emirpasic/gods/sets/hashset"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/store"
	"go.uber.org/atomic"
	"math"
	"strconv"
	"sync"
)

var (
	// Use package-private instance var to enforce the limit so testing
	// can use less electricity:
	actualMaxDocs = MAX_POSITION
)

const (
	// MAX_DOCS Hard limit on maximum number of documents that may be added to the index.
	// If you try to add more than this you'll hit IllegalArgumentException
	MAX_DOCS = math.MaxInt32 - 128

	// MAX_POSITION Maximum value of the token position in an indexed field.
	MAX_POSITION = math.MaxInt32 - 128

	UNBOUNDED_MAX_MERGE_SEGMENTS = -1

	// WRITE_LOCK_NAME Name of the write lock in the index.
	WRITE_LOCK_NAME = "write.lock"

	// SOURCE Key for the source of a segment in the diagnostics.
	SOURCE = "source"

	// SOURCE_MERGE Source of a segment which results from a merge of other segments.
	SOURCE_MERGE = "merge"

	// SOURCE_FLUSH Source of a segment which results from a Flush.
	SOURCE_FLUSH = "Flush"

	// SOURCE_ADDINDEXES_READERS Source of a segment which results from a call to addIndexes(CodecReader...).
	SOURCE_ADDINDEXES_READERS = "addIndexes(CodecReader...)"

	BYTE_BLOCK_SHIFT = 15
	BYTE_BLOCK_SIZE  = 1 << BYTE_BLOCK_SHIFT

	// MAX_TERM_LENGTH Absolute hard maximum length for a term, in bytes once encoded as UTF8.
	// If a term arrives from the analyzer longer than this length, an IllegalArgumentException
	// is thrown and a message is printed to infoStream, if set (see IndexWriterConfig.setInfoStream(InfoStream)).
	MAX_TERM_LENGTH = BYTE_BLOCK_SIZE - 2

	MAX_STORED_STRING_LENGTH = math.MaxInt
)

type IndexWriter struct {
	enableTestPoints         bool
	directoryOrig            store.Directory
	directory                store.Directory
	changeCount              *atomic.Int64
	lastCommitChangeCount    int64
	rollbackSegments         []*SegmentCommitInfo
	pendingCommit            *SegmentInfos
	pendingSeqNo             int64
	pendingCommitChangeCount int64
	filesToCommit            []string
	segmentInfos             *SegmentInfos
	globalFieldNumberMap     *FieldNumbers
	docWriter                *DocumentsWriter
	eventQueue               *EventQueue
	mergeSource              MergeSource
	writeDocValuesLock       sync.RWMutex
	deleter                  *IndexFileDeleter
	segmentsToMerge          *hashmap.Map
	mergeMaxNumSegments      int
	writeLock                store.Lock
	closed                   bool
	closing                  bool
	_maybeMerge              *atomic.Bool
	commitUserData           []map[string]string
	mergingSegments          *hashset.Set
	mergeScheduler           MergeScheduler
	runningAddIndexesMerges  *hashset.Set
	pendingMerges            []*OneMerge
	runningMerges            *hashset.Set
	mergeExceptions          []*OneMerge
	mergeGen                 int64
	merges                   *Merges
	didMessageState          bool
	flushCount               *atomic.Int64
	flushDeletesCount        *atomic.Int64
	readerPool               *ReaderPool
	mergeFinishedGen         *atomic.Int64
	config                   *IndexWriterConfig
	startCommitTime          int64
	pendingNumDocs           *atomic.Int64
	softDeletesEnabled       bool
}

func NewIndexWriter(d store.Directory, conf *IndexWriterConfig) (*IndexWriter, error) {
	writer := &IndexWriter{
		changeCount:    atomic.NewInt64(0),
		pendingNumDocs: atomic.NewInt64(0),
	}
	conf.setIndexWriter(writer)
	writer.config = conf
	writer.softDeletesEnabled = conf.getSoftDeletesField() != ""

	writer.directoryOrig = d
	writer.directory = d
	writer.mergeScheduler = writer.config.GetMergeScheduler()
	writer.mergeScheduler.Initialize(writer.directoryOrig)
	mode := writer.config.GetOpenMode()

	var err error
	var indexExists, create bool
	switch mode {
	case CREATE:
		indexExists, err = IndexExists(writer.directory)
		create = true
	case APPEND:
		indexExists = true
		create = false
	default:
		// CREATE_OR_APPEND - create only if an index does not exist
		indexExists, err = IndexExists(writer.directory)
		create = !indexExists
	}

	// If index is too old, reading the segments will throw
	// IndexFormatTooOldException.
	files, err := writer.directory.ListAll()
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
		sis := NewSegmentInfos(writer.config.GetIndexCreatedVersionMajor())
		if indexExists {
			previous, err := ReadLatestCommit(writer.directory)
			if err != nil {
				return nil, err
			}
			sis.UpdateGenerationVersionAndCounter(previous)
		}
		writer.segmentInfos = sis
		writer.rollbackSegments = writer.segmentInfos.CreateBackupSegmentInfos()

		// Record that we have a change (zero out all
		// segments) pending:
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
		lastCommit, err = ReadCommit(writer.directoryOrig, writer.segmentInfos.GetSegmentsFileName())
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

		lastSegmentsFile := GetLastCommitSegmentsFileName(files)
		if lastSegmentsFile == "" {
			return nil, errors.New("no segments* file found")
		}

		// Do not use SegmentInfos.read(Directory) since the spooky
		// retrying it does is not necessary here (we hold the write lock):
		writer.segmentInfos, err = ReadCommit(writer.directoryOrig, lastSegmentsFile)
		if err != nil {
			return nil, err
		}

		if commit != nil {
			// Swap out all segments, but, keep metadata in
			// SegmentInfos, like version & generation, to
			// preserve write-once.  This is important if
			// readers are open against the future commit
			// points.

			if commit.GetDirectory() != writer.directoryOrig {
				return nil, errors.New("IndexCommit's directory doesn't match my directory")
			}

			oldInfos, err := ReadCommit(writer.directoryOrig, commit.GetSegmentsFileName())
			if err != nil {
				return nil, err
			}
			writer.segmentInfos.Replace(oldInfos)
			writer.Changed()
		}
		writer.rollbackSegments = writer.segmentInfos.CreateBackupSegmentInfos()
	}

	writer.commitUserData = []map[string]string{writer.segmentInfos.GetUserData()}
	writer.pendingNumDocs.Swap(writer.segmentInfos.TotalMaxDoc())

	// start with previous field numbers, but new FieldInfos
	// NOTE: this is correct even for an NRT reader because we'll pull FieldInfos even for the un-committed segments:
	writer.globalFieldNumberMap = writer.getFieldNumberMap()

	if err := writer.validateIndexSort(); err != nil {
		return nil, err
	}

	// TODO: liveIndexWriterConfig 转换成interface
	//writer.config.GetFlushPolicy().Init(writer.config.)

	//writer.segmentInfos = NewSegmentInfos(conf.GetIndexCreatedVersionMajor())
	//
	//writer.globalFieldNumberMap = writer.getFieldNumberMap()

	writer.docWriter = NewDocumentsWriter(writer.segmentInfos.getIndexCreatedVersionMajor(), writer.pendingNumDocs,
		writer.enableTestPoints, writer.newSegmentName,
		writer.config.liveIndexWriterConfig, writer.directoryOrig, writer.directory, writer.globalFieldNumberMap)

	return writer, nil
}

// Confirms that the incoming index sort (if any) matches the existing index sort (if any).
func (i *IndexWriter) validateIndexSort() error {
	indexSort := i.config.GetIndexSort()
	if indexSort != nil {
		for _, info := range i.segmentInfos.segments {
			segmentIndexSort := info.info.GetIndexSort()
			if segmentIndexSort == nil || isCongruentSort(indexSort, segmentIndexSort) == false {
				return errors.New("cannot change previous indexSort")
			}
		}
	}
	return nil
}

func isCongruentSort(indexSort, otherSort *Sort) bool {
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

// AddDocument Adds a document to this index.
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
func (i *IndexWriter) AddDocument(doc *document.Document) (int64, error) {
	return i.UpdateDocument(nil, doc)
}

// UpdateDocument Updates a document by first deleting the document(s) containing term and then adding
// the new document. The delete and then add are atomic as seen by a reader on the same index
// (Flush may happen only after the add).
// Params: term – the term to identify the document(s) to be deleted doc – the document to be added
// Returns: The sequence number for this operation
// Throws: 	CorruptIndexException – if the index is corrupt
//
//	IOException – if there is a low-level IO error
func (i *IndexWriter) UpdateDocument(term *Term, doc *document.Document) (int64, error) {
	var node *Node
	if term != nil {
		node = &Node{item: term}
	}
	return i.updateDocuments(node, []*document.Document{doc})
}

func (i *IndexWriter) Commit(ctx context.Context) error {
	return i.docWriter.Flush(ctx)
}

func (i *IndexWriter) Close() error {
	if i.config.GetCommitOnClose() {
		return i.shutdown()
	}
	return i.shutdown()
}

func (i *IndexWriter) updateDocuments(delNode *Node, docs []*document.Document) (int64, error) {
	seqNo, err := i.docWriter.updateDocuments(docs, delNode)
	if err != nil {
		return 0, err
	}

	seqNo, err = i.maybeProcessEvents(seqNo)
	if err != nil {
		return 0, err
	}
	return seqNo, nil
}

func (i *IndexWriter) maybeProcessEvents(seqNo int64) (int64, error) {
	if seqNo < 0 {
		seqNo = -seqNo
		if err := i.processEvents(true); err != nil {
			return 0, err
		}
	}
	return seqNo, nil
}

func (i *IndexWriter) processEvents(triggerMerge bool) error {
	if err := i.eventQueue.processEvents(); err != nil {
		return err
	}

	if triggerMerge {
		return i.maybeMerge(i.config.GetMergePolicy(), EXPLICIT, UNBOUNDED_MAX_MERGE_SEGMENTS)
	}
	return nil
}

func (i *IndexWriter) maybeMerge(mergePolicy *MergePolicy, trigger MergeTrigger, maxNumSegments int) error {
	err := i.ensureOpenV1(false)
	if err != nil {
		return err
	}

	if i.updatePendingMerges(mergePolicy, trigger, maxNumSegments) != nil {
		return i.executeMerge(trigger)
	}
	return nil
}

func (i *IndexWriter) ensureOpen() error {
	// TODO: fix it
	return nil
}

func (i *IndexWriter) ensureOpenV1(failIfClosing bool) error {
	// TODO: fix it
	return nil
}

func (i *IndexWriter) executeMerge(trigger MergeTrigger) error {
	return i.mergeScheduler.Merge(i.mergeSource, trigger)
}

func (i *IndexWriter) updatePendingMerges(policy *MergePolicy, trigger MergeTrigger, segments int) *MergeSpecification {
	// TODO: impl it
	return nil
}

func (i *IndexWriter) newSegmentName() string {
	i.changeCount.Inc()
	i.segmentInfos.Changed()
	v := i.segmentInfos.counter
	i.segmentInfos.counter++
	return fmt.Sprintf("_%s", strconv.FormatInt(v, 36))
}

func (i *IndexWriter) getFieldNumberMap() *FieldNumbers {
	mp := NewFieldNumbers(i.config.softDeletesField)

	for _, info := range i.segmentInfos.segments {
		fis := readFieldInfos(info)
		for _, fi := range fis.values {
			_, err := mp.AddOrGet(fi.Name(), fi.Number(), fi.GetIndexOptions(), fi.GetDocValuesType(),
				fi.GetPointDimensionCount(), fi.GetPointIndexDimensionCount(),
				fi.GetPointNumBytes(), fi.IsSoftDeletesField())
			if err != nil {
				return nil
			}
		}
	}
	return mp
}

// Gracefully closes (commits, waits for merges), but calls rollback if there's an exc so
// the IndexWriter is always closed. This is called from close when IndexWriterConfig.commitOnClose is true.
func (i *IndexWriter) shutdown() error {
	err := i.docWriter.Flush(nil)
	if err != nil {
		return err
	}
	return nil
}

func (i *IndexWriter) Changed() {
	i.changeCount.Inc()
	i.segmentInfos.Changed()
}

func readFieldInfos(si *SegmentCommitInfo) *FieldInfos {
	//codec := si.info.GetCodec()
	//reader := codec.FieldInfosFormat()
	panic("")
}

func GetActualMaxDocs() int {
	return actualMaxDocs
}

// IndexReaderWarmer If DirectoryReader.open(IndexWriter) has been called (ie, this writer is in near real-time mode), then after a merge completes, this class can be invoked to warm the reader on the newly merged segment, before the merge commits. This is not required for near real-time search, but will reduce search latency on opening a new near real-time reader after a merge completes.
// lucene.experimental
//
// NOTE: warm(LeafReader) is called before any deletes have been carried over to the merged segment.
type IndexReaderWarmer interface {
	Warm(reader LeafReader) error
}
