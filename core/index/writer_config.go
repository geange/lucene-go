package index

import (
	"fmt"
	"github.com/geange/lucene-go/core/analysis"
	"github.com/geange/lucene-go/core/analysis/standard"
	"sync"
)

type IndexWriterConfig struct {
	*liveIndexWriterConfig

	sync.Once

	// indicates whether this config instance is already attached to a writer.
	// not final so that it can be cloned properly.
	writer *Writer

	flushPolicy FlushPolicy
}

func NewIndexWriterConfig(codec Codec, similarity Similarity) *IndexWriterConfig {
	cfg := &IndexWriterConfig{}
	analyzer := standard.NewStandardAnalyzer(analysis.EMPTY_SET)
	cfg.liveIndexWriterConfig = newLiveIndexWriterConfig(analyzer, codec, similarity)
	return cfg
}

func (c *IndexWriterConfig) setIndexWriter(writer *Writer) {
	c.writer = writer
}

func (c *IndexWriterConfig) getSoftDeletesField() string {
	return c.softDeletesField
}

// SetIndexSort Set the Sort order to use for all (flushed and merged) segments.
func (c *IndexWriterConfig) SetIndexSort(sort *Sort) error {
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

	// DISABLE_AUTO_FLUSH Denotes a Flush trigger is disabled.
	DISABLE_AUTO_FLUSH = -1

	// DEFAULT_MAX_BUFFERED_DELETE_TERMS Disabled by default (because IndexWriter flushes by RAM usage by default).
	DEFAULT_MAX_BUFFERED_DELETE_TERMS = DISABLE_AUTO_FLUSH

	// DEFAULT_MAX_BUFFERED_DOCS Disabled by default (because IndexWriter flushes by RAM usage by default).
	DEFAULT_MAX_BUFFERED_DOCS = DISABLE_AUTO_FLUSH

	// DEFAULT_RAM_BUFFER_SIZE_MB Default item is 16 MB (which means Flush when buffered docs consume
	// approximately 16 MB RAM).
	DEFAULT_RAM_BUFFER_SIZE_MB = 16.0

	// DEFAULT_READER_POOLING Default setting (true) for setReaderPooling.
	// We Changed this default to true with concurrent deletes/updates (LUCENE-7868),
	// because we will otherwise need to open and close segment readers more frequently.
	// False is still supported, but will have worse performance since readers will
	// be forced to aggressively move all state to disk.
	DEFAULT_READER_POOLING = true

	// DEFAULT_RAM_PER_THREAD_HARD_LIMIT_MB Default item is 1945. Change using setRAMPerThreadHardLimitMB(int)
	DEFAULT_RAM_PER_THREAD_HARD_LIMIT_MB = 1945

	// DEFAULT_USE_COMPOUND_FILE_SYSTEM Default item for compound file system for newly
	// written segments (set to true). For batch indexing with very large ram buffers use false
	DEFAULT_USE_COMPOUND_FILE_SYSTEM = true

	// DEFAULT_COMMIT_ON_CLOSE Default item for whether calls to IndexWriter.close() include a commit.
	DEFAULT_COMMIT_ON_CLOSE = true

	// DEFAULT_MAX_FULL_FLUSH_MERGE_WAIT_MILLIS Default item for time to wait for merges
	// on commit or getReader (when using a MergePolicy that implements MergePolicy.findFullFlushMerges).
	DEFAULT_MAX_FULL_FLUSH_MERGE_WAIT_MILLIS = 0
)
