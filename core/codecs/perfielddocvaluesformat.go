package codecs

import "github.com/geange/lucene-go/core/interface/index"

type PerFieldDocValuesFormat interface {
	index.DocValuesFormat
}
