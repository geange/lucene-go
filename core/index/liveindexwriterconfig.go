package index

import (
	"github.com/geange/lucene-go/core/analysis"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/util/version"
)

type LiveIndexWriterConfig interface {
	GetAnalyzer() analysis.Analyzer

	SetMaxBufferedDocs(maxBufferedDocs int) LiveIndexWriterConfig

	// GetMaxBufferedDocs Returns the number of buffered added documents that will trigger a flush if enabled.
	// See Also: setMaxBufferedDocs(int)
	GetMaxBufferedDocs() int

	// SetMergePolicy
	// Expert: MergePolicy is invoked whenever there are changes to the segments in the index.
	// Its role is to select which merges to do, if any, and return a MergePolicy.MergeSpecification
	// describing the merges. It also selects merges to do for forceMerge.
	// Takes effect on subsequent merge selections. Any merges in flight or any merges already registered by
	// the previous MergePolicy are not affected.
	SetMergePolicy(mergePolicy MergePolicy) LiveIndexWriterConfig

	// SetMergedSegmentWarmer
	// Set the merged segment warmer. See IndexWriter.ReaderWarmer.
	//Takes effect on the next merge.
	SetMergedSegmentWarmer(mergeSegmentWarmer ReaderWarmer) LiveIndexWriterConfig

	// GetMergedSegmentWarmer Returns the current merged segment warmer. See IndexWriter.ReaderWarmer.
	GetMergedSegmentWarmer() ReaderWarmer

	// GetIndexCreatedVersionMajor Return the compatibility version to use for this index.
	// See Also: IndexWriterConfig.setIndexCreatedVersionMajor
	GetIndexCreatedVersionMajor() int

	// GetIndexDeletionPolicy Returns the IndexDeletionPolicy specified in
	// IndexWriterConfig.setIndexDeletionPolicy(IndexDeletionPolicy) or the default KeepOnlyLastCommitDeletionPolicy/
	GetIndexDeletionPolicy() IndexDeletionPolicy

	// GetIndexCommit Returns the IndexCommit as specified in IndexWriterConfig.setIndexCommit(IndexCommit) or the
	// default, null which specifies to open the latest index commit point.
	GetIndexCommit() IndexCommit

	// GetSimilarity Expert: returns the Similarity implementation used by this IndexWriter.
	GetSimilarity() index.Similarity

	// GetMergeScheduler Returns the MergeScheduler that was set by IndexWriterConfig.setMergeScheduler(MergeScheduler).
	GetMergeScheduler() MergeScheduler

	// GetCodec Returns the current Codec.
	GetCodec() index.Codec

	// GetMergePolicy Returns the current MergePolicy in use by this writer.
	// See Also: IndexWriterConfig.setMergePolicy(MergePolicy)
	GetMergePolicy() MergePolicy

	// GetReaderPooling Returns true if IndexWriter should pool readers even if
	// DirectoryReader.open(IndexWriter) has not been called.
	GetReaderPooling() bool

	// GetIndexingChain Returns the indexing chain.
	GetIndexingChain() IndexingChain

	// GetFlushPolicy See Also:
	//IndexWriterConfig.setFlushPolicy(FlushPolicy)
	GetFlushPolicy() FlushPolicy

	// SetUseCompoundFile
	// Sets if the IndexWriter should pack newly written segments in a compound file.
	// Default is true.
	// Use false for batch indexing with very large ram buffer settings.
	// Note: To control compound file usage during segment merges see MergePolicy.setNoCFSRatio(double)
	// and MergePolicy.setMaxCFSSegmentSizeMB(double). This setting only applies to newly created segments.
	SetUseCompoundFile(useCompoundFile bool) LiveIndexWriterConfig

	// GetUseCompoundFile Returns true iff the IndexWriter packs newly written segments in a compound file.
	// Default is true.
	GetUseCompoundFile() bool

	// GetCommitOnClose Returns true if IndexWriter.close() should first commit before closing.
	GetCommitOnClose() bool

	// GetIndexSort Get the index-time Sort order, applied to all (flushed and merged) segments.
	GetIndexSort() index.Sort

	// GetIndexSortFields Returns the field names involved in the index sort
	GetIndexSortFields() map[string]struct{}

	// GetLeafSorter Returns a comparator for sorting leaf readers. If not null, this comparator is
	// used to sort leaf readers within DirectoryReader opened from the IndexWriter of this configuration.
	// Returns: a comparator for sorting leaf readers
	GetLeafSorter() func(a, b index.LeafReader) int

	// IsCheckPendingFlushOnUpdate Expert: Returns if indexing threads check for pending flushes on update
	//in order to help our flushing indexing buffers to disk
	//lucene.experimental
	IsCheckPendingFlushOnUpdate() bool

	// SetCheckPendingFlushUpdate
	// Expert: sets if indexing threads check for pending flushes on update
	// in order to help our flushing indexing buffers to disk. As a consequence, threads calling
	// DirectoryReader.openIfChanged(DirectoryReader, IndexWriter) or IndexWriter.flush() will be the
	// only thread writing segments to disk unless flushes are falling behind. If indexing is stalled due
	// to too many pending flushes indexing threads will help our writing pending segment flushes to disk.
	//lucene.experimental
	SetCheckPendingFlushUpdate(checkPendingFlushOnUpdate bool) LiveIndexWriterConfig

	// GetSoftDeletesField Returns the soft deletes field or null if soft-deletes are disabled.
	// See IndexWriterConfig.setSoftDeletesField(String) for details.
	GetSoftDeletesField() string

	// GetMaxFullFlushMergeWaitMillis Expert: return the amount of time to wait for merges returned by
	// by MergePolicy.findFullFlushMerges(...). If this time is reached, we proceed with the commit
	// based on segments merged up to that point. The merges are not cancelled, and may still run to
	// completion independent of the commit.
	GetMaxFullFlushMergeWaitMillis() int64

	GetOpenMode() OpenMode
}

