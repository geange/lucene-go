package search

import (
	context2 "context"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util"
	"io"
	"math"
)

var _ index.Query = &MatchAllDocsQuery{}

type MatchAllDocsQuery struct {
}

func NewMatchAllDocsQuery() *MatchAllDocsQuery {
	return &MatchAllDocsQuery{}
}

func (m *MatchAllDocsQuery) String(field string) string {
	return "*:*"
}

func (m *MatchAllDocsQuery) CreateWeight(searcher index.IndexSearcher, scoreMode index.ScoreMode, boost float64) (index.Weight, error) {
	return newMatchAllDocsQueryWeight(boost, m, scoreMode), nil
}

func (m *MatchAllDocsQuery) Rewrite(reader index.IndexReader) (index.Query, error) {
	return m, nil
}

func (m *MatchAllDocsQuery) Visit(visitor index.QueryVisitor) error {
	return visitor.VisitLeaf(m)
}

var _ index.Weight = &matchAllDocsWeight{}

type matchAllDocsWeight struct {
	*ConstantScoreWeight

	scoreMode index.ScoreMode
}

func newMatchAllDocsQueryWeight(score float64, query index.Query, scoreMode index.ScoreMode) *matchAllDocsWeight {
	weight := &matchAllDocsWeight{
		scoreMode: scoreMode,
	}
	weight.ConstantScoreWeight = NewConstantScoreWeight(score, query, weight)
	return weight
}

func (c *matchAllDocsWeight) Scorer(context index.LeafReaderContext) (index.Scorer, error) {
	maxDoc := context.Reader().MaxDoc()
	return NewConstantScoreScorer(c, c.score, c.scoreMode, types.DocIdSetIteratorAll(maxDoc))
}

func (c *matchAllDocsWeight) IsCacheable(ctx index.LeafReaderContext) bool {
	return true
}

func (c *matchAllDocsWeight) BulkScorer(readerContext index.LeafReaderContext) (index.BulkScorer, error) {
	if c.scoreMode.IsExhaustive() == false {
		return c.ConstantScoreWeight.BulkScorer(readerContext)
	}

	score := c.score
	maxDoc := readerContext.Reader().MaxDoc()

	return &BaseBulkScorer{
		FnScoreRange: func(collector index.LeafCollector, acceptDocs util.Bits, fromDoc, toDoc int) (int, error) {
			toDoc = min(maxDoc, toDoc)
			scorer := NewScoreAndDoc()
			scorer.score = score
			if err := collector.SetScorer(scorer); err != nil {
				return 0, err
			}
			for doc := fromDoc; doc < toDoc; doc++ {
				scorer.doc = doc
				if acceptDocs == nil || acceptDocs.Test(uint(doc)) {
					if err := collector.Collect(context2.Background(), doc); err != nil {
						return 0, err
					}
				}
			}
			if toDoc == maxDoc {
				return math.MaxInt, io.EOF
			}
			return toDoc, nil
		},
		FnCost: func() int64 {
			return int64(maxDoc)
		},
	}, nil
}
