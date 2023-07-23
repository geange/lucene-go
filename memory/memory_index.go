package memory

import (
	"errors"
	"fmt"
	"github.com/geange/gods-generic/maps/treemap"
	"github.com/geange/lucene-go/core/analysis"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/search"
	"github.com/geange/lucene-go/core/tokenattr"
	"github.com/geange/lucene-go/core/util"
	"go.uber.org/atomic"
	"reflect"
)

// High-performance single-document main memory Apache Lucene fulltext search index.

// MemoryIndex High-performance single-document main memory Apache Lucene fulltext search index.
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
// Each instance can hold at most one Lucene "document", with a document containing zero or more "fields",
// each field having a name and a fulltext value. The fulltext value is tokenized (split and transformed) into
// zero or more index Terms (aka words) on addField(), according to the policy implemented by an Analyzer.
// For example, Lucene analyzers can split on whitespace, normalize to lower case for case insensitivity,
// ignore common Terms with little discriminatory value such as "he", "in", "and" (stop words), reduce the Terms
// to their natural linguistic root form such as "fishing" being reduced to "fish" (stemming), resolve
// synonyms/inflexions/thesauri (upon indexing and/or querying), etc. For details, see Lucene Analyzer Intro.
// Arbitrary Lucene queries can be run against this class - see Lucene Query Syntax as well as Query Parser Rules.
// Note that a Lucene query selects on the field names and associated (indexed) tokenized Terms, not on the
// original fulltext(s) - the latter are not stored but rather thrown away immediately after tokenization.
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
// MemoryIndex is not normally thread-safe for adds or queries. However, queries are thread-safe after
// freeze() has been called.
// Performance Notes
// Internally there's a new data structure geared towards efficient indexing and searching, plus the necessary
// support code to seamlessly plug into the Lucene framework.
// This class performs very well for very small texts (e.g. 10 chars) as well as for large texts (e.g. 10 MB)
// and everything in between. Typically, it is about 10-100 times faster than RAMDirectory. Note that
// RAMDirectory has particularly large efficiency overheads for small to medium sized texts, both in time and
// space. Indexing a field with N tokens takes O(N) in the best case, and O(N logN) in the worst case.
// Memory consumption is probably larger than for RAMDirectory.
// Example throughput of many simple term queries over a single MemoryIndex: ~500000 queries/sec on a
// MacBook Pro, jdk 1.5.0_06, server VM. As always, your mileage may vary.
// If you're curious about the whereabouts of bottlenecks, run java 1.5 with the non-perturbing
// '-server -agentlib:hprof=cpu=samples,depth=10' flags, then study the trace log and correlate its hotspot
// trailer with its call stack headers (see hprof tracing ).
type MemoryIndex struct {
	fields *treemap.Map[string, *Info]

	storeOffsets  bool
	storePayloads bool

	byteBlockPool     *util.ByteBlockPool
	intBlockPool      *util.IntBlockPool
	postingsWriter    *util.SliceWriter
	payloadsBytesRefs *util.BytesRefArray //non null only when storePayloads

	bytesUsed        *atomic.Int64
	frozen           bool
	normSimilarity   index.Similarity
	defaultFieldType *document.FieldType
}

func NewNewMemoryIndexDefault() (*MemoryIndex, error) {
	return NewMemoryIndex(false, false, 0)
}

func NewMemoryIndex(storeOffsets, storePayloads bool, maxReusedBytes int64) (*MemoryIndex, error) {
	similarity, err := search.NewBM25Similarity()
	if err != nil {
		return nil, err
	}

	index := MemoryIndex{
		fields:           treemap.New[string, *Info](),
		storeOffsets:     storeOffsets,
		storePayloads:    storePayloads,
		bytesUsed:        atomic.NewInt64(0),
		frozen:           false,
		normSimilarity:   similarity,
		defaultFieldType: document.NewFieldType(),
	}

	options := document.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS
	if !storeOffsets {
		options = document.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS
	}
	if err = index.defaultFieldType.SetIndexOptions(options); err != nil {
		return nil, err
	}

	maxBufferedByteBlocks := (int)((maxReusedBytes / 2) / util.BYTE_BLOCK_SIZE)
	maxBufferedIntBlocks := (int(maxReusedBytes) - (maxBufferedByteBlocks * util.BYTE_BLOCK_SIZE)) / (util.INT_BLOCK_SIZE * 4)

	allocator := util.NewRecyclingByteBlockAllocator(util.BYTE_BLOCK_SIZE, maxBufferedByteBlocks)
	index.byteBlockPool = util.NewByteBlockPool(allocator)

	intsAllocator := util.NewRecyclingIntBlockAllocator(util.INT_BLOCK_SIZE, maxBufferedIntBlocks)
	index.intBlockPool = util.NewIntBlockPool(intsAllocator)

	index.postingsWriter = util.NewSliceWriter(index.intBlockPool)

	return &index, nil
}

