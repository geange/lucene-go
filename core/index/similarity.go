package index

import (
	"github.com/geange/lucene-go/core/types"
)

// Similarity defines the components of Lucene scoring.
// Expert: Scoring API.
// This is a low-level API, you should only extend this API if you want to implement an information retrieval
// model. If you are instead looking for a convenient way to alter Lucene's scoring, consider just tweaking the
// default implementation: BM25Similarity or extend SimilarityBase, which makes it easy to compute a Score
// from index statistics.
// Similarity determines how Lucene weights terms, and Lucene interacts with this class at both index-time
// and query-time.
// Indexing Time At indexing time, the indexer calls computeNorm(FieldInvertState), allowing the Similarity
// implementation to set a per-document value for the field that will be later accessible via
// org.apache.lucene.index.LeafReader.getNormValues(String). Lucene makes no assumption about what is in this
// norm, but it is most useful for encoding length normalization information.
// Implementations should carefully consider how the normalization is encoded: while Lucene's BM25Similarity
// encodes length normalization information with SmallFloat into a single byte, this might not be suitable
// for all purposes.
// Many formulas require the use of average document length, which can be computed via a combination of
// CollectionStatistics.sumTotalTermFreq() and CollectionStatistics.docCount().
// Additional scoring factors can be stored in named NumericDocValuesFields and accessed at query-time with
// org.apache.lucene.index.LeafReader.getNumericDocValues(String). However this should not be done in the
// Similarity but externally, for instance by using FunctionScoreQuery.
// Finally, using index-time boosts (either via folding into the normalization byte or via DocValues), is
// an inefficient way to boost the scores of different fields if the boost will be the same for every document,
// instead the Similarity can simply take a constant boost parameter C, and PerFieldSimilarityWrapper can return
// different instances with different boosts depending upon field name.
// Query time At query-time, Queries interact with the Similarity via these steps:
// The scorer(float, CollectionStatistics, TermStatistics...) method is called a single time, allowing the
// implementation to compute any statistics (such as IDF, average document length, etc) across the entire collection.
// The TermStatistics and CollectionStatistics passed in already contain all of the raw statistics involved,
// so a Similarity can freely use any combination of statistics without causing any additional I/O. Lucene makes
// no assumption about what is stored in the returned Similarity.SimScorer object.
// Then Similarity.SimScorer.Score(float, long) is called for every matching document to compute its Score.
// Explanations When IndexSearcher.explain(org.apache.lucene.search.Query, int) is called, queries consult the
// Similarity's DocScorer for an explanation of how it computed its Score. The query passes in a the document id
// and an explanation of how the frequency was computed.
// See Also:
// org.apache.lucene.index.IndexWriterConfig.setSimilarity(Similarity), IndexSearcher.setSimilarity(Similarity)
type Similarity interface {

	// ComputeNorm Computes the normalization value for a field, given the accumulated state of term processing
	// for this field (see FieldInvertState).
	// Matches in longer fields are less precise, so implementations of this method usually set smaller values
	// when state.getLength() is large, and larger values when state.getLength() is small.
	// Note that for a given term-document frequency, greater unsigned norms must produce scores that are
	// lower or equal, ie. for two encoded norms n1 and n2 so that Long.compareUnsigned(n1, n2) > 0 then
	// SimScorer.Score(freq, n1) <= SimScorer.Score(freq, n2) for any legal freq.
	// 0 is not a legal norm, so 1 is the norm that produces the highest scores.
	// Params: state – current processing state for this field
	// Returns: computed norm value
	ComputeNorm(state *FieldInvertState) int64

	// Scorer Compute any collection-level weight (e.g. IDF, average document length, etc) needed for scoring a query.
	// Params: 	boost – a multiplicative factor to apply to the produces scores
	//			collectionStats – collection-level statistics, such as the number of tokens in the collection.
	//			termStats – term-level statistics, such as the document frequency of a term across the collection.
	// Returns: SimWeight object with the information this Similarity needs to Score a query.
	Scorer(boost float64, collectionStats *types.CollectionStatistics, termStats []types.TermStatistics) SimScorer
}

// SimScorer Stores the weight for a query across the indexed collection. This abstract implementation is empty;
// descendants of Similarity should subclass SimWeight and define the statistics they require in the subclass.
// Examples include idf, average field length, etc.
type SimScorer interface {
	// Score a single document. freq is the document-term sloppy frequency and must be finite and positive.
	// norm is the encoded normalization factor as computed by computeNorm(FieldInvertState) at index time,
	// or 1 if norms are disabled. norm is never 0.
	// Score must not decrease when freq increases, ie. if freq1 > freq2,
	// then Score(freq1, norm) >= Score(freq2, norm) for any value of norm that may be produced by
	// computeNorm(FieldInvertState).
	// Score must not increase when the unsigned norm increases, ie. if Long.compareUnsigned(norm1, norm2) > 0
	// then Score(freq, norm1) <= Score(freq, norm2) for any legal freq.
	// As a consequence, the maximum Score that this scorer can produce is bound by Score(Float.MAX_VALUE, 1).
	// Params: 	freq – sloppy term frequency, must be finite and positive
	// 			norm – encoded normalization factor or 1 if norms are disabled
	// Returns: document's Score
	Score(freq float64, norm int64) float64
}
