package index

import (
	"context"

	"github.com/geange/gods-generic/sets/treeset"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util"
	"github.com/geange/lucene-go/core/util/automaton"
)

type ScoreDoc interface {
	GetScore() float64
	SetScore(score float64)
	GetDoc() int
	SetDoc(doc int)
	GetShardIndex() int
	SetShardIndex(shardIndex int)
}

// Scorable
// Allows access to the Score of a Query
// 允许访问查询的分数
type Scorable interface {
	// Score
	// Returns the Score of the current document matching the query.
	Score() (float64, error)

	// SmoothingScore
	// Returns the smoothing Score of the current document matching the query. This Score
	// is used when the query/term does not appear in the document, and behaves like an idf. The smoothing
	// Score is particularly important when the Scorer returns a product of probabilities so that the
	// document Score does not go to zero when one probability is zero. This can return 0 or a smoothing Score.
	//
	// Smoothing scores are described in many papers, including: Metzler, D. and Croft, W. B. , "Combining
	// the Language Model and Inference Network Approaches to Retrieval," Information Processing and Management
	// Special Issue on Bayesian Networks and Information Retrieval, 40(5), pp.735-750.
	SmoothingScore(docId int) (float64, error)

	// DocID
	// Returns the doc ID that is currently being scored.
	DocID() int

	// SetMinCompetitiveScore
	// Optional method: Tell the scorer that its iterator may safely ignore all
	// documents whose Score is less than the given minScore. This is a no-op by default. This method
	// may only be called from collectors that use ScoreMode.TOP_SCORES, and successive calls may
	// only set increasing values of minScore.
	SetMinCompetitiveScore(minScore float64) error

	// GetChildren
	// Returns child sub-scorers positioned on the current document
	GetChildren() ([]ChildScorable, error)
}

type ChildScorable interface {
	GetChild() Scorable
	GetRelationship() string
}

// SegmentCacheable
// Interface defining whether or not an object can be cached against a LeafReader Objects
// that depend only on segment-immutable structures such as Points or postings lists can
// just return true from isCacheable(LeafReaderContext) Objects that depend on doc values
// should return DocValues.isCacheable(LeafReaderContext, String...), which will check to
// see if the doc values fields have been updated. Updated doc values fields are not suitable
// for cacheing. Objects that are not segment-immutable, such as those that rely on global
// statistics or scores, should return false
type SegmentCacheable interface {

	// IsCacheable
	// Returns: true if the object can be cached against a given leaf
	IsCacheable(ctx LeafReaderContext) bool
}

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
// maintain the relationship between the searcher's top-level ReaderContext and the context used to create
//
// 由于权重为给定的LeafReaderContext（Scorer（LeafReaderContext））创建记分器实例，因此调用程序必须保持搜索器的顶级索引
// ReaderContext和用于创建记分器的上下文之间的关系。
//
// a Scorer.
// A Weight is used in the following way:
// A Weight is constructed by a top-level query, given a IndexSearcher (Query.createWeight(IndexSearcher, ScoreMode, float)).
// A Scorer is constructed by scorer(LeafReaderContext).
// Since: 2.9
type Weight interface {
	SegmentCacheable

	ExtractTerms(terms *treeset.Set[Term]) error

	// Matches
	// Returns Matches for a specific document, or null if the document does not match the parent query
	// A query match that contains no position information (for example, a Point or DocValues query) will
	// return MatchesUtils.MATCH_WITH_NO_TERMS
	// context: the reader's context to create the Matches for
	// doc: the document's id relative to the given context's reader
	Matches(readerContext LeafReaderContext, doc int) (Matches, error)

	// Explain
	// An explanation of the score computation for the named document.
	// context: the readers context to create the Explanation for.
	// doc: the document's id relative to the given context's reader
	// Returns: an Explanation for the score
	// Throws: 	IOException – if an IOException occurs
	Explain(readerContext LeafReaderContext, doc int) (types.Explanation, error)

	// GetQuery The query that this concerns.
	GetQuery() Query

	// Scorer
	// Returns a Scorer which can iterate in order over all matching documents and assign them a score.
	// NOTE: null can be returned if no documents will be scored by this query.
	// NOTE: The returned Scorer does not have LeafReader.getLiveDocs() applied, they need to be checked on top.
	// ctx: the LeafReaderContext for which to return the Scorer.
	// a Scorer which scores documents in/out-of order.
	Scorer(ctx LeafReaderContext) (Scorer, error)

	// ScorerSupplier
	// Optional method. Get a ScorerSupplier, which allows to know the cost of the Scorer before building it.
	// The default implementation calls scorer and builds a ScorerSupplier wrapper around it.
	ScorerSupplier(ctx LeafReaderContext) (ScorerSupplier, error)

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
	BulkScorer(ctx LeafReaderContext) (BulkScorer, error)
}

