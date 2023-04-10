package index

// Scorable Allows access to the score of a Query
type Scorable interface {
	// Score Returns the score of the current document matching the query.
	Score() (float32, error)

	// SmoothingScore Returns the smoothing score of the current document matching the query. This score is used when the query/term does not appear in the document, and behaves like an idf. The smoothing score is particularly important when the Scorer returns a product of probabilities so that the document score does not go to zero when one probability is zero. This can return 0 or a smoothing score.
	// Smoothing scores are described in many papers, including: Metzler, D. and Croft, W. B. , "Combining the Language Model and Inference Network Approaches to Retrieval," Information Processing and Management Special Issue on Bayesian Networks and Information Retrieval, 40(5), pp.735-750.
	SmoothingScore(docId int) (float32, error)
}
