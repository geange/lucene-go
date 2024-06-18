package search

import (
	"github.com/geange/lucene-go/core/interface/index"
	index2 "github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util/attribute"
	"github.com/geange/lucene-go/core/util/bytesref"
)

// MultiTermQuery
// An abstract Query that matches documents containing a subset of terms provided by a
// FilteredTermsEnum enumeration.
// This query cannot be used directly; you must subclass it and define getTermsEnum(Terms, AttributeSource) to
// provide a FilteredTermsEnum that iterates through the terms to be matched.
// NOTE: if setRewriteMethod is either CONSTANT_SCORE_BOOLEAN_REWRITE or SCORING_BOOLEAN_REWRITE, you may
// encounter a BooleanQuery.TooManyClauses exception during searching, which happens when the number of terms
// to be searched exceeds BooleanQuery.getMaxClauseCount(). Setting setRewriteMethod to ConstantScoreRewrite
// prevents this.
// The recommended rewrite method is ConstantScoreRewrite: it doesn't spend CPU computing unhelpful scores,
// and is the most performant rewrite method given the query. If you need scoring (like FuzzyQuery,
// use MultiTermQuery.TopTermsScoringBooleanQueryRewrite which uses a priority queue to only collect
// competitive terms and not hit this limitation. Note that org.apache.lucene.queryparser.classic.QueryParser
// produces MultiTermQueries using ConstantScoreRewrite by default.
type MultiTermQuery interface {
	Query

	// GetField
	// Returns the field name for this query
	GetField() string

	// GetTermsEnum
	// Construct the enumeration to be used, expanding the pattern term.
	// This method should only be called if the field exists
	// (ie, implementations can assume the field does exist).
	// This method should not return null (should instead return TermsEnum.EMPTY if no terms match).
	// The TermsEnum must already be positioned to the first matching term.
	// The given AttributeSource is passed by the MultiTermQuery.RewriteMethod to
	// share information between segments, for example TopTermsRewrite uses it to
	// share maximum competitive boosts
	GetTermsEnum(terms index.Terms, atts *attribute.Source) (index.TermsEnum, error)

	// GetRewriteMethod
	// See Also: setRewriteMethod
	GetRewriteMethod() RewriteMethod

	// SetRewriteMethod
	// Sets the rewrite method to be used when executing the query. You can use one of the four core methods,
	// or implement your own subclass of MultiTermQuery.RewriteMethod.
	SetRewriteMethod(method RewriteMethod)
}

type MultiTermQueryPlus interface {
}

// RewriteMethod
// Abstract class that defines how the query is rewritten.
type RewriteMethod interface {
	Rewrite(reader index.IndexReader, query MultiTermQuery) (Query, error)

	// GetTermsEnum
	// Returns the MultiTermQuerys TermsEnum
	// See Also: getTermsEnum(Terms, AttributeSource)
	GetTermsEnum(query MultiTermQuery, terms index.Terms, atts *attribute.Source) (index.TermsEnum, error)
}

var _ RewriteMethod = &constantScoreRewrite{}

type constantScoreRewrite struct {
}

func (c *constantScoreRewrite) Rewrite(reader index.IndexReader, query MultiTermQuery) (Query, error) {
	//TODO implement me
	panic("implement me")
}

func (c *constantScoreRewrite) GetTermsEnum(query MultiTermQuery, terms index.Terms, atts *attribute.Source) (index.TermsEnum, error) {
	return query.GetTermsEnum(terms, atts)
}

var _ Query = &MultiTermQueryConstantScoreWrapper{}

type MultiTermQueryConstantScoreWrapper struct {
	query MultiTermQuery
	field string
}

func (m *MultiTermQueryConstantScoreWrapper) String(field string) string {
	// query.toString should be ok for the filter, too, if the query boost is 1.0f
	return m.query.String(field)
}

func (m *MultiTermQueryConstantScoreWrapper) GetQuery() Query {
	return m.query
}

// GetField
// Returns the field name for this query
func (m *MultiTermQueryConstantScoreWrapper) GetField() string {
	return m.query.GetField()
}

func (m *MultiTermQueryConstantScoreWrapper) CreateWeight(searcher *IndexSearcher, scoreMode ScoreMode, boost float64) (Weight, error) {
	//TODO implement me
	panic("implement me")
}