type ScorerSupplier interface {
	// Get
	// Get the Scorer. This may not return null and must be called at most once.
	// leadCost: Cost of the scorer that will be used in order to lead iteration. This can be
	//			interpreted as an upper bound of the number of times that DocIdSetIterator.nextDoc,
	//			DocIdSetIterator.advance and TwoPhaseIterator.matches will be called. Under doubt,
	//			pass Long.MAX_VALUE, which will produce a Scorer that has good iteration capabilities.
	Get(leadCost int64) (Scorer, error)

	// Cost
	// Get an estimate of the Scorer that would be returned by get. This may be a costly operation,
	// so it should only be called if necessary.
	// See Also: DocIdSetIterator.cost
	Cost() int64
}

// MatchesIterator
// An iterator over match positions (and optionally offsets) for a single document and field To iterate over
// the matches, call next() until it returns false, retrieving positions and/or offsets after each call.
// You should not call the position or offset methods before next() has been called, or after next() has
// returned false. Matches from some queries may span multiple positions. You can retrieve the positions
// of individual matching terms on the current match by calling getSubMatches(). Matches are ordered by
// start position, and then by end position. Match intervals may overlap.
// See Also: Weight.matches(LeafReaderContext, int)
type MatchesIterator interface {

	// Next
	// Advance the iterator to the next match position
	// Returns: true if matches have not been exhausted
	Next() (bool, error)

	// StartPosition
	// The start position of the current match OccurShould only be called after next() has returned true
	StartPosition() int

	// EndPosition
	// The end position of the current match OccurShould only be called after next() has returned true
	EndPosition() int

	// StartOffset
	// The starting offset of the current match, or -1 if offsets are not available OccurShould only be
	// called after next() has returned true
	StartOffset() (int, error)

	// EndOffset
	// The ending offset of the current match, or -1 if offsets are not available OccurShould only be
	// called after next() has returned true
	EndOffset() (int, error)

	// GetSubMatches
	// Returns a MatchesIterator that iterates over the positions and offsets of individual
	// terms within the current match Returns null if there are no submatches (ie the current iterator is
	// at the leaf level) OccurShould only be called after next() has returned true
	GetSubMatches() (MatchesIterator, error)

	// GetQuery
	// Returns the Query causing the current match If this MatchesIterator has been returned from a
	// getSubMatches() call, then returns a TermQuery equivalent to the current match OccurShould only be called
	// after next() has returned true
	GetQuery() Query
}

type IndexSearcher interface {
	SetQueryCache(queryCache QueryCache)
	GetQueryCache() QueryCache
	SetQueryCachingPolicy(queryCachingPolicy QueryCachingPolicy)
	GetQueryCachingPolicy() QueryCachingPolicy
	Slices(leaves []LeafReaderContext) []LeafSlice
	GetIndexReader() IndexReader
	Doc(docID int) (*document.Document, error)
	DocWithVisitor(docID int, fieldVisitor document.StoredFieldVisitor) (*document.Document, error)
	DocLimitFields(docID int, fieldsToLoad []string) (*document.Document, error)
	SetSimilarity(similarity Similarity)
	GetSimilarity() Similarity
	Count(query Query) (int, error)
	GetSlices() []LeafSlice
	CreateWeight(query Query, scoreMode ScoreMode, boost float64) (Weight, error)
	TermStatistics(term Term, docFreq, totalTermFreq int) (types.TermStatistics, error)
	CollectionStatistics(field string) (types.CollectionStatistics, error)
	GetTopReaderContext() IndexReaderContext
	Search(query Query, results Collector) error
	SearchTopN(query Query, n int) (TopDocs, error)
	SearchCollector(query Query, results Collector) error
	Search3(leaves []LeafReaderContext, weight Weight, collector Collector) error
}

