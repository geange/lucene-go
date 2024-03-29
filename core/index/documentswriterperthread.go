package index

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/version"
)

const (
	INFO_VERBOSE = false
)

type DocumentsWriterPerThread struct {
	lock sync.RWMutex

	codec                  Codec
	directory              store.Directory
	consumer               DocConsumer
	pendingUpdates         *BufferedUpdates // updates for our still-in-RAM (to be flushed next) segment
	segmentInfo            *SegmentInfo     // current segment we are working on
	aborted                bool
	flushPending           bool
	lastCommittedBytesUsed int64
	hasFlushed             bool
	fieldInfos             *FieldInfosBuilder
	numDocsInRAM           int
	deleteQueue            *DocumentsWriterDeleteQueue
	deleteSlice            *DeleteSlice
	pendingNumDocs         *atomic.Int64
	indexWriterConfig      *liveIndexWriterConfig
	enableTestPoints       bool
	deleteDocIDs           []int
	numDeletedDocIds       int
}

func NewDocumentsWriterPerThread(indexVersionCreated int, segmentName string,
	dirOrig, dir store.Directory,
	indexWriterConfig *liveIndexWriterConfig, deleteQueue *DocumentsWriterDeleteQueue,
	fieldInfos *FieldInfosBuilder, pendingNumDocs *atomic.Int64, enableTestPoints bool) *DocumentsWriterPerThread {

	codec := indexWriterConfig.GetCodec()

	segmentInfo := NewSegmentInfo(dirOrig, version.Last,
		version.Last, segmentName, -1,
		false, codec, map[string]string{}, []byte(""),
		map[string]string{}, indexWriterConfig.GetIndexSort())

	perThread := &DocumentsWriterPerThread{
		directory:         dir,
		fieldInfos:        fieldInfos,
		indexWriterConfig: indexWriterConfig,
		codec:             codec,
		pendingNumDocs:    pendingNumDocs,
		pendingUpdates:    NewBufferedUpdates(WithSegmentName(segmentName)),
		deleteQueue:       deleteQueue,
		deleteSlice:       deleteQueue.newSlice(),
		segmentInfo:       segmentInfo,
		enableTestPoints:  enableTestPoints,
	}

	perThread.consumer = indexWriterConfig.GetIndexingChain().
		GetChain(indexVersionCreated, segmentInfo, perThread.directory, fieldInfos, indexWriterConfig)
	return perThread
}

// Anything that will add N docs to the index should reserve first to make sure it's allowed.
// 保证添加的文档的数量在可允许的范围内
func (d *DocumentsWriterPerThread) reserveOneDoc() error {
	if d.pendingNumDocs.Add(1) > int64(GetActualMaxDocs()) {
		// Reserve failed: put the one doc back and throw exc:
		d.pendingNumDocs.Add(-1)
		return errors.New("number of documents in the index cannot exceed")
	}
	return nil
}

func (d *DocumentsWriterPerThread) updateDocuments(ctx context.Context, docs []*document.Document, deleteNode *Node) (int64, error) {
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
		if err := d.consumer.ProcessDocument(ctx, d.numDocsInRAM, doc); err != nil {
			return 0, err
		}
		d.numDocsInRAM++
	}
	return d.finishDocuments(deleteNode, docsInRamBefore)
}

func (d *DocumentsWriterPerThread) finishDocuments(deleteNode *Node, docIdUpTo int) (int64, error) {
	// here we actually finish the document in two steps
	// 1. push the delete into the queue and update our slice.
	// 2. increment the DWPT private document id.
	//
	// the updated slice we get from 1. holds all the deletes that have occurred
	// since we updated the slice the last time.

	// Apply delTerm only after all indexing has succeeded, but apply it only to
	// docs prior to when this batch started:
	if deleteNode != nil {
		seqNo := d.deleteQueue.Add(deleteNode, d.deleteSlice)
		d.deleteSlice.Apply(d.pendingUpdates, docIdUpTo)
		return seqNo, nil
	}

	seqNo := d.deleteQueue.UpdateSlice(d.deleteSlice)
	if seqNo < 0 {
		seqNo = -seqNo
		d.deleteSlice.Apply(d.pendingUpdates, docIdUpTo)
	} else {
		d.deleteSlice.Reset()
	}

	return seqNo, nil
}

func (d *DocumentsWriterPerThread) GetNumDocsInRAM() int {
	return d.numDocsInRAM
}

func (d *DocumentsWriterPerThread) Flush(ctx context.Context) error {
	if err := d.segmentInfo.SetMaxDoc(d.numDocsInRAM); err != nil {
		return err
	}

	flushState := NewSegmentWriteState(d.directory, d.segmentInfo, d.fieldInfos.Finish(), d.pendingUpdates)
	if _, err := d.consumer.Flush(ctx, flushState); err != nil {
		return err
	}
	return nil
}

type IndexingChain interface {
	GetChain(indexCreatedVersionMajor int, segmentInfo *SegmentInfo, directory store.Directory,
		fieldInfos *FieldInfosBuilder, indexWriterConfig *liveIndexWriterConfig) DocConsumer
}

var _ IndexingChain = &defaultIndexingChain{}
var defaultIndexingChainInstance = &defaultIndexingChain{}

type defaultIndexingChain struct {
}

func (*defaultIndexingChain) GetChain(indexCreatedVersionMajor int, segmentInfo *SegmentInfo,
	dir store.Directory, fieldInfos *FieldInfosBuilder, indexWriterConfig *liveIndexWriterConfig) DocConsumer {
	return NewDefaultIndexingChain(indexCreatedVersionMajor, segmentInfo, dir, fieldInfos, indexWriterConfig)
}
