package index

import (
	"github.com/geange/lucene-go/core/analysis"
	"io"
)

// LiveIndexWriterConfig Holds all the configuration used by IndexWriter with few setters for settings
// that can be Changed on an IndexWriter instance "live".
// Since: 4.0
type LiveIndexWriterConfig struct {
	analyzer analysis.Analyzer

	maxBufferedDocs int
	ramBufferSizeMB int

	mergedSegmentWarmer IndexReaderWarmer

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
	similarity Similarity

	// MergeScheduler to use for running merges.
	mergeScheduler MergeScheduler

	// DocumentsWriterPerThread.IndexingChain that determines how documents are indexed.
	indexingChain IndexingChain

	// Codec used to write new segments.
	codec Codec

	// InfoStream for debugging messages.
	infoStream io.Writer

	// MergePolicy for selecting merges.
	mergePolicy *MergePolicy

	// True if readers should be pooled.
	readerPooling bool

	// FlushPolicy to control when segments are flushed.
	flushPolicy *FlushPolicy

	// Sets the hard upper bound on RAM usage for a single segment, after which the segment is forced to flush.
	perThreadHardLimitMB int

	// True if segment flushes should use compound file format
	useCompoundFile bool

	// True if calls to IndexWriter.close() should first do a commit.
	commitOnClose bool

	// The sort order to use to write merged segments.
	indexSort *Sort

	// The comparator for sorting leaf readers.
	leafSorter func(a, b LeafReader) int

	// The field names involved in the index sort
	indexSortFields map[string]struct{}

	// if an indexing thread should check for pending flushes on update in order to help out on a full flush
	checkPendingFlushOnUpdate bool

	softDeletesField string

	// Amount of time to wait for merges returned by MergePolicy.findFullFlushMerges(...)
	maxFullFlushMergeWaitMillis int64
}

func NewLiveIndexWriterConfig(analyzer analysis.Analyzer, codec Codec, similarity Similarity) *LiveIndexWriterConfig {
	return &LiveIndexWriterConfig{
		analyzer:                    analyzer,
		maxBufferedDocs:             DEFAULT_MAX_BUFFERED_DOCS,
		ramBufferSizeMB:             DEFAULT_RAM_BUFFER_SIZE_MB,
		mergedSegmentWarmer:         nil,
		delPolicy:                   nil,
		commit:                      nil,
		openMode:                    CREATE_OR_APPEND,
		createdVersionMajor:         0,
		similarity:                  similarity,
		mergeScheduler:              nil,
		indexingChain:               defaultIndexingChainInstance,
		codec:                       codec,
		infoStream:                  nil,
		mergePolicy:                 nil,
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

	// CREATE_OR_APPEND Creates a new index if one does not exist, otherwise it opens the index and documents will be appended.
	CREATE_OR_APPEND
)

func (r *LiveIndexWriterConfig) GetIndexSort() *Sort {
	return r.indexSort
}

func (r *LiveIndexWriterConfig) GetMergePolicy() *MergePolicy {
	return r.mergePolicy
}

// GetSimilarity Expert: returns the Similarity implementation used by this IndexWriter.
func (r *LiveIndexWriterConfig) GetSimilarity() Similarity {
	return r.similarity
}

func (r *LiveIndexWriterConfig) GetAnalyzer() analysis.Analyzer {
	return r.analyzer
}

func (r *LiveIndexWriterConfig) GetCodec() Codec {
	return r.codec
}

func (r *LiveIndexWriterConfig) getIndexingChain() IndexingChain {
	return r.indexingChain
}
