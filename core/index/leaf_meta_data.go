package index

import "github.com/geange/lucene-go/core/util"

// LeafMetaData Provides read-only metadata about a leaf.
type LeafMetaData struct {
	createdVersionMajor int
	minVersion          *util.Version
	sort                *Sort
}

func NewLeafMetaData(createdVersionMajor int, minVersion *util.Version, sort *Sort) *LeafMetaData {
	return &LeafMetaData{
		createdVersionMajor: createdVersionMajor,
		minVersion:          minVersion,
		sort:                sort,
	}
}
