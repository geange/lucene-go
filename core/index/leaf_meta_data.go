package index

import "github.com/geange/lucene-go/core/util"

// LeafMetaData Provides read-only metadata about a leaf.
type LeafMetaData struct {
	createdVersionMajor int
	minVersion          *util.Version
	// Sort sort
}
