package search

import (
	"fmt"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util"
	"math"
)

var (
	LENGTH_TABLE [256]float64
)

func init() {
	for i := 0; i < 256; i++ {
		LENGTH_TABLE[i] = float64(util.Byte4ToInt(byte(i)))
	}
}

var _ index.Similarity = &BM25Similarity{}

// BM25Similarity
// BM25 Similarity. Introduced in Stephen E. Robertson, Steve Walker, Susan Jones,
// Micheline Hancock-Beaulieu, and Mike Gatford. Okapi at TREC-3. In Proceedings of the Third Text REtrieval
// Conference (TREC 1994). Gaithersburg, USA, November 1994.
type BM25Similarity struct {
	k1 float64
	b  float64

	// True if overlap tokens (tokens with a position of increment of zero) are discounted from the document's length.
	discountOverlaps bool
}

// NewBM25Similarity
// BM25 with these default values:
// * k1 = 1.2
// * b = 0.75
func NewBM25Similarity() (*BM25Similarity, error) {
	return NewBM25SimilarityV1(1.2, 0.75)
}

func NewCastBM25Similarity() (*BM25Similarity, error) {
	similarity, err := NewBM25SimilarityV1(1.2, 0.75)
	if err != nil {
		return nil, err
	}
	return similarity, nil
}

// NewBM25SimilarityV1
// BM25 with the supplied parameter values.
// k1: Controls non-linear term frequency normalization (saturation).
// b: Controls to what degree document length normalizes tf values.
// Throws: 	IllegalArgumentException – if k1 is infinite or negative, or if b is not within the range [0..1]
func NewBM25SimilarityV1(k1, b float64) (*BM25Similarity, error) {
	if k1 < 0 {
		return nil, fmt.Errorf("illegal k1 value: %f, must be a non-negative finite value", k1)
	}

	if b < 0 || b > 1 {
		return nil, fmt.Errorf("illegal b value: %f, must be between 0 and 1", b)
	}

	similarity := &BM25Similarity{k1: k1, b: b, discountOverlaps: true}
	return similarity, nil
}

// SetDiscountOverlaps
// Sets whether overlap tokens (Tokens with 0 position increment) are ignored when
// computing norm. By default this is true, meaning overlap tokens do not count when computing norms.
func (b *BM25Similarity) SetDiscountOverlaps(v bool) {
	b.discountOverlaps = v
}

// GetDiscountOverlaps
// Returns true if overlap tokens are discounted from the document's length.
// See Also: setDiscountOverlaps
func (b *BM25Similarity) GetDiscountOverlaps() bool {
	return b.discountOverlaps
}

func (b *BM25Similarity) ComputeNorm(state *index.FieldInvertState) int64 {
	numTerms := 0
	if state.GetIndexOptions() == document.INDEX_OPTIONS_DOCS && state.GetIndexCreatedVersionMajor() >= 8 {
		numTerms = state.GetUniqueTermCount()
	} else if b.discountOverlaps {
		numTerms = state.GetLength() - state.GetNumOverlap()
	} else {
		numTerms = state.GetLength()
	}
	return int64(numTerms)
}

// IdfExplain
// Computes a score factor for a simple term and returns an explanation for that score factor.
// The default implementation uses:
//
//	idf(docFreq, docCount);
//
// Note that CollectionStatistics.docCount() is used instead of Reader#numDocs() because
// also TermStatistics.docFreq() is used, and when the latter is inaccurate, so is
// CollectionStatistics.docCount(), and in the same direction. In addition, CollectionStatistics.docCount()
// does not skew when fields are sparse.
// Params:  collectionStats – collection-level statistics
//
//	termStats – term-level statistics for the term
//
// Returns: an Explain object that includes both an idf score factor and an explanation for the term.
func (b *BM25Similarity) IdfExplain(
	collectionStats *types.CollectionStatistics, termStats *types.TermStatistics) *types.Explanation {

	df := termStats.DocFreq()
	docCount := collectionStats.DocCount()
	idfValue := idf(df, docCount)

	exp1 := types.NewExplanation(true, df, "n, number of documents containing term")

	exp2 := types.NewExplanation(true, docCount, "N, total number of documents with field")

	return types.NewExplanation(true, idfValue,
		"idf, computed as log(1 + (N - n + 0.5) / (n + 0.5)) from:",
		exp1, exp2)
}

