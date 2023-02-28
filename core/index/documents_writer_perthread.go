package index

import (
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util"
	"github.com/pkg/errors"
	"go.uber.org/atomic"
	"io"
	"sync"
)

const (
	INFO_VERBOSE = false
)

type DocumentsWriterPerThread struct {
	codec     Codec
	directory *store.TrackingDirectoryWrapper
	consumer  DocConsumer

	// Updates for our still-in-RAM (to be flushed next) segment
	pendingUpdates *BufferedUpdates
	// Current segment we are working on
	segmentInfo            *SegmentInfo
	aborted                bool
	flushPending           bool
	lastCommittedBytesUsed int64
	hasFlushed             bool
	fieldInfos             *FieldInfosBuilder
	infoStream             io.Writer
	numDocsInRAM           int
	deleteQueue            *DocumentsWriterDeleteQueue
	deleteSlice            *DeleteSlice
	pendingNumDocs         *atomic.Int64
	indexWriterConfig      *LiveIndexWriterConfig
	enableTestPoints       bool
	lock                   sync.RWMutex
	deleteDocIDs           []int
	numDeletedDocIds       int
}

func NewDocumentsWriterPerThread(indexVersionCreated int, segmentName string, directoryOrig, directory store.Directory,
	indexWriterConfig *LiveIndexWriterConfig, deleteQueue *DocumentsWriterDeleteQueue,
	fieldInfos *FieldInfosBuilder, pendingNumDocs *atomic.Int64, enableTestPoints bool) *DocumentsWriterPerThread {

	codec := indexWriterConfig.GetCodec()

	segmentInfo := NewSegmentInfo(directoryOrig, util.VersionLast,
		util.VersionLast, segmentName, -1,
		false, codec, map[string]string{}, []byte(""),
		map[string]string{}, indexWriterConfig.GetIndexSort())

	perThread := &DocumentsWriterPerThread{
		directory:         store.NewTrackingDirectoryWrapper(directory),
		fieldInfos:        fieldInfos,
		indexWriterConfig: indexWriterConfig,
		codec:             codec,
		pendingNumDocs:    pendingNumDocs,
		pendingUpdates:    NewBufferedUpdatesV1(segmentName),
		deleteQueue:       deleteQueue,
		segmentInfo:       segmentInfo,
		enableTestPoints:  enableTestPoints,
	}

	perThread.consumer = indexWriterConfig.getIndexingChain().
		GetChain(indexVersionCreated, segmentInfo, perThread.directory, fieldInfos, indexWriterConfig)
	return perThread
}

// Anything that will add N docs to the index should reserve first to make sure it's allowed.
func (d *DocumentsWriterPerThread) reserveOneDoc() error {
	if d.pendingNumDocs.Inc() > int64(GetActualMaxDocs()) {
		// Reserve failed: put the one doc back and throw exc:
		d.pendingNumDocs.Dec()
		return errors.New("number of documents in the index cannot exceed")
	}
	return nil
}

func (d *DocumentsWriterPerThread) updateDocuments(docs []*document.Document, deleteNode Node) (int64, error) {
	docsInRamBefore := d.numDocsInRAM

	for _, doc := range docs {
		// Even on exception, the document is still added (but marked
		// deleted), so we don't need to un-reserve at that point.
		// Aborting exceptions will actually "lose" more than one
		// document, so the counter will be "wrong" in that case, but
		// it's very hard to fix (we can't easily distinguish aborting
		// vs non-aborting exceptions):
		if err := d.reserveOneDoc(); err != nil {
			return 0, err
		}
		if err := d.consumer.ProcessDocument(d.numDocsInRAM, doc); err != nil {
			return 0, err
		}
	}
	return d.finishDocuments(deleteNode, docsInRamBefore)
}

func (d *DocumentsWriterPerThread) finishDocuments(deleteNode Node, docIdUpTo int) (int64, error) {
	// here we actually finish the document in two steps 1. push the delete into
	// the queue and update our slice. 2. increment the DWPT private document id.
	//
	// the updated slice we get from 1. holds all the deletes that have occurred
	// since we updated the slice the last time.

	// Apply delTerm only after all indexing has
	// succeeded, but apply it only to docs prior to when
	// this batch started:
	var seqNo int64

	if deleteNode != nil {
		seqNo = d.deleteQueue.Add(deleteNode, d.deleteSlice)
		d.deleteSlice.Apply(d.pendingUpdates, docIdUpTo)
		return seqNo, nil
	} else {
		seqNo = d.deleteQueue.UpdateSlice(d.deleteSlice)
		if seqNo < 0 {
			seqNo = -seqNo
			d.deleteSlice.Apply(d.pendingUpdates, docIdUpTo)
		} else {
			d.deleteSlice.Reset()
		}
	}
	return seqNo, nil
}

func (d *DocumentsWriterPerThread) GetNumDocsInRAM() int {
	return d.numDocsInRAM
}

type IndexingChain interface {
	GetChain(indexCreatedVersionMajor int, segmentInfo *SegmentInfo, directory store.Directory,
		fieldInfos *FieldInfosBuilder, indexWriterConfig *LiveIndexWriterConfig) DocConsumer
}

var _ IndexingChain = &defaultIndexingChain{}
var defaultIndexingChainInstance = &defaultIndexingChain{}

type defaultIndexingChain struct {
}

func (*defaultIndexingChain) GetChain(indexCreatedVersionMajor int, segmentInfo *SegmentInfo,
	directory store.Directory, fieldInfos *FieldInfosBuilder, indexWriterConfig *LiveIndexWriterConfig) DocConsumer {
	return NewDefaultIndexingChain(indexCreatedVersionMajor, segmentInfo, directory, fieldInfos, indexWriterConfig)
}
