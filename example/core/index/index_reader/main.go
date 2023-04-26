package main

import (
	"fmt"

	"github.com/geange/lucene-go/codecs/simpletext"
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

	reader, err := index.DirectoryReaderOpen(writer)
	if err != nil {
		panic(err)
	}

	//maxDoc := reader.MaxDoc()
	//
	//for i := 0; i < maxDoc; i++ {
	//	doc, err := reader.DocumentV2(i, map[string]struct{}{"sequence": {}})
	//	if err != nil {
	//		fmt.Println(err)
	//		continue
	//	}
	//	if doc == nil {
	//		continue
	//	}
	//	terms, err := doc.GetField("sequence")
	//	if err != nil {
	//		fmt.Println(err)
	//		continue
	//	}
	//	fmt.Println(terms.Name(), terms.Value())
	//}

	searcher, err := search.NewIndexSearcher(reader)
	if err != nil {
		panic(err)
	}
	topDocs, err := searcher.SearchTopN(search.NewMatchAllDocsQuery(), 100)
	if err != nil {
		panic(err)
	}

	result := topDocs.GetScoreDocs()
	for _, scoreDoc := range result {
		docID := scoreDoc.GetDoc()
		document, err := reader.Document(docID)
		if err != nil {
			panic(err)
		}
		value, err := document.Get("sequence")
		if err != nil {
			return
		}
		fmt.Printf("段内排序后的文档号: %d  VS 段内排序前的文档: %s\n",
			scoreDoc.GetDoc(), value)
	}

	//searchSortField1 := index.NewSortedSetSortFieldV1("sort0", true, index.MAX)
	//searchSortField2 := index.NewSortedSetSortFieldV1("sort1", true, index.MIN)
	//searchSortFields := []index.SortField{searchSortField1, searchSortField2}
	//searchSort := index.NewSort(searchSortFields)
	////
	//search.

	//docs := reader.NumDocs()
	//
	//searcher.Search(search.NewNamedMatches(), 100, searchSort).coreDocs

	//fmt.Println(docs)
	//
	//{
	//	doc := document.NewDocument()
	//	doc.Add(document.NewStoredFieldAny("a", 74, document.STORED_ONLY))
	//	doc.Add(document.NewStoredFieldAny("a1", 86, document.STORED_ONLY))
	//	doc.Add(document.NewStoredFieldAny("a2", 1237, document.STORED_ONLY))
	//	docID, err := writer.AddDocument(doc)
	//	if err != nil {
	//		panic(err)
	//	}
	//	fmt.Println(docID)
	//}
	//
	//{
	//	doc := document.NewDocument()
	//	doc.Add(document.NewStoredFieldAny("a", 123, document.STORED_ONLY))
	//	doc.Add(document.NewStoredFieldAny("a1", 123, document.STORED_ONLY))
	//	doc.Add(document.NewStoredFieldAny("a2", 789, document.STORED_ONLY))
	//
	//	docID, err := writer.AddDocument(doc)
	//	if err != nil {
	//		panic(err)
	//	}
	//	fmt.Println(docID)
	//}
	//
	//{
	//	doc := document.NewDocument()
	//	doc.Add(document.NewStoredFieldAny("a", 741, document.STORED_ONLY))
	//	doc.Add(document.NewStoredFieldAny("a1", 861, document.STORED_ONLY))
	//	doc.Add(document.NewStoredFieldAny("a2", 12137, document.STORED_ONLY))
	//	docID, err := writer.AddDocument(doc)
	//	if err != nil {
	//		panic(err)
	//	}
	//	fmt.Println(docID)
	//}
}
