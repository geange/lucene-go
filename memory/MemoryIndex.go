package memory

import (
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/search/similarities"
	"go.uber.org/atomic"
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
// zero or more index terms (aka words) on addField(), according to the policy implemented by an Analyzer.
// For example, Lucene analyzers can split on whitespace, normalize to lower case for case insensitivity,
// ignore common terms with little discriminatory value such as "he", "in", "and" (stop words), reduce the terms
// to their natural linguistic root form such as "fishing" being reduced to "fish" (stemming), resolve
// synonyms/inflexions/thesauri (upon indexing and/or querying), etc. For details, see Lucene Analyzer Intro.
// Arbitrary Lucene queries can be run against this class - see Lucene Query Syntax as well as Query Parser Rules.
// Note that a Lucene query selects on the field names and associated (indexed) tokenized terms, not on the
// original fulltext(s) - the latter are not stored but rather thrown away immediately after tokenization.
// For some interesting background information on search technology, see Bob Wyman's Prospective Search,
// Jim Gray's A Call to Arms - Custom subscriptions, and Tim Bray's On Search, the Series.
// Example Usage
//   Analyzer analyzer = new SimpleAnalyzer(version);
//   MemoryIndex index = new MemoryIndex();
//   index.addField("content", "Readings about Salmons and other select Alaska fishing Manuals", analyzer);
//   index.addField("author", "Tales of James", analyzer);
//   QueryParser parser = new QueryParser(version, "content", analyzer);
//   float score = index.search(parser.parse("+author:james +salmon~ +fish* manual~"));
//   if (score > 0.0f) {
//       System.out.println("it's a match");
//   } else {
//       System.out.println("no match found");
//   }
//   System.out.println("indexData=" + index.toString());
//
// Example XQuery Usage
//   (: An XQuery that finds all books authored by James that have something to do
//   with "salmon fishing manuals", sorted by relevance :)
//   declare namespace lucene = "java:nux.xom.pool.FullTextUtil";
//   declare variable $query := "+salmon~ +fish* manual~"; (: any arbitrary Lucene query can go here :)
//
//   for $book in /books/book[author="James" and lucene:match(abstract, $query) > 0.0]
//   let $score := lucene:match($book/abstract, $query)
//   order by $score descending
//   return $book
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
	fields map[string]*Info

	storeOffsets  bool
	storePayloads bool

	bytesUsed        atomic.Int64
	frozen           bool
	normSimilarity   similarities.Similarity
	defaultFieldType *document.FieldType
}

type Info struct {
	fieldInfo *index.FieldInfo
	norm      int64
}
