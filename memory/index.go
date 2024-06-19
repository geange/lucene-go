package memory

import (
	"errors"
	index2 "github.com/geange/lucene-go/core/interface/index"
	search2 "github.com/geange/lucene-go/core/interface/search"
	"reflect"

	"github.com/geange/gods-generic/maps/treemap"
	"github.com/geange/lucene-go/core/analysis"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/search"
	"github.com/geange/lucene-go/core/util/bytesref"
	"github.com/geange/lucene-go/core/util/ints"
)

// High-performance single-document main memory Apache Lucene fulltext search index.

// Index High-performance single-document main memory Apache Lucene fulltext search index.
// Overview
// This class is a replacement/substitute for a large subset of RAMDirectory functionality. It is designed to
// enable maximum efficiency for on-the-fly matchmaking combining structured and fuzzy fulltext search in
// realtime streaming applications such as Nux XQuery based XML message queues, publish-subscribe systems for
// Blogs/newsfeeds, text chat, data acquisition and distribution systems, application level routers, firewalls,
// classifiers, etc. Rather than targeting fulltext search of infrequent queries over huge persistent data
// archives (historic search), this class targets fulltext search of huge numbers of queries over comparatively
// small transient realtime data (prospective search). For example as in float
// score = search(String text, Query query)
//
// 这个类是RAMDirectory功能的一个子集的替代品。它的设计目的是在实时流应用程序中实现最大效率的动态匹配，
// 将结构化和模糊全文搜索相结合，如基于Nux XQuery的XML消息队列、博客/新闻源的发布订阅系统、文本聊天、数据采集和分发系统、
// 应用程序级路由器、防火墙、分类器等。该类的目标不是在巨大的持久数据档案中进行不频繁查询的全文搜索（历史搜索），
// 而是在相对较小的瞬态实时数据上进行大量查询的全文检索（前瞻性搜索）。例如，如
// score = search(String text, Query query)
//
// Each instance can hold at most one Lucene "document", with a document containing zero or more "fields",
// each field having a name and a fulltext value. The fulltext value is tokenized (split and transformed) into
// zero or more index terms (aka words) on addField(), according to the policy implemented by an Analyzer.
// For example, Lucene analyzers can split on whitespace, normalize to lower case for case insensitivity,
// ignore common terms with little discriminatory value such as "he", "in", "and" (stop words), reduce the terms
// to their natural linguistic root form such as "fishing" being reduced to "fish" (stemming), resolve
// synonyms/inflexions/thesauri (upon indexing and/or querying), etc. For details, see Lucene Analyzer intro.
//
// 每个实例最多可以包含一个Lucene“文档”，其中一个文档包含零个或多个“字段”，每个字段都有一个名称和一个全文值。
// 根据分析器实现的策略，全文值在addField（）上被标记（拆分和转换）为零个或多个索引项（也称为单词）。
// 例如，Lucene分析器可以对空白进行拆分，针对不区分大小写的情况将其标准化为小写，忽略“he”、“in”和“（停止词）
// 等几乎没有区别性的常用术语，将术语简化为其自然语言词根形式，如将“fishing”简化为“fish”（词干），
// 解析同义词/屈折词/词库（在索引和/或查询时）等。有关详细信息，请参阅Lucene Analyzer简介。
//
// Arbitrary Lucene queries can be run against this class - see Lucene Query Syntax as well as Query Parser Rules.
// Note that a Lucene query selects on the field names and associated (indexed) tokenized terms, not on the
// original fulltext(s) - the latter are not stored but rather thrown away immediately after tokenization.
//
// 可以针对此类运行任意Lucene查询-请参阅Lucene查询语法以及查询分析器规则。
// 请注意，Lucene查询在字段名称和相关联的（索引的）标记化术语上进行选择，
// 而不是在原始全文上进行选择——后者不会存储，而是在标记化后立即丢弃。
//
// For some interesting background information on search technology, see Bob Wyman's Prospective Search,
// Jim Gray's A Call to Arms - Custom subscriptions, and Tim Bray's On Search, the Series.
// Example Usage
//
//	Analyzer analyzer = new SimpleAnalyzer(version);
//	MemoryIndex index = new MemoryIndex();
//	index.addField("content", "Readings about Salmons and other select Alaska fishing Manuals", analyzer);
//	index.addField("author", "Tales of James", analyzer);
//	QueryParser parser = new QueryParser(version, "content", analyzer);
//	float score = index.search(parser.parse("+author:james +salmon~ +fish* manual~"));
//	if (score > 0.0f) {
//	    System.out.println("it's a match");
//	} else {
//	    System.out.println("no match found");
//	}
//	System.out.println("indexData=" + index.toString());
//
// Example XQuery Usage
//
//	(: An XQuery that finds all books authored by James that have something to do
//	with "salmon fishing manuals", sorted by relevance :)
//	declare namespace lucene = "java:nux.xom.pool.FullTextUtil";
//	declare variable $query := "+salmon~ +fish* manual~"; (: any arbitrary Lucene query can go here :)
//
//	for $book in /books/book[author="James" and lucene:match(abstract, $query) > 0.0]
//	let $score := lucene:match($book/abstract, $query)
//	order by $score descending
//	return $book
//
// Thread safety guarantees
// Index is not normally thread-safe for adds or queries. However, queries are thread-safe after
// freeze() has been called.
// Performance Notes
// Internally there's a new data structure geared towards efficient indexing and searching, plus the necessary
// support code to seamlessly plug into the Lucene framework.
// This class performs very well for very small texts (e.g. 10 chars) as well as for large texts (e.g. 10 MB)
// and everything in between. Typically, it is about 10-100 times faster than RAMDirectory. Note that
// RAMDirectory has particularly large efficiency overheads for small to medium sized texts, both in time and
// space. Indexing a field with N tokens takes O(N) in the best case, and O(N logN) in the worst case.
// Memory consumption is probably larger than for RAMDirectory.
// Example throughput of many simple term queries over a single Index: ~500000 queries/sec on a
// MacBook Pro, jdk 1.5.0_06, server VM. As always, your mileage may vary.
// If you're curious about the whereabouts of bottlenecks, run java 1.5 with the non-perturbing
// '-server -agentlib:hprof=cpu=samples,depth=10' flags, then study the trace log and correlate its hotspot
// trailer with its call stack headers (see hprof tracing ).
type Index struct {
	fields            *treemap.Map[string, *info]
	storeOffsets      bool
	storePayloads     bool
	byteBlockPool     *bytesref.BlockPool
	intBlockPool      *ints.BlockPool
	postingsWriter    *ints.SliceWriter
	payloadsBytesRefs *bytesref.Array //non null only when storePayloads
	frozen            bool
	normSimilarity    index2.Similarity
	defaultFieldType  *document.FieldType
}

