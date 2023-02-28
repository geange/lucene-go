package main

import (
	"fmt"
	"github.com/geange/lucene-go/codecs/simpletext"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/search"
	"github.com/geange/lucene-go/core/store"
)

func main() {
	dir, err := store.NewNIOFSDirectory("data")
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

	doc := document.NewDocument()
	doc.Add(document.NewStringFieldByString("a", "123", true))
	doc.Add(document.NewStringFieldByString("a", "456", true))
	doc.Add(document.NewStringFieldByString("a", "789", true))

	docID, err := writer.AddDocument(doc)
	if err != nil {
		panic(err)
	}
	fmt.Println(docID)
}
