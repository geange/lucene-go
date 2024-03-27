package search

// TotalHits
// Description of the total number of hits of a query.
// The total hit count can't generally be computed accurately without visiting all matches,
// which is costly for queries that match lots of documents. Given that it is often enough
// to have a lower bounds of the number of hits, such as "there are more than 1000 hits",
// Lucene has options to stop counting as soon as a threshold has been reached in order to
// improve query times.
type TotalHits struct {
	Value    int64
	Relation TotalHitsRelation
}

func NewTotalHits(value int64, relation TotalHitsRelation) *TotalHits {
	return &TotalHits{Value: value, Relation: relation}
}

// TotalHitsRelation
// How the value should be interpreted.
type TotalHitsRelation int

const (
	EQUAL_TO                 = TotalHitsRelation(iota) // The total hit count is equal to value.
	GREATER_THAN_OR_EQUAL_TO                           // The total hit count is greater than or equal to value.
)
