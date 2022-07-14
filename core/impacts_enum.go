package core

import "github.com/geange/lucene-go/core/index"

// ImpactsEnum Extension of PostingsEnum which also provides information about upcoming impacts.
type ImpactsEnum interface {
	index.PostingsEnum
}
