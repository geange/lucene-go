package main

import (
	"fmt"

	"github.com/geange/lucene-go/codecs/simpletext"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/search"
	"github.com/geange/lucene-go/core/store"
)

func main() {
	dir, err := store.NewNIOFSDirectory("./data")
	if err != nil {
		panic(err)
	}

	codec := simpletext.NewSimpleTextCodec()
	similarity := search.NewCastBM25Similarity()

	config := index.NewIndexWriterConfig(codec, similarity)

	writer, err := index.NewIndexWriter(dir, config)
	if err != nil {
		panic(err)
	}

	reader, err := index.DirectoryReaderOpen(writer)
	if err != nil {
		panic(err)
	}

	searcher, err := search.NewIndexSearcher(reader)
	if err != nil {
		panic(err)
	}

	q1 := search.NewTermQuery(index.NewTerm("content", []byte("a")))
	q2 := search.NewTermQuery(index.NewTerm("content", []byte("c")))
	q3 := search.NewTermQuery(index.NewTerm("content", []byte("e")))
	q4 := search.NewTermQuery(index.NewTerm("author", []byte("author4")))

	builder := search.NewBooleanQueryBuilder()
	builder.AddQuery(q1, search.OccurMust)
	builder.AddQuery(q2, search.OccurMust)
	builder.AddQuery(q3, search.OccurMust)
	builder.AddQuery(q4, search.OccurMust)
	query, err := builder.Build()
	if err != nil {
		panic(err)
	}

	topDocs, err := searcher.SearchTopN(query, 5)
	if err != nil {
		panic(err)
	}

	for i, doc := range topDocs.GetScoreDocs() {
		fmt.Printf("result%d: 文档%d\n", i, doc.GetDoc())
	}
}
