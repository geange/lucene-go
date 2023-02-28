package index

import (
	"fmt"
	"github.com/geange/lucene-go/core/analysis"
	"github.com/geange/lucene-go/core/analysis/standard"
	"sync"
)

type IndexWriterConfig struct {
	*LiveIndexWriterConfig

	sync.Once

	// indicates whether this config instance is already attached to a writer.
	// not final so that it can be cloned properly.
	writer *IndexWriter
}

func NewIndexWriterConfig(codec Codec, similarity Similarity) *IndexWriterConfig {
	cfg := &IndexWriterConfig{}
	analyzer := standard.NewStandardAnalyzer(analysis.EMPTY_SET)
	cfg.LiveIndexWriterConfig = NewLiveIndexWriterConfig(analyzer, codec, similarity)
	return cfg
}

func (c *IndexWriterConfig) setIndexWriter(writer *IndexWriter) {
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

func (c *IndexWriterConfig) getIndexCreatedVersionMajor() int {
	return c.createdVersionMajor
}

const (

	// DISABLE_AUTO_FLUSH Denotes a flush trigger is disabled.
	DISABLE_AUTO_FLUSH = -1

	// DEFAULT_MAX_BUFFERED_DELETE_TERMS Disabled by default (because IndexWriter flushes by RAM usage by default).
	DEFAULT_MAX_BUFFERED_DELETE_TERMS = DISABLE_AUTO_FLUSH

	// DEFAULT_MAX_BUFFERED_DOCS Disabled by default (because IndexWriter flushes by RAM usage by default).
	DEFAULT_MAX_BUFFERED_DOCS = DISABLE_AUTO_FLUSH

	// DEFAULT_RAM_BUFFER_SIZE_MB Default value is 16 MB (which means flush when buffered docs consume
	// approximately 16 MB RAM).
	DEFAULT_RAM_BUFFER_SIZE_MB = 16.0

	// DEFAULT_READER_POOLING Default setting (true) for setReaderPooling.
	// We Changed this default to true with concurrent deletes/updates (LUCENE-7868),
	// because we will otherwise need to open and close segment readers more frequently.
	// False is still supported, but will have worse performance since readers will
	// be forced to aggressively move all state to disk.
	DEFAULT_READER_POOLING = true

	// DEFAULT_RAM_PER_THREAD_HARD_LIMIT_MB Default value is 1945. Change using setRAMPerThreadHardLimitMB(int)
	DEFAULT_RAM_PER_THREAD_HARD_LIMIT_MB = 1945

	// DEFAULT_USE_COMPOUND_FILE_SYSTEM Default value for compound file system for newly
	// written segments (set to true). For batch indexing with very large ram buffers use false
	DEFAULT_USE_COMPOUND_FILE_SYSTEM = true

	// DEFAULT_COMMIT_ON_CLOSE Default value for whether calls to IndexWriter.close() include a commit.
	DEFAULT_COMMIT_ON_CLOSE = true

	// DEFAULT_MAX_FULL_FLUSH_MERGE_WAIT_MILLIS Default value for time to wait for merges
	// on commit or getReader (when using a MergePolicy that implements MergePolicy.findFullFlushMerges).
	DEFAULT_MAX_FULL_FLUSH_MERGE_WAIT_MILLIS = 0
)