func FromDocument(document *document.Document, analyzer analysis.Analyzer,
	storeOffsets, storePayloads bool, maxReusedBytes int64) (*MemoryIndex, error) {

	index, err := NewMemoryIndex(storeOffsets, storePayloads, maxReusedBytes)
	if err != nil {
		return nil, err
	}

	fn := document.Iterator()

	for {
		field := fn()
		if field == nil {
			break
		}

		err := index.AddField(field, analyzer)
		if err != nil {
			return nil, err
		}
	}
	return index, nil
}

// AddField Adds a lucene IndexableField to the MemoryIndex using the provided analyzer. Also stores doc
// values based on IndexableFieldType.docValuesType() if set.
// Params: field – the field to add
// analyzer – the analyzer to use for term analysis
func (m *MemoryIndex) AddField(field document.IndexableField, analyzer analysis.Analyzer) error {
	info, err := m.getInfo(field.Name(), field.FieldType())
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
		err := m.storeTerms(info, tokenStream, positionIncrementGap, offsetGap)
		if err != nil {
			return err
		}
	}

	docValuesType := field.FieldType().DocValuesType()

	switch docValuesType {
	case document.DOC_VALUES_TYPE_NONE:
		break
	case document.DOC_VALUES_TYPE_BINARY, document.DOC_VALUES_TYPE_SORTED, document.DOC_VALUES_TYPE_SORTED_SET:
		err := m.storeDocValues(info, docValuesType, field.Value())
		if err != nil {
			return err
		}
	case document.DOC_VALUES_TYPE_NUMERIC, document.DOC_VALUES_TYPE_SORTED_NUMERIC:
		err := m.storeDocValues(info, docValuesType, field.Value())
		if err != nil {
			return err
		}
	default:
		return errors.New("unknown doc values types")
	}

	if field.FieldType().PointIndexDimensionCount() > 0 {
		err := m.storePointValues(info, field.Value().([]byte))
		if err != nil {
			return err
		}
	}

	return nil
}

// SetSimilarity Set the Similarity to be used for calculating field norms
func (m *MemoryIndex) SetSimilarity(similarity index.Similarity) error {
	if m.frozen {
		return errors.New("cannot set Similarity when MemoryIndex is frozen")
	}

	if reflect.DeepEqual(m.normSimilarity, similarity) {
		return nil
	}

	m.fields.Each(func(key string, value *Info) {
		value.norm = nil
	})

	return nil
}

func (m *MemoryIndex) CreateSearcher() *search.IndexSearcher {
	reader := m.NewMemoryIndexReader(m.fields)
	searcher, _ := search.NewIndexSearcher(reader)
	searcher.SetSimilarity(m.normSimilarity)
	searcher.SetQueryCache(nil)
	return searcher
}

// Freeze Prepares the MemoryIndex for querying in a non-lazy way.
// After calling this you can query the MemoryIndex from multiple threads, but you cannot subsequently add new data.
func (m *MemoryIndex) Freeze() {
	m.frozen = true
	m.fields.Each(func(key string, value *Info) {
		value.freeze()
	})
}

// Search Convenience method that efficiently returns the relevance score by matching this index against the
// given Lucene query expression.
// Params: query – an arbitrary Lucene query to run against this index
// Returns: the relevance score of the matchmaking; A number in the range [0.0 .. 1.0], with 0.0 indicating
//
//	no match. The higher the number the better the match.
func (m *MemoryIndex) Search(query search.Query) float64 {
	if query == nil {
		return 0
	}

	searcher := m.CreateSearcher()

	scores := make([]float64, 1)
	collector := newSimpleCollector(scores)
	err := searcher.Search(query, collector)
	if err != nil {
		return 0
	}
	score := scores[0]
	return score
}

func (m *MemoryIndex) getInfo(fieldName string, fieldType document.IndexableFieldType) (*Info, error) {
	if m.frozen {
		return nil, errors.New("cannot call addField() when MemoryIndex is frozen")
	}

	if fieldName == "" {
		return nil, errors.New("fieldName must not be null")
	}

	var info *Info
	v, ok := m.fields.Get(fieldName)
	if !ok {
		info = m.NewInfo(m.createFieldInfo(fieldName, m.fields.Size(), fieldType), m.byteBlockPool)
		m.fields.Put(fieldName, info)
	} else {
		info = v
	}

	if !ok {
		info = m.NewInfo(m.createFieldInfo(fieldName, m.fields.Size(), fieldType), m.byteBlockPool)
		m.fields.Put(fieldName, info)
	}

	if fieldType.PointDimensionCount() != info.fieldInfo.GetPointDimensionCount() {
		if fieldType.PointDimensionCount() > 0 {
			err := info.fieldInfo.SetPointDimensions(
				fieldType.PointDimensionCount(),
				fieldType.PointIndexDimensionCount(),
				fieldType.PointNumBytes())
			if err != nil {
				return nil, err
			}
		}
	}

	if fieldType.DocValuesType() != info.fieldInfo.GetDocValuesType() {
		if fieldType.DocValuesType() != document.DOC_VALUES_TYPE_NONE {
			err := info.fieldInfo.SetDocValuesType(fieldType.DocValuesType())
			if err != nil {
				return nil, err
			}
		}
	}

	return info, nil
}

