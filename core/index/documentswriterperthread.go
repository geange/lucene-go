package index

import (
	"context"
	"errors"
	"github.com/bits-and-blooms/bitset"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util"
	"github.com/google/uuid"
	"golang.org/x/exp/maps"
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
	lock                   sync.RWMutex
	codec                  Codec            // 编码类型
	directory              store.Directory  // 存储目录
	consumer               DocConsumer      // 文档消费者，用于处理添加的Document类型
	pendingUpdates         *BufferedUpdates // updates for our still-in-RAM (to be flushed next) segment
	segmentInfo            *SegmentInfo     // current segment we are working on
	aborted                *atomic.Bool     // 是否中断
	flushPending           *atomic.Bool     // 是否准备flush
	lastCommittedBytesUsed int64
	hasFlushed             *atomic.Bool
	fieldInfos             *FieldInfosBuilder
	numDocsInRAM           *atomic.Int64
	deleteQueue            *DocumentsWriterDeleteQueue
	deleteSlice            *DeleteSlice
	pendingNumDocs         *atomic.Int64
	indexWriterConfig      *liveIndexWriterConfig
	enableTestPoints       bool
	deleteDocIDs           []int
	numDeletedDocIds       int
	filesToDelete          map[string]struct{}
}

func NewDocumentsWriterPerThread(indexVersionCreated int, segmentName string, dirOrig, dir store.Directory,
	indexWriterConfig *liveIndexWriterConfig, deleteQueue *DocumentsWriterDeleteQueue,
	fieldInfos *FieldInfosBuilder, pendingNumDocs *atomic.Int64, enableTestPoints bool) *DocumentsWriterPerThread {

	codec := indexWriterConfig.GetCodec()

	segmentInfo := NewSegmentInfo(dirOrig, version.Last,
		version.Last, segmentName, -1,
		false, codec, map[string]string{}, util.RandomId(),
		map[string]string{}, indexWriterConfig.GetIndexSort())

	consumer := indexWriterConfig.GetIndexingChain().
		GetChain(indexVersionCreated, segmentInfo, dir, fieldInfos, indexWriterConfig)

	return &DocumentsWriterPerThread{
		lock:                   sync.RWMutex{},
		codec:                  codec,
		directory:              store.NewTrackingDirectoryWrapper(dir),
		consumer:               consumer,
		pendingUpdates:         NewBufferedUpdates(WithSegmentName(segmentName)),
		segmentInfo:            segmentInfo,
		aborted:                new(atomic.Bool),
		flushPending:           new(atomic.Bool),
		lastCommittedBytesUsed: 0,
		hasFlushed:             new(atomic.Bool),
		fieldInfos:             fieldInfos,
		numDocsInRAM:           new(atomic.Int64),
		deleteQueue:            deleteQueue,
		deleteSlice:            deleteQueue.newSlice(),
		pendingNumDocs:         new(atomic.Int64),
		indexWriterConfig:      indexWriterConfig,
		enableTestPoints:       false,
		deleteDocIDs:           make([]int, 0),
		numDeletedDocIds:       0,
		filesToDelete:          make(map[string]struct{}),
	}
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
	docsInRamBefore := int(d.numDocsInRAM.Load())

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

		if err := d.consumer.ProcessDocument(ctx, int(d.numDocsInRAM.Load()), doc); err != nil {
			return 0, err
		}
		d.numDocsInRAM.Add(1)
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
		if err := d.deleteSlice.Apply(d.pendingUpdates, docIdUpTo); err != nil {
			return 0, err
		}
		return seqNo, nil
	}

	seqNo := d.deleteQueue.UpdateSlice(d.deleteSlice)
	if seqNo < 0 {
		seqNo = -seqNo
		if err := d.deleteSlice.Apply(d.pendingUpdates, docIdUpTo); err != nil {
			return 0, err
		}
	} else {
		d.deleteSlice.Reset()
	}

	return seqNo, nil
}

// This method marks the last N docs as deleted. This is used
// in the case of a non-aborting exception. There are several cases
// where we fail a document ie. due to an exception during analysis
// that causes the doc to be rejected but won't cause the DWPT to be
// stale nor the entire IW to abort and shutdown. In such a case
// we only mark these docs as deleted and turn it into a livedocs
// during flush
// TODO
func (d *DocumentsWriterPerThread) deleteLastDocs(docCount int) error {
	panic("implement me")
}

func (d *DocumentsWriterPerThread) GetNumDocsInRAM() int {
	return int(d.numDocsInRAM.Load())
}