// liveIndexWriterConfig
// Holds all the configuration used by IndexWriter with few setters for settings
// that can be Changed on an IndexWriter instance "live".
// Since: 4.0
type liveIndexWriterConfig struct {
	analyzer analysis.Analyzer

	maxBufferedDocs int
	ramBufferSizeMB int

	mergedSegmentWarmer ReaderWarmer

	// modified by IndexWriterConfig
	// IndexDeletionPolicy controlling when commit points are deleted.
	delPolicy IndexDeletionPolicy

	// IndexCommit that IndexWriter is opened on.
	commit IndexCommit

	//IndexWriterConfig.OpenMode that IndexWriter is opened with.
	openMode OpenMode

	// Compatibility version to use for this index.
	createdVersionMajor int

	// Similarity to use when encoding norms.
	similarity index.Similarity

	// MergeScheduler to use for running merges.
	mergeScheduler MergeScheduler

	// DocumentsWriterPerThread.IndexingChain that determines how documents are indexed.
	indexingChain IndexingChain

	// Codec used to write new segments.
	codec index.Codec

	// InfoStream for debugging messages.
	// infoStream io.Writer

	// MergePolicy for selecting merges.
	mergePolicy MergePolicy

	// True if readers should be pooled.
	readerPooling bool

	// FlushPolicy to control when segments are flushed.
	flushPolicy FlushPolicy

	// Sets the hard upper bound on RAM usage for a single segment, after which the segment is forced to Flush.
	perThreadHardLimitMB int

	// True if segment flushes should use compound file format
	useCompoundFile bool

	// True if calls to IndexWriter.close() should first do a commit.
	commitOnClose bool

	// The sort order to use to write merged segments.
	indexSort index.Sort

	// The comparator for sorting leaf readers.
	leafSorter func(a, b index.LeafReader) int

	// The field names involved in the index sort
	indexSortFields map[string]struct{}

	// if an indexing thread should check for pending flushes on update in order to help out on a full Flush
	checkPendingFlushOnUpdate bool

	softDeletesField string

	// Amount of time to wait for merges returned by MergePolicy.findFullFlushMerges(...)
	maxFullFlushMergeWaitMillis int64
}

func newLiveIndexWriterConfig(analyzer analysis.Analyzer, codec index.Codec, similarity index.Similarity) *liveIndexWriterConfig {
	return &liveIndexWriterConfig{
		analyzer:                    analyzer,
		maxBufferedDocs:             DEFAULT_MAX_BUFFERED_DOCS,
		ramBufferSizeMB:             DEFAULT_RAM_BUFFER_SIZE_MB,
		mergedSegmentWarmer:         nil,
		delPolicy:                   NewKeepOnlyLastCommitDeletionPolicy(),
		commit:                      nil,
		openMode:                    CREATE_OR_APPEND,
		createdVersionMajor:         int(version.Last.Major()),
		similarity:                  similarity,
		mergeScheduler:              NewNoMergeScheduler(),
		indexingChain:               defaultIndexingChainInstance,
		codec:                       codec,
		mergePolicy:                 NewNoMergePolicy(),
		readerPooling:               DEFAULT_READER_POOLING,
		flushPolicy:                 nil,
		perThreadHardLimitMB:        DEFAULT_RAM_PER_THREAD_HARD_LIMIT_MB,
		useCompoundFile:             DEFAULT_USE_COMPOUND_FILE_SYSTEM,
		commitOnClose:               false,
		indexSort:                   nil,
		leafSorter:                  nil,
		indexSortFields:             nil,
		checkPendingFlushOnUpdate:   false,
		softDeletesField:            "",
		maxFullFlushMergeWaitMillis: DEFAULT_MAX_FULL_FLUSH_MERGE_WAIT_MILLIS,
	}
}