// IdfExplainV1
// Computes a score factor for a phrase.
// The default implementation sums the idf factor for each term in the phrase.
// collectionStats: collection-level statistics
// termStats: term-level statistics for the terms in the phrase
// Returns: an Explain object that includes both an idf score factor for the phrase and an explanation for each term.
func (b *BM25Similarity) IdfExplainV1(
	collectionStats *types.CollectionStatistics, termStats []types.TermStatistics) *types.Explanation {

	idfValue := 0.0
	details := make([]*types.Explanation, 0)
	for _, stat := range termStats {
		idfExplain := b.IdfExplain(collectionStats, &stat)
		details = append(details, idfExplain)
		v, ok := idfExplain.GetValue().(float64)
		if ok {
			idfValue += v
		}
	}
	return types.NewExplanation(true, idfValue, "idf, sum of:", details...)
}

func (b *BM25Similarity) Scorer(boost float64,
	collectionStats *types.CollectionStatistics, termStats []types.TermStatistics) index.SimScorer {

	var idfValue *types.Explanation
	if len(termStats) == 1 {
		idfValue = b.IdfExplain(collectionStats, &termStats[0])
	} else {
		idfValue = b.IdfExplainV1(collectionStats, termStats)
	}

	avgdl := avgFieldLength(collectionStats)

	cache := make([]float64, 256)
	for i := range cache {
		cache[i] = 1.0 / (b.k1 * ((1 - b.b) + b.b*LENGTH_TABLE[i]/avgdl))
	}

	return NewBM25Scorer(boost, b.k1, b.b, idfValue, avgdl, cache)
}

func (b *BM25Similarity) String() string {
	return fmt.Sprintf("BM25(k1=%f,b=%f)", b.k1, b.b)
}

func (b *BM25Similarity) GetK1() float64 {
	return b.k1
}

func (b *BM25Similarity) GetB() float64 {
	return b.b
}

var _ index.SimScorer = &BM25Scorer{}

type BM25Scorer struct {
	*index.BaseSimScorer

	boost  float64            // query boost
	k1     float64            // k1 value for scale factor
	b      float64            // b value for length normalization impact
	idf    *types.Explanation // BM25's idf
	avgdl  float64            //The average document length.
	cache  []float64          // precomputed norm[256] with k1 * ((1 - b) + b * dl / avgdl)
	weight float64            // weight (idf * boost)
}

func NewBM25Scorer(boost, k1, b float64, idf *types.Explanation, avgdl float64, cache []float64) *BM25Scorer {
	scorer := &BM25Scorer{
		boost:  boost,
		k1:     k1,
		b:      b,
		idf:    idf,
		avgdl:  avgdl,
		cache:  cache,
		weight: boost * idf.GetValue().(float64),
	}
	scorer.BaseSimScorer = index.NewBaseSimScorer(scorer)
	return scorer
}

func (b *BM25Scorer) Score(freq float64, norm int64) float64 {
	// In order to guarantee monotonicity with both freq and norm without
	// promoting to doubles, we rewrite freq / (freq + norm) to
	// 1 - 1 / (1 + freq * 1/norm).
	// freq * 1/norm is guaranteed to be monotonic for both freq and norm due
	// to the fact that multiplication and division round to the nearest
	// float. And then monotonicity is preserved through composition via
	// x -> 1 + x and x -> 1 - 1/x.
	// Finally we expand weight * (1 - 1 / (1 + freq * 1/norm)) to
	// weight - weight / (1 + freq * 1/norm), which runs slightly faster.
	normInverse := b.cache[(byte(norm))&0xFF]
	return b.weight - b.weight/(1.0+freq*normInverse)
}

// Implemented as log(1 + (docCount - docFreq + 0.5)/(docFreq + 0.5)).
func idf(docFreq, docCount int64) float64 {
	return math.Log(1 + (float64(docCount-docFreq)+0.5)/(float64(docFreq)+0.5))
}

// The default implementation computes the average as sumTotalTermFreq / docCount
func avgFieldLength(collectionStats *types.CollectionStatistics) float64 {
	return float64(collectionStats.SumTotalTermFreq()) / float64(collectionStats.DocCount())
}
