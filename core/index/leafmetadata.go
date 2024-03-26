package index

import (
	"github.com/geange/lucene-go/core/util/version"
)

// LeafMetaData Provides read-only metadata about a leaf.
type LeafMetaData struct {
	createdVersionMajor int
	minVersion          *version.Version
	sort                *Sort
}

func NewLeafMetaData(createdVersionMajor int, minVersion *version.Version, sort *Sort) *LeafMetaData {
	return &LeafMetaData{
		createdVersionMajor: createdVersionMajor,
		minVersion:          minVersion,
		sort:                sort,
	}
}

func (l *LeafMetaData) GetSort() *Sort {
	return l.sort
}