func (m *MemoryIndex) storeTerms(info *Info, tokenStream analysis.TokenStream, positionIncrementGap, offsetGap int) error {
	pos := -1
	offset := 0
	if info.numTokens > 0 {
		pos = info.lastPosition + positionIncrementGap
		offset = info.lastOffset + offsetGap
	}

	stream := tokenStream

	termAtt := stream.AttributeSource().CharTerm()
	posIncrAttribute := stream.AttributeSource().PositionIncrement()
	offsetAtt := stream.AttributeSource().Offset()
	//packedAttr := stream.AttributeSource().PackedTokenAttribute()
	//bytesAttr := stream.AttributeSource().BytesTerm()
	payloadAtt := stream.AttributeSource().Payload()

	err := stream.Reset()
	if err != nil {
		return err
	}

	for {
		ok, err := stream.IncrementToken()
		if err != nil {
			return err
		}
		if !ok {
			break
		}

		info.numTokens++
		posIncr := posIncrAttribute.GetPositionIncrement()
		if posIncr == 0 {
			info.numOverlapTokens++
		}

		pos += posIncr
		ord, err := info.terms.Add([]byte(string(termAtt.Buffer())))
		if err != nil {
			return err
		}
		if ord < 0 {
			ord = (-ord) - 1
			m.postingsWriter.Reset(info.sliceArray.end[ord])
		} else {
			info.sliceArray.start[ord] = m.postingsWriter.StartNewSlice()
		}
		info.sliceArray.freq[ord]++
		info.maxTermFrequency = max(info.maxTermFrequency, info.sliceArray.freq[ord])
		info.sumTotalTermFreq++
		m.postingsWriter.WriteInt(pos)
		if m.storeOffsets {
			m.postingsWriter.WriteInt(offsetAtt.StartOffset() + offset)
			m.postingsWriter.WriteInt(offsetAtt.EndOffset() + offset)
		}

		if m.storePayloads {
			payload := payloadAtt.(tokenattr.PayloadAttribute).GetPayload()
			pIndex := 0
			if payload == nil || len(payload) == 0 {
				pIndex = -1
			} else {
				pIndex = m.payloadsBytesRefs.Append(payload)
			}
			m.postingsWriter.WriteInt(pIndex)
		}
		info.sliceArray.end[ord] = m.postingsWriter.GetCurrentOffset()
	}

	err = stream.End()
	if err != nil {
		return err
	}

	if info.numTokens > 0 {
		info.lastPosition = pos
		info.lastOffset = offsetAtt.EndOffset() + offset
	}

	return nil
}

func (m *MemoryIndex) storeDocValues(info *Info, docValuesType document.DocValuesType, docValuesValue interface{}) error {
	fieldName := info.fieldInfo.Name()
	existingDocValuesType := info.fieldInfo.GetDocValuesType()
	if existingDocValuesType == document.DOC_VALUES_TYPE_NONE {
		info.fieldInfo = document.NewFieldInfo(info.fieldInfo.Name(), info.fieldInfo.Number(), info.fieldInfo.HasVectors(),
			info.fieldInfo.HasPayloads(), info.fieldInfo.HasPayloads(), info.fieldInfo.GetIndexOptions(), docValuesType,
			-1, info.fieldInfo.Attributes(), info.fieldInfo.GetPointDimensionCount(), info.fieldInfo.GetPointIndexDimensionCount(),
			info.fieldInfo.GetPointNumBytes(), info.fieldInfo.IsSoftDeletesField())
	} else if existingDocValuesType != docValuesType {
		return fmt.Errorf(
			`can't add ["%v"] doc values field ["%v"], because ["%v"] doc values field already exists`,
			docValuesType, fieldName, existingDocValuesType,
		)
	}
	return nil
}

func (m *MemoryIndex) createFieldInfo(fieldName string, ord int, fieldType document.IndexableFieldType) *document.FieldInfo {
	indexOptions := document.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS
	if m.storeOffsets {
		indexOptions = document.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS
	}

	return document.NewFieldInfo(fieldName, ord, fieldType.StoreTermVectors(), fieldType.OmitNorms(), m.storePayloads,
		indexOptions, fieldType.DocValuesType(), -1, map[string]string{},
		fieldType.PointDimensionCount(), fieldType.PointIndexDimensionCount(), fieldType.PointNumBytes(), false)
}

func (m *MemoryIndex) storePointValues(info *Info, pointValue []byte) error {
	if len(info.pointValues) == 0 {
		info.pointValues = make([][]byte, 4)
	}
	info.pointValues = append(info.pointValues, pointValue)
	return nil
}