type LeafSlice struct {
	Leaves []LeafReaderContext
}

// QueryCache
// A cache for queries.
// See Also: LRUQueryCache
type QueryCache interface {

	// DoCache
	// Return a wrapper around the provided weight that will cache matching docs per-segment accordingly to
	// the given policy. NOTE: The returned weight will only be equivalent if scores are not needed.
	// See Also: Collector.scoreMode()
	DoCache(weight Weight, policy QueryCachingPolicy) Weight
}

// QueryCachingPolicy
// A policy defining which filters should be cached. Implementations of this class must be thread-safe.
// See Also: UsageTrackingQueryCachingPolicy, LRUQueryCache
type QueryCachingPolicy interface {
	// OnUse
	// Callback that is called every time that a cached filter is used. This is typically useful if the
	// policy wants to track usage statistics in order to make decisions.
	OnUse(query Query)

	// ShouldCache
	// Whether the given Query is worth caching. This method will be called by the QueryCache to
	// know whether to cache. It will first attempt to load a DocIdSet from the cache. If it is not cached yet
	// and this method returns true then a cache entry will be generated. Otherwise an uncached scorer will be returned.
	ShouldCache(query Query) (bool, error)
}

// Query
// The abstract base class for queries.
// * Instantiable subclasses are:
// * TermQuery
// * BooleanQuery
// * WildcardQuery
// * PhraseQuery
// * PrefixQuery
// * MultiPhraseQuery
// * FuzzyQuery
// * RegexpQuery
// * TermRangeQuery
// * PointRangeQuery
// * ConstantScoreQuery
// * DisjunctionMaxQuery
// * MatchAllDocsQuery
// See also the family of Span Queries and additional queries available in the Queries module
type Query interface {

	// CreateWeight
	// Expert: Constructs an appropriate Weight implementation for this query.
	// Only implemented by primitive queries, which re-write to themselves.
	// scoreMode: How the produced scorers will be consumed.
	// boost: The boost that is propagated by the parent queries.
	CreateWeight(searcher IndexSearcher, scoreMode ScoreMode, boost float64) (Weight, error)

	// Rewrite
	// Expert: called to re-write queries into primitive queries. For example, a PrefixQuery will be
	// rewritten into a BooleanQuery that consists of TermQuerys.
	Rewrite(reader IndexReader) (Query, error)

	// Visit
	// Recurse through the query tree, visiting any child queries
	// visitor: a QueryVisitor to be called by each query in the tree
	Visit(visitor QueryVisitor) error

	// String
	// Convert a query to a string, with field assumed to be the default field and omitted.
	String(field string) string
}

type QueryExt interface {
	IsPointQuery() bool
}

// QueryVisitor
// Allows recursion through a query tree
// See Also: Query.visit(QueryVisitor)
type QueryVisitor interface {

	// ConsumeTerms
	// Called by leaf queries that match on specific terms
	// query: the leaf query
	// terms: the terms the query will match on
	ConsumeTerms(query Query, terms ...Term)

	// ConsumeTermsMatching
	// Called by leaf queries that match on a class of terms
	// query: the leaf query
	// field: the field queried against
	// automaton: a supplier for an automaton defining which terms match
	ConsumeTermsMatching(query Query, field string, automaton func() *automaton.ByteRunAutomaton)

	// VisitLeaf
	// Called by leaf queries that do not match on terms
	// query: the query
	VisitLeaf(query Query) (err error)

	// AcceptField
	// Whether or not terms from this field are of interest to the visitor Implement this to
	// avoid collecting terms from heavy queries such as TermInSetQuery that are not running
	// on fields of interest
	AcceptField(field string) bool

	// GetSubVisitor
	// Pulls a visitor instance for visiting child clauses of a query The default implementation
	// returns this, unless occur is equal to BooleanClause.Occur.OccurMustNot in which case it
	// returns EMPTY_VISITOR
	// occur: the relationship between the parent and its children
	// parent: the query visited
	GetSubVisitor(occur Occur, parent Query) QueryVisitor
}

