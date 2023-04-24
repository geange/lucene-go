package search

type CollectorManager interface {
	NewCollector() (Collector, error)
	Reduce(collectors []TopScoreDocCollector) (any, error)
}

var _ CollectorManager = &CollectorManagerDefault{}

type CollectorManagerDefault struct {

	// NewCollector
	// Return a new Collector. This must return a different instance on each call.
	FnNewCollector func() (Collector, error)

	// Reduce the results of individual collectors into a meaningful result.
	// For instance a TopDocsCollector would compute the top docs of each collector
	// and then merge them using TopDocs.merge(int, TopDocs[]). This method must be
	// called after collection is finished on all provided collectors.
	FnReduce func(collectors []TopScoreDocCollector) (any, error)
}

func (c *CollectorManagerDefault) NewCollector() (Collector, error) {
	return c.FnNewCollector()
}

func (c *CollectorManagerDefault) Reduce(collectors []TopScoreDocCollector) (any, error) {
	return c.FnReduce(collectors)
}