func (d *DocumentsWriterPerThread) Flush(ctx context.Context) error {
	if err := d.segmentInfo.SetMaxDoc(int(d.numDocsInRAM.Load())); err != nil {
		return err
	}

	flushState := NewSegmentWriteState(d.directory, d.segmentInfo, d.fieldInfos.Finish(), d.pendingUpdates, nil)
	if _, err := d.consumer.Flush(ctx, flushState); err != nil {
		return err
	}

	return d.directory.Close()
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

type FlushedSegment struct {
	segmentInfo    *SegmentCommitInfo
	fieldInfos     *FieldInfos
	segmentUpdates *FrozenBufferedUpdates
	liveDocs       *bitset.BitSet
	sortMap        *DocMap
	delCount       int
}

func newFlushedSegment(segmentInfo *SegmentCommitInfo, fieldInfos *FieldInfos,
	segmentUpdates *BufferedUpdates, liveDocs *bitset.BitSet, delCount int, sortMap *DocMap) *FlushedSegment {

	segment := &FlushedSegment{
		segmentInfo:    segmentInfo,
		fieldInfos:     fieldInfos,
		segmentUpdates: nil,
		liveDocs:       liveDocs,
		delCount:       delCount,
		sortMap:        sortMap,
	}

	if segmentUpdates != nil && segmentUpdates.Any() {
		segment.segmentUpdates = NewFrozenBufferedUpdates(segmentUpdates, segmentInfo)
	}

	return segment
}

// Flush all pending docs to a new segment
func (d *DocumentsWriterPerThread) flush(ctx context.Context, flushNotifications FlushNotifications) (*FlushedSegment, error) {
	if err := d.segmentInfo.SetMaxDoc(int(d.numDocsInRAM.Load())); err != nil {
		return nil, err
	}

	flushInfo := store.NewFlushInfo(int(d.numDocsInRAM.Load()), d.lastCommittedBytesUsed)
	ioContext := store.NewIOContext(store.WithFlushInfo(flushInfo))

	flushState := NewSegmentWriteState(d.directory, d.segmentInfo, d.fieldInfos.Finish(),
		d.pendingUpdates, ioContext)

	// Apply delete-by-docID now (delete-byDocID only
	// happens when an exception is hit processing that
	// doc, eg if analyzer has some problem w/ the text):
	if d.numDeletedDocIds > 0 {
		numDocsInRAM := uint(d.numDocsInRAM.Load())
		flushState.LiveDocs = bitset.New(numDocsInRAM)
		flushState.LiveDocs.FlipRange(0, numDocsInRAM)
		for i := 0; i < d.numDeletedDocIds; i++ {
			flushState.LiveDocs.Clear(uint(d.deleteDocIDs[i]))
		}

		flushState.DelCountOnFlush = d.numDeletedDocIds
		d.deleteDocIDs = make([]int, 0)
	}

	if d.aborted.Load() {
		return nil, nil
	}

	var sortMap *DocMap

	var softDeletedDocs types.DocIdSetIterator
	if d.indexWriterConfig.GetSoftDeletesField() != "" {
		softDeletedDocs = d.consumer.GetHasDocValues(d.indexWriterConfig.GetSoftDeletesField())
	} else {
		softDeletedDocs = nil
	}
	sortMap, err := d.consumer.Flush(ctx, flushState)
	if err != nil {
		return nil, err
	}

	if softDeletedDocs == nil {
		flushState.SoftDelCountOnFlush = 0
	} else {
		softDeletes, err := countSoftDeletes(softDeletedDocs, flushState.LiveDocs)
		if err != nil {
			return nil, err
		}
		flushState.SoftDelCountOnFlush = softDeletes
	}
	// We clear this here because we already resolved them (private to this segment) when writing postings:
	d.pendingUpdates.ClearDeleteTerms()
	d.segmentInfo.SetFiles(d.directory.(*store.TrackingDirectoryWrapper).GetCreatedFiles())

	segmentInfoPerCommit := NewSegmentCommitInfo(d.segmentInfo, 0, flushState.SoftDelCountOnFlush, -1, -1, -1, []byte(uuid.New().String()))

	var segmentDeletes *BufferedUpdates
	if len(d.pendingUpdates.deleteQueries) == 0 && d.pendingUpdates.numFieldUpdates.Load() == 0 {
		d.pendingUpdates.Clear()
		segmentDeletes = nil
	} else {
		segmentDeletes = d.pendingUpdates
	}

	fs := newFlushedSegment(segmentInfoPerCommit, flushState.FieldInfos,
		segmentDeletes, flushState.LiveDocs, flushState.DelCountOnFlush, sortMap)
	if err := d.sealFlushedSegment(ctx, fs, sortMap, flushNotifications); err != nil {
		return nil, err
	}

	return fs, nil
}

// Seals the SegmentInfo for the new flushed segment and persists the deleted documents FixedBitSet.
func (d *DocumentsWriterPerThread) sealFlushedSegment(ctx context.Context, flushedSegment *FlushedSegment, sortMap *DocMap, flushNotifications FlushNotifications) error {
	newSegment := flushedSegment.segmentInfo

	if err := SetDiagnostics(newSegment.info, SOURCE_FLUSH, nil); err != nil {
		return err
	}

	maxDoc, err := newSegment.info.MaxDoc()
	if err != nil {
		return err
	}
	sizeInBytes, err := newSegment.SizeInBytes()
	if err != nil {
		return err
	}
	ioContext := store.NewIOContext(store.WithFlushInfo(store.NewFlushInfo(maxDoc, sizeInBytes)))

	if d.indexWriterConfig.GetUseCompoundFile() {
		originalFiles := newSegment.info.Files()
		// TODO: like addIndexes, we are relying on createCompoundFile to successfully cleanup...
		if err := CreateCompoundFile(ctx, store.NewTrackingDirectoryWrapper(d.directory),
			newSegment.info, ioContext, flushNotifications.DeleteUnusedFiles); err != nil {
			return err
		}
		maps.Copy(d.filesToDelete, originalFiles)
		newSegment.info.SetUseCompoundFile(true)
	}

	// Have codec write SegmentInfo.  Must do this after
	// creating CFS so that 1) .si isn't slurped into CFS,
	// and 2) .si reflects useCompoundFile=true change
	// above:
	if err := d.codec.SegmentInfoFormat().Write(ctx, d.directory, newSegment.info, ioContext); err != nil {
		return err
	}

	// TODO: ideally we would freeze newSegment here!!
	// because any changes after writing the .si will be
	// lost...

	// Must write deleted docs after the CFS so we don't
	// slurp the del file into CFS:
	if flushedSegment.liveDocs != nil {
		delCount := flushedSegment.delCount

		// TODO: we should prune the segment if it's 100%
		// deleted... but merge will also catch it.

		// TODO: in the NRT case it'd be better to hand
		// this del vector over to the
		// shortly-to-be-opened SegmentReader and let it
		// carry the changes; there's no reason to use
		// filesystem as intermediary here.

		info := flushedSegment.segmentInfo
		codec := info.info.GetCodec()
		var bits *bitset.BitSet
		if sortMap == nil {
			bits = flushedSegment.liveDocs
		} else {
			bits = sortLiveDocs(flushedSegment.liveDocs, sortMap)
		}
		if err := codec.LiveDocsFormat().WriteLiveDocs(ctx, bits, d.directory, info, delCount, ioContext); err != nil {
			return err
		}
		newSegment.SetDelCount(delCount)
		newSegment.AdvanceDelGen()
	}

	return nil
}

func (d *DocumentsWriterPerThread) PendingFilesToDelete() map[string]struct{} {
	return d.filesToDelete
}

func (d *DocumentsWriterPerThread) GetSegmentInfo() *SegmentInfo {
	return d.segmentInfo
}

func (d *DocumentsWriterPerThread) prepareFlush() (*FrozenBufferedUpdates, error) {
	globalUpdates := d.deleteQueue.freezeGlobalBuffer(d.deleteSlice)
	// deleteSlice can possibly be null if we have hit non-aborting exceptions during indexing and
	// never succeeded adding a document.
	if d.deleteSlice != nil {
		// apply all deletes before we flush and release the delete slice
		if err := d.deleteSlice.Apply(d.pendingUpdates, int(d.numDocsInRAM.Load())); err != nil {
			return nil, err
		}
		d.deleteSlice.Reset()
	}

	return globalUpdates, nil
}

// TODO: fix it
// Called if we hit an exception at a bad time (when updating the index files) and must discard all currently
// buffered docs. This resets our state, discarding any docs added since last flush.
func (d *DocumentsWriterPerThread) abort() error {
	d.aborted.Store(true)
	d.pendingNumDocs.Add(-d.numDocsInRAM.Load())

	err := d.consumer.Abort()
	if err != nil {
		return err
	}
	d.pendingUpdates.Clear()

	return nil
}

func sortLiveDocs(liveDocs *bitset.BitSet, sortMap *DocMap) *bitset.BitSet {
	sortedLiveDocs := bitset.New(liveDocs.Len())
	sortedLiveDocs.FlipRange(0, liveDocs.Len())

	size := liveDocs.Len()
	for i := uint(0); i < size; i++ {
		if liveDocs.Test(i) == false {
			idx := int(i)
			newIdx := uint(sortMap.OldToNew(idx))
			sortedLiveDocs.Clear(newIdx)
		}
	}
	return sortedLiveDocs
}

/**
  private FixedBitSet sortLiveDocs(Bits liveDocs, Sorter.DocMap sortMap) {
    assert liveDocs != null && sortMap != null;
    FixedBitSet sortedLiveDocs = new FixedBitSet(liveDocs.length());
    sortedLiveDocs.set(0, liveDocs.length());
    for (int i = 0; i < liveDocs.length(); i++) {
      if (liveDocs.get(i) == false) {
        sortedLiveDocs.clear(sortMap.oldToNew(i));
      }
    }
    return sortedLiveDocs;
  }
*/