// Occur
// Specifies how clauses are to occur in matching documents.
type Occur string

func (o Occur) String() string {
	return string(o)
}

func OccurValues() []Occur {
	return []Occur{
		OccurMust, OccurFilter, OccurShould, OccurMustNot,
	}
}

const (
	// OccurMust
	// Use this operator for clauses that must appear in the matching documents.
	// 等同于 AND
	OccurMust = Occur("+")

	// OccurFilter
	// Like OccurMust except that these clauses do not participate in scoring.
	OccurFilter = Occur("#")

	// OccurShould
	// Use this operator for clauses that should appear in the matching documents.
	// For a BooleanQuery with no OccurMust clauses one or more OccurShould clauses must match
	// a document for the BooleanQuery to match.
	// See Also: BooleanQuery.BooleanQueryBuilder.setMinimumNumberShouldMatch
	// 等同于 OR
	OccurShould = Occur("")

	// OccurMustNot
	// Use this operator for clauses that must not appear in the matching documents.
	// Note that it is not possible to search for queries that only consist of a OccurMustNot clause.
	// These clauses do not contribute to the score of documents.
	// 等同于 NOT
	OccurMustNot = Occur("-")
)

// Matches
// Reports the positions and optionally offsets of all matching terms in a query for a single document
// To obtain a MatchesIterator for a particular field, call GetMatches(String). Note that you can call
// GetMatches(String) multiple times to retrieve new iterators, but it is not thread-safe.
// 报告单个文档的查询中所有匹配项的位置和可选偏移量，以获取特定字段的匹配迭代器，称为getMatches（String）。
// 注意，可以多次调用getMatches（String）来检索新的迭代器，但它不是线程安全的。
type Matches interface {
	Strings() []string

	// GetMatches
	// Returns a MatchesIterator over the matches for a single field, or null if there are no matches
	// in that field.
	GetMatches(field string) (MatchesIterator, error)

	// GetSubMatches
	// Returns a collection of Matches that make up this instance; if it is not a composite,
	// then this returns an empty list
	GetSubMatches() []Matches
}

// Scorer
// Expert: Common scoring functionality for different types of queries.
// 不同类型查询的通用评分功能。
//
// A Scorer exposes an iterator() over documents matching a query in increasing order of doc Id.
// 计分器暴露一个迭代器，这个迭代器按照文档id递增顺序
//
// Document scores are computed using a given Similarity implementation.
// NOTE: The values Float.Nan, Float.NEGATIVE_INFINITY and Float.POSITIVE_INFINITY are not valid scores.
// Certain collectors (eg TopScoreDocCollector) will not properly collect hits with these scores.
type Scorer interface {
	Scorable

	// GetWeight
	// returns parent Weight
	GetWeight() Weight

	// Iterator
	// Return a DocIdSetIterator over matching documents. The returned iterator will either
	// be positioned on -1 if no documents have been scored yet, DocIdSetIterator.NO_MORE_DOCS if all
	// documents have been scored already, or the last document id that has been scored otherwise.
	// The returned iterator is a view: calling this method several times will return iterators
	// that have the same state.
	Iterator() types.DocIdSetIterator

	// TwoPhaseIterator
	// Optional method: Return a TwoPhaseIterator view of this Scorer. A return value
	// of null indicates that two-phase iteration is not supported. Note that the returned
	// TwoPhaseIterator's approximation must advance synchronously with the iterator(): advancing
	// the approximation must advance the iterator and vice-versa. Implementing this method is
	// typically useful on Scorers that have a high per-document overhead in order to confirm
	// matches. The default implementation returns null.
	TwoPhaseIterator() TwoPhaseIterator

	// AdvanceShallow
	// Advance to the block of documents that contains target in order to get scoring information
	// about this block. This method is implicitly called by DocIdSetIterator.advance(int) and
	// DocIdSetIterator.nextDoc() on the returned doc ID. Calling this method doesn't modify the
	// current DocIdSetIterator.docID(). It returns a number that is greater than or equal to all
	// documents contained in the current block, but less than any doc IDS of the next block.
	// target must be >= docID() as well as all targets that have been passed to advanceShallow(int) so far.
	AdvanceShallow(target int) (int, error)

	// GetMaxScore
	// Return the maximum score that documents between the last target that this iterator
	// was shallow-advanced to included and upTo included.
	GetMaxScore(upTo int) (float64, error)
}

