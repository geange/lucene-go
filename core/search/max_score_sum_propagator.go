package search

// MaxScoreSumPropagator
// Utility class to propagate scoring information in BooleanQuery, which compute the score as the sum of the scores of its matching clauses. This helps propagate information about the maximum produced score
type MaxScoreSumPropagator struct {
}
