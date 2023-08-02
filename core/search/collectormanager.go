package search

import (
	"github.com/geange/lucene-go/core/interface/index"
)

// CollectorManager
// A manager of collectors. This class is useful to parallelize execution of search requests and has two main methods:
//   - NewCollector() which must return a NEW collector which will be used to collect a certain set of leaves.
//   - Reduce(Collection) which will be used to reduce the results of individual collections into a meaningful result.
//     This method is only called after all leaves have been fully collected.
//
// See Also: IndexSearcher.search(Query, CollectorManager)
// lucene.experimental
type CollectorManager interface {
	NewCollector() (index.Collector, error)
	Reduce(collectors []index.Collector) (any, error)
}