// TwoPhaseIterator
// Returned by Scorer.TwoPhaseIterator() to expose an approximation of a DocIdSetIterator.
// When the approximation()'s DocIdSetIterator.nextDoc() or DocIdSetIterator.advance(int) return,
// matches() needs to be checked in order to know whether the returned doc ID actually matches.
//
// # GPT3.5
//
// 在Lucene中，`TwoPhaseIterator`是一个用于执行两阶段迭代的工具类。
// 它可以与Scorer一起使用，用于更高效地过滤和评分匹配文档。
//
// 在搜索过程中，通常会使用一个Scorer来进行文档匹配，并对匹配的文档进行评分。然而，有时候在进行文档匹配之前，
// 可以使用一些更快速的方法来过滤掉不符合条件的文档，从而减少评分操作的开销。
//
// `TwoPhaseIterator`类就提供了这样的功能。它通过两个阶段的迭代来实现过滤和评分的分离。
//
// 在第一阶段，`TwoPhaseIterator`会对文档进行快速的过滤操作，根据一些预先计算的条件（例如，布尔表达式或位集合），
// 判断文档是否可能匹配查询条件。这个过滤操作通常比完全匹配文档的评分操作更快。
//
// 在第二阶段，对于通过第一阶段过滤的文档，`TwoPhaseIterator`会将这些文档传递给实际的Scorer进行详细的匹配和评分操作。
//
// 使用`TwoPhaseIterator`的好处是，它可以减少不必要的评分操作，只对通过过滤的文档进行实际的匹配和评分，从而提高搜索性能。
//
// `TwoPhaseIterator`类主要包含以下方法：
//
// 1. `approximation()`：返回用于快速过滤的近似评分器（approximation scorer）。
//
// 2. `matches()`：在第一阶段中，检查当前文档是否匹配查询条件。
//
// 3. `matchCost()`：返回第一阶段中过滤操作的成本。用于估算在第一阶段过滤后剩余的文档数量。
//
// 通过使用`TwoPhaseIterator`，可以在搜索过程中根据具体需求进行过滤和评分的优化，提高搜索性能并降低开销。
type TwoPhaseIterator interface {
	Approximation() types.DocIdSetIterator

	// Matches
	// Return whether the current doc ID that approximation() is on matches.
	// This method should only be called when the iterator is positioned -- ie. not when DocIdSetIterator.docID() is -1 or DocIdSetIterator.NO_MORE_DOCS -- and at most once.
	Matches() (bool, error)

	// MatchCost
	// An estimate of the expected cost to determine that a single document matches().
	// This can be called before iterating the documents of approximation().
	// Returns an expected cost in number of simple operations like addition, multiplication, comparing two numbers and indexing an array. The returned value must be positive.
	MatchCost() float64
}

// BulkScorer
// This class is used to Score a range of documents at once, and is returned by Weight.bulkScorer.
// Only queries that have a more optimized means of scoring across a range of documents need to override this.
// Otherwise, a default implementation is wrapped around the Scorer returned by Weight.scorer.
//
// GPT3.5：
// 这个类用于一次对一系列文档进行评分，它是由Weight.bulkScorer返回的。
// 只有那些在一系列文档上有更优化的评分方法的查询才需要覆盖它。
// 否则，会使用默认实现，该实现会封装在Weight.scorer返回的Scorer周围。
type BulkScorer interface {
	// Score Scores and collects all matching documents.
	// Params: 	collector – The collector to which all matching documents are passed.
	//			acceptDocs – Bits that represents the allowed documents to match, or null if they are all allowed to match.
	Score(collector LeafCollector, acceptDocs util.Bits) error

	// ScoreRange
	// Params:
	// 		collector – The collector to which all matching documents are passed.
	// 		acceptDocs – Bits that represents the allowed documents to match, or null if they are all allowed to match.
	// 		min – Score starting at, including, this document
	// 		max – Score up to, but not including, this doc
	// Returns: an under-estimation of the next matching doc after max
	ScoreRange(collector LeafCollector, acceptDocs util.Bits, min, max int) (int, error)

	// Cost Same as DocIdSetIterator.cost() for bulk scorers.
	Cost() int64
}

