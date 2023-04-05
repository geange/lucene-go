package search

type CollectorManager struct {

	// NewCollector
	// Return a new Collector. This must return a different instance on each call.
	NewCollector func() (Collector, error)

	// Reduce the results of individual collectors into a meaningful result.
	// For instance a TopDocsCollector would compute the top docs of each collector
	// and then merge them using TopDocs.merge(int, TopDocs[]). This method must be
	// called after collection is finished on all provided collectors.
	Reduce func(collectors []Collector) (any, error)
}
