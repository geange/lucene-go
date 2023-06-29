package search

import (
	"errors"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util"
	"io"
	"math"
)

// Weight
// Expert: Calculate query weights and build query scorers.
// 计算查询权重并构建查询记分器。
//
// The purpose of Weight is to ensure searching does not modify a Query, so that a Query instance can be reused.
// IndexSearcher dependent state of the query should reside in the Weight. LeafReader dependent state should
// reside in the Scorer.
// 权重的目的是确保搜索不会修改查询，以便可以重用查询实例。查询的IndexSearcher依赖状态应位于权重中。LeafReader 相关状态应位于记分器中。
//
// Since Weight creates Scorer instances for a given LeafReaderContext (scorer(LeafReaderContext)) callers must
// maintain the relationship between the searcher's top-level IndexReaderContext and the context used to create
//
// 由于权重为给定的LeafReaderContext（Scorer（LeafReaderContext））创建记分器实例，因此调用程序必须保持搜索器的顶级索引
// ReaderContext和用于创建记分器的上下文之间的关系。
// a Scorer.
// A Weight is used in the following way:
// A Weight is constructed by a top-level query, given a IndexSearcher (Query.createWeight(IndexSearcher, ScoreMode, float)).
// A Scorer is constructed by scorer(LeafReaderContext).
// Since: 2.9
type Weight interface {
	SegmentCacheable

	ExtractTerms(terms []*index.Term) error

	// Matches Returns Matches for a specific document, or null if the document does not match the parent query A query match that contains no position information (for example, a Point or DocValues query) will return MatchesUtils.MATCH_WITH_NO_TERMS
	// Params: 	context – the reader's context to create the Matches for
	//			doc – the document's id relative to the given context's reader
	Matches(context *index.LeafReaderContext, doc int) (Matches, error)

	// Explain An explanation of the score computation for the named document.
	// Params: 	context – the readers context to create the Explanation for.
	//			doc – the document's id relative to the given context's reader
	// Returns: an Explanation for the score
	// Throws: 	IOException – if an IOException occurs
	Explain(ctx *index.LeafReaderContext, doc int) (*types.Explanation, error)

	// GetQuery The query that this concerns.
	GetQuery() Query

	// Scorer Returns a Scorer which can iterate in order over all matching documents and assign them a score.
	//NOTE: null can be returned if no documents will be scored by this query.
	//NOTE: The returned Scorer does not have LeafReader.getLiveDocs() applied, they need to be checked on top.
	//Params:
	//context – the LeafReaderContext for which to return the Scorer.
	//Returns:
	//a Scorer which scores documents in/out-of order.
	//Throws:
	//IOException – if there is a low-level I/O error
	Scorer(ctx *index.LeafReaderContext) (Scorer, error)

	// ScorerSupplier Optional method. Get a ScorerSupplier, which allows to know the cost of the Scorer before building it. The default implementation calls scorer and builds a ScorerSupplier wrapper around it.
	//See Also:
	//scorer
	ScorerSupplier(ctx *index.LeafReaderContext) (ScorerSupplier, error)

	// BulkScorer
	// Optional method, to return a BulkScorer to score the query and send hits to a Collector.
	// Only queries that have a different top-level approach need to override this;
	// the default implementation pulls a normal Scorer and iterates and collects
	// the resulting hits which are not marked as deleted.
	//
	// context: the LeafReaderContext for which to return the Scorer.
	//
	// Returns: a BulkScorer which scores documents and passes them to a collector.
	// Throws: 	IOException – if there is a low-level I/O error
	//
	// GPT3.5:
	// 可选方法，用于返回一个BulkScorer，对查询进行评分并将结果传递给Collector。
	// 只有那些具有不同顶层方法的查询才需要覆盖此方法；默认实现获取一个普通的Scorer，
	// 迭代并收集未标记为删除的结果hits。
	//
	// 参数：
	// context - 要返回Scorer的LeafReaderContext。
	// 返回： 一个BulkScorer，对文档进行评分并将其传递给Collector。
	BulkScorer(ctx *index.LeafReaderContext) (BulkScorer, error)
}

type WeightSPI interface {
	Scorer(ctx *index.LeafReaderContext) (Scorer, error)
}

