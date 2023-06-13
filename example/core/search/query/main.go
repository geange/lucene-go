package main

import (
	"fmt"

	"github.com/geange/lucene-go/codecs/simpletext"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/search"
	"github.com/geange/lucene-go/core/store"
)

func main() {

	//q1 := search.NewTermQuery(index.NewTerm("content", []byte("h")))
	//q2 := search.NewTermQuery(index.NewTerm("content", []byte("f")))
	//
	//builder := search.NewBooleanQueryBuilder()
	//builder.AddQuery(q1, search.SHOULD)
	//builder.AddQuery(q2, search.SHOULD)
	//builder.SetMinimumNumberShouldMatch(1)
	//
	//query, err := builder.Build()
	//if err != nil {
	//	panic(err)
	//}

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

	query0 := search.NewTermQuery(index.NewTerm("content", []byte("a")))
	query1 := search.NewTermQuery(index.NewTerm("content", []byte("e")))
	builder := search.NewBooleanQueryBuilder()
	builder.AddQuery(query0, search.MUST)
	builder.AddQuery(query1, search.MUST)
	query, err := builder.Build()
	if err != nil {
		return
	}

	topDocs, err := searcher.SearchTopN(query, 5)
	if err != nil {
		panic(err)
	}

	for i, doc := range topDocs.GetScoreDocs() {
		fmt.Printf("result%d: 文档%d\n", i, doc.GetDoc())
	}

	//q2 := search.NewTermQuery(index.NewTerm("content", []byte("f")))
	//
	//builder := search.NewBooleanQueryBuilder()
	//builder.AddQuery(q1, search.SHOULD)
	//builder.AddQuery(q2, search.SHOULD)
	//builder.SetMinimumNumberShouldMatch(1)
	//
	//query, err := builder.Build()
	//if err != nil {
	//	panic(err)
	//}

}
