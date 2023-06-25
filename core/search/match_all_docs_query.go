package search

import (
	context2 "context"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/util"
	"io"
	"math"
)

var _ Query = &MatchAllDocsQuery{}

type MatchAllDocsQuery struct {
}

func NewMatchAllDocsQuery() *MatchAllDocsQuery {
	return &MatchAllDocsQuery{}
}

func (m *MatchAllDocsQuery) String(field string) string {
	return "*:*"
}

func (m *MatchAllDocsQuery) CreateWeight(_ *IndexSearcher, scoreMode *ScoreMode, boost float64) (Weight, error) {
	return newConstantScoreWeight(boost, m, scoreMode), nil
}

func (m *MatchAllDocsQuery) Rewrite(reader index.IndexReader) (Query, error) {
	return m, nil
}

func (m *MatchAllDocsQuery) Visit(visitor QueryVisitor) error {
	return visitor.VisitLeaf(m)
}

var _ Weight = &constantScoreWeight{}

type constantScoreWeight struct {
	*ConstantScoreWeight

	scoreMode *ScoreMode
}

func newConstantScoreWeight(score float64, query Query, scoreMode *ScoreMode) *constantScoreWeight {
	weight := &constantScoreWeight{
		scoreMode: scoreMode,
	}
	weight.ConstantScoreWeight = NewConstantScoreWeight(score, query, weight)
	return weight
}

func (c *constantScoreWeight) Scorer(context *index.LeafReaderContext) (Scorer, error) {
	maxDoc := context.Reader().MaxDoc()
	return NewConstantScoreScorer(c, c.score, c.scoreMode, index.DocIdSetIteratorAll(maxDoc))
}

func (c *constantScoreWeight) IsCacheable(ctx *index.LeafReaderContext) bool {
	return true
}

func (c *constantScoreWeight) BulkScorer(context *index.LeafReaderContext) (BulkScorer, error) {
	if c.scoreMode.IsExhaustive() == false {
		return c.ConstantScoreWeight.BulkScorer(context)
	}

	score := c.score
	maxDoc := context.Reader().MaxDoc()

	return &BulkScorerDefault{
		FnScoreRange: func(collector LeafCollector, acceptDocs util.Bits, fromDoc, toDoc int) (int, error) {
			toDoc = min(maxDoc, toDoc)
			scorer := NewScoreAndDoc()
			scorer.score = score
			if err := collector.SetScorer(scorer); err != nil {
				return 0, err
			}
			for doc := fromDoc; doc < toDoc; doc++ {
				scorer.doc = doc
				if acceptDocs == nil || acceptDocs.Test(uint(doc)) {
					err := collector.Collect(context2.Background(), doc)
					if err != nil {
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