type WeightDefault struct {
	WeightSPI

	parentQuery Query
}

func NewWeight(parentQuery Query, extra WeightSPI) *WeightDefault {
	return &WeightDefault{
		WeightSPI:   extra,
		parentQuery: parentQuery,
	}
}

func (r *WeightDefault) ExtractTerms(terms []*index.Term) error {
	return nil
}

func (r *WeightDefault) GetQuery() Query {
	return r.parentQuery
}

func (r *WeightDefault) IsCacheable(ctx *index.LeafReaderContext) bool {
	return false
}

func (r *WeightDefault) Matches(ctx *index.LeafReaderContext, doc int) (Matches, error) {
	scorerSupplier, err := r.ScorerSupplier(ctx)
	if err != nil {
		return nil, err
	}
	if scorerSupplier == nil {
		return nil, nil
	}

	scorer, err := scorerSupplier.Get(1)
	if err != nil {
		return nil, err
	}
	twoPhase := scorer.TwoPhaseIterator()
	if twoPhase == nil {
		advance, err := scorer.Iterator().Advance(doc)
		if err != nil {
			return nil, err
		}
		if advance != doc {
			return nil, nil
		}
	} else {
		advance, err := twoPhase.Approximation().Advance(doc)
		if err != nil {
			return nil, err
		}

		if ok, _ := twoPhase.Matches(); advance != doc || !ok {
			return nil, nil
		}
	}
	return nil, errors.New("MATCH_WITH_NO_TERMS")
}

func (r *WeightDefault) ScorerSupplier(ctx *index.LeafReaderContext) (ScorerSupplier, error) {
	scorer, err := r.Scorer(ctx)
	if err != nil {
		return nil, err
	}
	if scorer == nil {
		return nil, nil
	}

	return &scorerSupplier{scorer: scorer}, nil
}

//func (r *WeightDefault) Scorer(ctx *index.LeafReaderContext) (Scorer, error) {
//	return r.WeightSPI.(interface {
//		Scorer(ctx *index.LeafReaderContext) (Scorer, error)
//	}).Scorer(ctx)
//}

var _ ScorerSupplier = &scorerSupplier{}

type scorerSupplier struct {
	scorer Scorer
}

func (s *scorerSupplier) Get(leadCost int64) (Scorer, error) {
	return s.scorer, nil
}

func (s *scorerSupplier) Cost() int64 {
	return s.scorer.Iterator().Cost()
}

func (r *WeightDefault) BulkScorer(ctx *index.LeafReaderContext) (BulkScorer, error) {
	scorer, err := r.Scorer(ctx)
	if err != nil {
		return nil, err
	}

	if scorer == nil {
		return nil, nil
	}

	return NewDefaultBulkScorer(scorer), nil
}

var _ BulkScorer = &DefaultBulkScorer{}

type DefaultBulkScorer struct {
	scorer   Scorer
	iterator index.DocIdSetIterator
	twoPhase TwoPhaseIterator
}

func NewDefaultBulkScorer(scorer Scorer) *DefaultBulkScorer {
	return &DefaultBulkScorer{
		scorer:   scorer,
		iterator: scorer.Iterator(),
		twoPhase: scorer.TwoPhaseIterator(),
	}
}

func (d *DefaultBulkScorer) Score(collector LeafCollector, acceptDocs util.Bits) error {
	NoMoreDocs := math.MaxInt32
	_, err := d.ScoreRange(collector, acceptDocs, 0, NoMoreDocs)
	return err
}

