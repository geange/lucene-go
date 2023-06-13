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

	query := search.NewTermQuery(index.NewTerm("content", []byte("a")))

	topDocs, err := searcher.SearchTopN(query, 5)
	if err != nil {
		panic(err)
	}

	for i, doc := range topDocs.GetScoreDocs() {
		fmt.Printf("result%d: 文档%d\n", i, doc.GetDoc())
	}
}