type OpenMode int

const (
	// CREATE Creates a new index or overwrites an existing one.
	CREATE = OpenMode(iota)

	// APPEND Opens an existing index.
	APPEND

	// CREATE_OR_APPEND Creates a new index if one does not exist,
	// otherwise it opens the index and documents will be appended.
	CREATE_OR_APPEND
)

var _ LiveIndexWriterConfig = &liveIndexWriterConfig{}

func (r *liveIndexWriterConfig) GetIndexSort() index.Sort {
	return r.indexSort
}

// GetSimilarity Expert: returns the Similarity implementation used by this IndexWriter.
func (r *liveIndexWriterConfig) GetSimilarity() index.Similarity {
	return r.similarity
}

func (r *liveIndexWriterConfig) GetAnalyzer() analysis.Analyzer {
	return r.analyzer
}

func (r *liveIndexWriterConfig) GetCodec() index.Codec {
	return r.codec
}

func (r *liveIndexWriterConfig) GetIndexingChain() IndexingChain {
	return r.indexingChain
}

func (r *liveIndexWriterConfig) SetMaxBufferedDocs(maxBufferedDocs int) LiveIndexWriterConfig {
	r.maxBufferedDocs = maxBufferedDocs
	return r
}

func (r *liveIndexWriterConfig) GetMaxBufferedDocs() int {
	return r.maxBufferedDocs
}

func (r *liveIndexWriterConfig) SetMergePolicy(mergePolicy MergePolicy) LiveIndexWriterConfig {
	r.mergePolicy = mergePolicy
	return r
}

func (r *liveIndexWriterConfig) SetMergedSegmentWarmer(mergeSegmentWarmer ReaderWarmer) LiveIndexWriterConfig {
	r.mergedSegmentWarmer = mergeSegmentWarmer
	return r
}

func (r *liveIndexWriterConfig) GetMergedSegmentWarmer() ReaderWarmer {
	return r.mergedSegmentWarmer
}

func (r *liveIndexWriterConfig) GetIndexCreatedVersionMajor() int {
	return r.createdVersionMajor
}

func (r *liveIndexWriterConfig) GetIndexDeletionPolicy() IndexDeletionPolicy {
	return r.delPolicy
}

func (r *liveIndexWriterConfig) GetIndexCommit() IndexCommit {
	return r.commit
}

func (r *liveIndexWriterConfig) GetMergeScheduler() MergeScheduler {
	return r.mergeScheduler
}

func (r *liveIndexWriterConfig) GetMergePolicy() MergePolicy {
	return r.mergePolicy
}

func (r *liveIndexWriterConfig) GetReaderPooling() bool {
	return r.readerPooling
}

func (r *liveIndexWriterConfig) GetFlushPolicy() FlushPolicy {
	return r.flushPolicy
}

func (r *liveIndexWriterConfig) SetUseCompoundFile(useCompoundFile bool) LiveIndexWriterConfig {
	r.useCompoundFile = useCompoundFile
	return r
}

func (r *liveIndexWriterConfig) GetUseCompoundFile() bool {
	return r.useCompoundFile
}

func (r *liveIndexWriterConfig) GetCommitOnClose() bool {
	return r.commitOnClose
}

func (r *liveIndexWriterConfig) GetIndexSortFields() map[string]struct{} {
	return r.indexSortFields
}

func (r *liveIndexWriterConfig) GetLeafSorter() func(a, b index.LeafReader) int {
	return r.leafSorter
}

func (r *liveIndexWriterConfig) IsCheckPendingFlushOnUpdate() bool {
	return r.checkPendingFlushOnUpdate
}

func (r *liveIndexWriterConfig) SetCheckPendingFlushUpdate(checkPendingFlushOnUpdate bool) LiveIndexWriterConfig {
	r.checkPendingFlushOnUpdate = checkPendingFlushOnUpdate
	return r
}

func (r *liveIndexWriterConfig) GetSoftDeletesField() string {
	return r.softDeletesField
}

func (r *liveIndexWriterConfig) GetMaxFullFlushMergeWaitMillis() int64 {
	return r.maxFullFlushMergeWaitMillis
}

func (r *liveIndexWriterConfig) GetOpenMode() OpenMode {
	return r.openMode
}