func (d *DefaultBulkScorer) ScoreRange(collector LeafCollector, acceptDocs util.Bits, min, max int) (int, error) {
	err := collector.SetScorer(d.scorer)
	if err != nil {
		return 0, err
	}

	scorerIterator := func() index.DocIdSetIterator {
		if d.twoPhase == nil {
			return d.iterator
		}
		return d.twoPhase.Approximation()
	}()

	competitiveIterator, err := collector.CompetitiveIterator()
	if err != nil {
		return 0, err
	}

	var filteredIterator index.DocIdSetIterator
	if competitiveIterator == nil {
		filteredIterator = scorerIterator
	} else {
		// Wrap CompetitiveIterator and ScorerIterator start with (i.e., calling nextDoc()) the last
		// visited docID because ConjunctionDISI might have advanced to it in the previous
		// scoreRange, but we didn't process due to the range limit of scoreRange.
		if scorerIterator.DocID() != -1 {
			scorerIterator = NewStartDISIWrapper(scorerIterator)
		}

		if competitiveIterator.DocID() != -1 {
			competitiveIterator = NewStartDISIWrapper(competitiveIterator)
		}

		filteredIterator = IntersectIterators([]index.DocIdSetIterator{
			scorerIterator,
			competitiveIterator,
		})
	}

	if filteredIterator.DocID() == -1 && min == 0 && max == index.NO_MORE_DOCS {
		err := scoreAll(collector, filteredIterator, d.twoPhase, acceptDocs)
		if err != nil {
			return 0, err
		}
		return index.NO_MORE_DOCS, nil
	} else {
		doc := filteredIterator.DocID()
		if doc < min {
			doc, err = filteredIterator.Advance(min)
			if err != nil {
				return 0, err
			}
		}
		return scoreRange(collector, filteredIterator, d.twoPhase, acceptDocs, doc, max)
	}
}

func scoreAll(collector LeafCollector, iterator index.DocIdSetIterator,
	twoPhase TwoPhaseIterator, acceptDocs util.Bits) error {

	doc, err := iterator.NextDoc()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}

	if twoPhase == nil {
		for {
			if acceptDocs == nil || acceptDocs.Test(uint(doc)) {
				err := collector.Collect(nil, doc)
				if err != nil {
					return err
				}
			}

			doc, err = iterator.NextDoc()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}
				return err
			}
		}
	} else {
		// The scorer has an approximation, so run the approximation first, then check acceptDocs, then confirm
		for {
			if ok, _ := twoPhase.Matches(); ok && (acceptDocs == nil || acceptDocs.Test(uint(doc))) {
				if err := collector.Collect(nil, doc); err != nil {
					return err
				}
			}

			doc, err = iterator.NextDoc()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}
				return err
			}
		}
	}

}

func scoreRange(collector LeafCollector, iterator index.DocIdSetIterator, twoPhase TwoPhaseIterator,
	acceptDocs util.Bits, currentDoc, end int) (int, error) {

	var err error

	if twoPhase == nil {
		for currentDoc < end {
			if acceptDocs == nil || acceptDocs.Test(uint(currentDoc)) {
				err := collector.Collect(nil, currentDoc)
				if err != nil {
					return 0, err
				}
			}
			currentDoc, err = iterator.NextDoc()
			if err != nil {
				return 0, err
			}
		}
		return currentDoc, nil
	} else {
		for currentDoc < end {
			if ok, _ := twoPhase.Matches(); ok && (acceptDocs == nil || acceptDocs.Test(uint(currentDoc))) {
				err := collector.Collect(nil, currentDoc)
				if err != nil {
					return 0, err
				}
			}

			currentDoc, err = iterator.NextDoc()
			if err != nil {
				return 0, err
			}
		}
		return currentDoc, nil
	}
}

func (d *DefaultBulkScorer) Cost() int64 {
	return d.iterator.Cost()
}

var _ index.DocIdSetIterator = &StartDISIWrapper{}

type StartDISIWrapper struct {
	in         index.DocIdSetIterator
	startDocID int
	docID      int
}

func NewStartDISIWrapper(in index.DocIdSetIterator) *StartDISIWrapper {
	return &StartDISIWrapper{
		in:         in,
		startDocID: in.DocID(),
	}
}

func (s *StartDISIWrapper) DocID() int {
	return s.docID
}

func (s *StartDISIWrapper) NextDoc() (int, error) {
	return s.Advance(s.docID + 1)
}

func (s *StartDISIWrapper) Advance(target int) (int, error) {
	if target <= s.startDocID {
		s.docID = s.startDocID
		return s.docID, nil
	}
	var err error
	s.docID, err = s.in.Advance(target)
	return s.docID, err
}

func (s *StartDISIWrapper) SlowAdvance(target int) (int, error) {
	return index.SlowAdvance(s, target)
}

func (s *StartDISIWrapper) Cost() int64 {
	return s.in.Cost()
}