type wrapperConstantScoreWeight struct {
	*ConstantScoreWeight

	scoreMode ScoreMode
	p         *MultiTermQueryConstantScoreWrapper
}

func (r *wrapperConstantScoreWeight) BulkScorer(ctx index.LeafReaderContext) (BulkScorer, error) {
	weightOrBitSet, err := r.rewrite(ctx)
	if err != nil {
		return nil, err
	}
	if weightOrBitSet.weight != nil {
		return weightOrBitSet.weight.BulkScorer(ctx)
	}

	scorer, err := r.scorer(weightOrBitSet.set)
	if err != nil {
		return nil, err
	}
	if scorer == nil {
		return nil, nil
	}
	return NewDefaultBulkScorer(scorer), nil
}

func (r *wrapperConstantScoreWeight) Matches(context index.LeafReaderContext, doc int) (Matches, error) {
	terms, err := context.Reader().(index.LeafReader).Terms(r.p.query.GetField())
	if err != nil {
		return nil, err
	}
	if terms == nil {
		return nil, nil
	}
	if terms.HasPositions() == false {
		return r.ConstantScoreWeight.Matches(context, doc)
	}

	termsEnum, err := r.p.query.GetTermsEnum(terms, attribute.NewSource())
	if err != nil {
		return nil, err
	}

	return MatchesForField(r.p.query.GetField(), &matches{
		context: context,
		doc:     doc,
		query:   r.p.query,
		field:   r.p.query.GetField(),
		terms:   termsEnum,
	}), nil
}

var _ IOSupplier[MatchesIterator] = &matches{}

type matches struct {
	context index.LeafReaderContext
	doc     int
	query   Query
	field   string
	terms   bytesref.BytesIterator
}

func (r *matches) Get() (MatchesIterator, error) {
	return FromTermsEnumMatchesIterator(r.context, r.doc, r.query, r.field, r.terms)
}

func (r *wrapperConstantScoreWeight) Scorer(ctx index.LeafReaderContext) (Scorer, error) {
	weightOrBitSet, err := r.rewrite(ctx)
	if err != nil {
		return nil, err
	}
	if weightOrBitSet.weight != nil {
		return weightOrBitSet.weight.Scorer(ctx)
	}
	return r.scorer(weightOrBitSet.set)
}

func (r *wrapperConstantScoreWeight) IsCacheable(ctx index.LeafReaderContext) bool {
	return true
}

// Try to collect terms from the given terms enum and return true iff all
// terms could be collected. If {@code false} is returned, the enum is
// left positioned on the next term.
func (r *wrapperConstantScoreWeight) collectTerms(ctx index.LeafReaderContext, termsEnum index.TermsEnum, terms []*termAndState) {
	panic("")
}

// On the given leaf context, try to either rewrite to a disjunction if
// there are few terms, or build a bitset containing matching docs.
func (r *wrapperConstantScoreWeight) rewrite(ctx index.LeafReaderContext) (*weightOrDocIdSet, error) {
	panic("")
}

func (r *wrapperConstantScoreWeight) scorer(set DocIdSet) (Scorer, error) {
	if set == nil {
		return nil, nil
	}
	disi := set.Iterator()
	if disi == nil {
		return nil, nil
	}
	return NewConstantScoreScorer(r, r.Score(), r.scoreMode, disi)
}

type weightOrDocIdSet struct {
	weight Weight
	set    DocIdSet
}

type termAndState struct {
	term          []byte
	state         index2.TermState
	docFreq       int
	totalTermFreq int64
}

func newTermAndState(term []byte, state index2.TermState, docFreq int, totalTermFreq int64) *termAndState {
	return &termAndState{term: term, state: state, docFreq: docFreq, totalTermFreq: totalTermFreq}
}

func (m *MultiTermQueryConstantScoreWrapper) Rewrite(reader index.IndexReader) (Query, error) {
	return m, nil
}

func (m *MultiTermQueryConstantScoreWrapper) Visit(visitor QueryVisitor) (err error) {
	if visitor.AcceptField(m.GetField()) {
		return m.query.Visit(visitor.GetSubVisitor(OccurFilter, m))
	}
	return nil
}
