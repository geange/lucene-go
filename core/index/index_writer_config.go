package index

import "sync"

type IndexWriterConfig struct {
	LiveIndexWriterConfig

	sync.Once

	// indicates whether this config instance is already attached to a writer.
	// not final so that it can be cloned properly.
	writer *IndexWriter
}