// LeafCollector
// Collector decouples the score from the collected doc: the score computation is skipped entirely
// if it's not needed. Collectors that do need the score should implement the setScorer method,
// to hold onto the passed Scorer instance, and call Scorer.score() within the collect method
// to compute the current hit's score. If your collector may request the score for a single hit
// multiple times, you should use ScoreCachingWrappingScorer.
//
// NOTE: The doc that is passed to the collect method is relative to the current reader. If your
// collector needs to resolve this to the docID space of the Multi*Reader, you must re-base it by
// recording the docBase from the most recent setNextReader call. Here's a simple example showing
// how to collect docIDs into a BitSet:
//
//	IndexSearcher searcher = new IndexSearcher(indexReader);
//	final BitSet bits = new BitSet(indexReader.maxDoc());
//	searcher.search(query, new Collector() {
//
//	  public LeafCollector getLeafCollector(LeafReaderContext context)
//	      throws IOException {
//	    final int docBase = context.docBase;
//	    return new LeafCollector() {
//
//	      // ignore scorer
//	      public void setScorer(Scorer scorer) throws IOException {
//	      }
//
//	      public void collect(int doc) throws IOException {
//	        bits.set(docBase + doc);
//	      }
//
//	    };
//	  }
//
//	});
//
// Not all collectors will need to rebase the docID. For example, a collector that simply counts the total
// number of hits would skip it.
type LeafCollector interface {
	// SetScorer Called before successive calls to collect(int). Implementations that need the score of
	// the current document (passed-in to collect(int)), should save the passed-in Scorer and call
	// scorer.score() when needed.
	//
	// 调用此方法通过Scorer对象获得一篇文档的打分，对文档集合进行排序时，可以作为排序条件之一
	SetScorer(scorer Scorable) error

	// Collect Called once for every document matching a query, with the unbased document number.
	// Note: The collection of the current segment can be terminated by throwing a CollectionTerminatedException.
	// In this case, the last docs of the current org.apache.lucene.index.LeafReaderContext will be skipped
	// and IndexSearcher will swallow the exception and continue collection with the next leaf.
	// Note: This is called in an inner search loop. For good search performance, implementations of this
	// method should not call IndexSearcher.doc(int) or org.apache.lucene.index.Reader.document(int) on
	// every hit. Doing so can slow searches by an order of magnitude or more.
	//
	// 在这个方法中实现了对所有满足查询条件的文档进行
	// 排序（sorting）、过滤（filtering）或者用户自定义的操作的具体逻辑。
	Collect(ctx context.Context, doc int) error

	// CompetitiveIterator Optionally returns an iterator over competitive documents. Collectors should
	// delegate this method to their comparators if their comparators provide the skipping functionality
	// over non-competitive docs. The default is to return null which is interpreted as the collector
	// provide any competitive iterator.
	CompetitiveIterator() (types.DocIdSetIterator, error)
}

const (
	isExhaustiveShift = 0
	needsScoresShift  = 1
)

// ScoreMode
// Different modes of search.
type ScoreMode uint8

// NeedsScores
// Whether this ScoreMode needs to compute scores.
func (r ScoreMode) NeedsScores() bool {
	return r&1<<needsScoresShift > 0
}

// IsExhaustive
// Returns true if for this ScoreMode it is necessary to process all documents, or false if
// is enough to go through top documents only.
func (r ScoreMode) IsExhaustive() bool {
	return r&1<<isExhaustiveShift > 0
}
