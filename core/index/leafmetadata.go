package index

import (
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/util/version"
)

// LeafMetaData Provides read-only metadata about a leaf.
type leafMetaData struct {
	createdVersionMajor int
	minVersion          *version.Version
	sort                index.Sort
}

func NewLeafMetaData(createdVersionMajor int, minVersion *version.Version, sort index.Sort) index.LeafMetaData {
	return &leafMetaData{
		createdVersionMajor: createdVersionMajor,
		minVersion:          minVersion,
		sort:                sort,
	}
}

func (l *leafMetaData) GetSort() index.Sort {
	return l.sort
}