func NewIndex(options ...Option) (*Index, error) {
	opt := &option{
		storeOffsets:   false,
		storePayloads:  false,
		maxReusedBytes: 0,
	}
	for _, fn := range options {
		fn(opt)
	}

	return newIndex(opt.storeOffsets, opt.storePayloads, opt.maxReusedBytes)
}

type option struct {
	storeOffsets   bool
	storePayloads  bool
	maxReusedBytes int64
}

type Option func(*option)

func WithStoreOffsets(storeOffsets bool) Option {
	return func(o *option) {
		o.storeOffsets = storeOffsets
	}
}

func WithStorePayloads(storePayloads bool) Option {
	return func(o *option) {
		o.storePayloads = storePayloads
	}
}

func WithMaxReusedBytes(maxReusedBytes int64) Option {
	return func(o *option) {
		o.maxReusedBytes = maxReusedBytes
	}
}

func NewFromDocument(doc *document.Document, analyzer analysis.Analyzer, options ...Option) (*Index, error) {
	opt := &option{
		storeOffsets:   false,
		storePayloads:  false,
		maxReusedBytes: 0,
	}
	for _, fn := range options {
		fn(opt)
	}

	newIdx, err := newIndex(opt.storeOffsets, opt.storePayloads, opt.maxReusedBytes)
	if err != nil {
		return nil, err
	}

	for _, field := range doc.Fields() {
		if err = newIdx.AddIndexAbleField(field, analyzer); err != nil {
			return nil, err
		}
	}

	return newIdx, nil
}

