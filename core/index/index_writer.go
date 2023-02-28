package index

import (
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

	// SOURCE_FLUSH Source of a segment which results from a flush.
	SOURCE_FLUSH = "flush"

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
	commitUserData           any
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

	//files, err := writer.directory.ListAll()
	//if err != nil {
	//	return nil, err
	//}

	writer.segmentInfos = NewSegmentInfos(conf.getIndexCreatedVersionMajor())

	writer.globalFieldNumberMap = writer.getFieldNumberMap()

	writer.docWriter = NewDocumentsWriter(writer.segmentInfos.getIndexCreatedVersionMajor(), writer.pendingNumDocs,
		writer.enableTestPoints, writer.newSegmentName,
		writer.config.LiveIndexWriterConfig, writer.directoryOrig, writer.directory, writer.globalFieldNumberMap)

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
// (flush may happen only after the add).
// Params: term – the term to identify the document(s) to be deleted doc – the document to be added
// Returns: The sequence number for this operation
// Throws: 	CorruptIndexException – if the index is corrupt
//
//	IOException – if there is a low-level IO error
func (i *IndexWriter) UpdateDocument(term *Term, doc *document.Document) (int64, error) {
	var node Node
	if term != nil {
		node = &TermNode{item: term}
	}
	return i.updateDocuments(node, []*document.Document{doc})
}

func (i *IndexWriter) updateDocuments(delNode Node, docs []*document.Document) (int64, error) {
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
			mp.AddOrGet(fi.Name(), fi.Number(), fi.GetIndexOptions(), fi.GetDocValuesType(),
				fi.GetPointDimensionCount(), fi.GetPointIndexDimensionCount(),
				fi.GetPointNumBytes(), fi.IsSoftDeletesField())
		}
	}
	return mp
}

func readFieldInfos(si *SegmentCommitInfo) *FieldInfos {
	//codec := si.info.GetCodec()
	//reader := codec.FieldInfosFormat()
	panic("")
}

func GetActualMaxDocs() int {
	return actualMaxDocs
}

type IndexReaderWarmer func(reader LeafReader) error
