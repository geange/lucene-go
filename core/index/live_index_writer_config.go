package index

import "github.com/geange/lucene-go/core/analysis"

// LiveIndexWriterConfig Holds all the configuration used by IndexWriter with few setters for settings
// that can be changed on an IndexWriter instance "live".
// Since: 4.0
type LiveIndexWriterConfig struct {
	analyzer analysis.Analyzer

	maxBufferedDocs int
	ramBufferSizeMB int
}