func newIndex(storeOffsets, storePayloads bool, maxReusedBytes int64) (*Index, error) {
	similarity, err := search.NewBM25Similarity()
	if err != nil {
		return nil, err
	}

	newIdx := Index{
		fields:           treemap.New[string, *info](),
		storeOffsets:     storeOffsets,
		storePayloads:    storePayloads,
		frozen:           false,
		normSimilarity:   similarity,
		defaultFieldType: document.NewFieldType(),
	}

	options := document.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS
	if !storeOffsets {
		options = document.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS
	}
	if err = newIdx.defaultFieldType.SetIndexOptions(options); err != nil {
		return nil, err
	}

	maxBufferedByteBlocks := (int)((maxReusedBytes / 2) / bytesref.BYTE_BLOCK_SIZE)
	maxBufferedIntBlocks := (int(maxReusedBytes) - (maxBufferedByteBlocks * bytesref.BYTE_BLOCK_SIZE)) / (ints.INT_BLOCK_SIZE * 4)

	allocator := bytesref.GetAllocatorBuilder().NewRecyclingByteBlock(bytesref.BYTE_BLOCK_SIZE, maxBufferedByteBlocks)
	newIdx.byteBlockPool = bytesref.NewBlockPool(allocator)

	intsAllocator := ints.NewRecyclingIntBlockAllocator(ints.INT_BLOCK_SIZE, maxBufferedIntBlocks)
	newIdx.intBlockPool = ints.NewBlockPool(intsAllocator)

	newIdx.postingsWriter = ints.NewSliceWriter(newIdx.intBlockPool)

	return &newIdx, nil
}

func (r *Index) NewIndexReader(fields *treemap.Map[string, *info]) *IndexReader {
	fieldInfosArr := make([]*document.FieldInfo, fields.Size())
	i := 0
	it := fields.Iterator()

	for it.Next() {
		fInfo := it.Value()
		fInfo.prepareDocValuesAndPointValues()
		fieldInfosArr[i] = fInfo.fieldInfo
		i++
	}
	return newIndexReader(r.newFields(fields), index.NewFieldInfos(fieldInfosArr), fields)
}

// AddIndexAbleField Adds a lucene IndexableField to the Index using the provided analyzer. Also stores doc
// values based on IndexableFieldType.docValuesType() if set.
// Params: field – the field to add
// analyzer – the analyzer to use for term analysis
// TODO: 完善代码
func (r *Index) AddIndexAbleField(field document.IndexableField, analyzer analysis.Analyzer) error {
	fInfo, err := r.getInfo(field.Name(), field.FieldType())
	if err != nil {
		return err
	}

	offsetGap, positionIncrementGap := 0, 0
	var tokenStream analysis.TokenStream
	if analyzer != nil {
		offsetGap = analyzer.GetOffsetGap(field.Name())
		tokenStream, err = field.TokenStream(analyzer, nil)
		if err != nil {
			return err
		}
		positionIncrementGap = analyzer.GetPositionIncrementGap(field.Name())
	} else {
		offsetGap = 1
		tokenStream, err = field.TokenStream(nil, nil)
		if err != nil {
			return err
		}
		positionIncrementGap = 0
	}

	if tokenStream != nil {
		if err = r.storeTerms(fInfo, tokenStream, positionIncrementGap, offsetGap); err != nil {
			return err
		}
	}

	docValuesType := field.FieldType().DocValuesType()

	switch docValuesType {
	case document.DOC_VALUES_TYPE_NONE:
		//if err = r.storeDocValues(fInfo, docValuesType, nil); err != nil {
		//	return err
		//}
	case document.DOC_VALUES_TYPE_BINARY, document.DOC_VALUES_TYPE_SORTED,
		document.DOC_VALUES_TYPE_SORTED_SET, document.DOC_VALUES_TYPE_NUMERIC,
		document.DOC_VALUES_TYPE_SORTED_NUMERIC:
		if err = r.storeDocValues(fInfo, docValuesType, field); err != nil {
			return err
		}
	default:
		return errors.New("unknown doc values types")
	}

	if field.FieldType().PointIndexDimensionCount() > 0 {
		bytes, err := document.Bytes(field.Get())
		if err != nil {
			return err
		}

		if err = r.storePointValues(fInfo, bytes); err != nil {
			return err
		}
	}

	return nil
}

// Search Convenience method that efficiently returns the relevance score by matching this index against the
// given Lucene query expression.
// Params: query – an arbitrary Lucene query to run against this index
// Returns: the relevance score of the matchmaking; A number in the range [0.0 .. 1.0], with 0.0 indicating
//
//	no match. The higher the number the better the match.
func (r *Index) Search(query search2.Query) float64 {
	if query == nil {
		return 0
	}

	searcher := r.CreateSearcher()

	scores := make([]float64, 1)
	collector := newSimpleCollector(scores)
	err := searcher.Search(query, collector)
	if err != nil {
		return 0
	}
	score := scores[0]
	return score
}

type addFieldOption struct {
	positionIncrementGap int
	offsetGap            int
}

type AddFieldOption func(*addFieldOption)

func WithPositionIncrementGap(positionIncrementGap int) AddFieldOption {
	return func(fieldOption *addFieldOption) {
		fieldOption.positionIncrementGap = positionIncrementGap
	}
}

func WithOffsetGap(offsetGap int) AddFieldOption {
	return func(fieldOption *addFieldOption) {
		fieldOption.offsetGap = offsetGap
	}
}

// AddField Iterates over the given token stream and adds the resulting terms to the index;
// Equivalent to adding a tokenized, indexed, termVectorStored, unstored,
// Lucene org.apache.lucene.document.Field. Finally closes the token stream.
// Note that untokenized keywords can be added with this method via keywordTokenStream(Collection),
// the Lucene KeywordTokenizer or similar utilities.
func (r *Index) AddField(fieldName string, tokenStream analysis.TokenStream, options ...AddFieldOption) error {
	opt := &addFieldOption{
		positionIncrementGap: 0,
		offsetGap:            1,
	}
	for _, fn := range options {
		fn(opt)
	}

	fInfo, err := r.getInfo(fieldName, r.defaultFieldType)
	if err != nil {
		return err
	}
	return r.storeTerms(fInfo, tokenStream, opt.positionIncrementGap, opt.offsetGap)
}

func (r *Index) AddFieldString(fieldName string, text string, analyzer analysis.Analyzer) error {
	stream, err := analyzer.GetTokenStreamFromText(fieldName, text)
	if err != nil {
		return err
	}
	fInfo, err := r.getInfo(fieldName, r.defaultFieldType)
	if err != nil {
		return err
	}

	return r.storeTerms(fInfo, stream, analyzer.GetPositionIncrementGap(fieldName),
		analyzer.GetOffsetGap(fieldName))
}

// SetSimilarity Set the Similarity to be used for calculating field norms
func (r *Index) SetSimilarity(similarity index2.Similarity) error {
	if r.frozen {
		return errors.New("cannot set Similarity when MemoryIndex is frozen")
	}

	if reflect.DeepEqual(r.normSimilarity, similarity) {
		return nil
	}

	r.fields.Each(func(key string, value *info) {
		value.norm = nil
	})

	return nil
}

func (r *Index) CreateSearcher() search2.IndexSearcher {
	reader := r.NewIndexReader(r.fields)
	searcher, _ := search.NewIndexSearcher(reader)
	searcher.SetSimilarity(r.normSimilarity)
	searcher.SetQueryCache(nil)
	return searcher
}

// Freeze Prepares the Index for querying in a non-lazy way.
// After calling this you can query the Index from multiple threads, but you cannot subsequently add new data.
func (r *Index) Freeze() {
	r.frozen = true
	r.fields.Each(func(key string, value *info) {
		value.freeze()
	})
}

// Reset Resets the MemoryIndex to its initial state and recycles all internal buffers.
func (r *Index) Reset() error {
	panic("TODO")
}
