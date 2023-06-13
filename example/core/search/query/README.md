# README.md

## Use Java

write doc to disk

```java
package org.example;

import org.apache.lucene.analysis.Analyzer;
import org.apache.lucene.analysis.standard.StandardAnalyzer;
import org.apache.lucene.codecs.Codec;
import org.apache.lucene.codecs.simpletext.SimpleTextCodec;
import org.apache.lucene.document.Document;
import org.apache.lucene.document.Field;
import org.apache.lucene.document.TextField;
import org.apache.lucene.index.*;
import org.apache.lucene.search.IndexSearcher;
import org.apache.lucene.search.Query;
import org.apache.lucene.search.ScoreDoc;
import org.apache.lucene.search.TermQuery;
import org.apache.lucene.search.similarities.BM25Similarity;
import org.apache.lucene.store.Directory;
import org.apache.lucene.store.FSDirectory;

import java.io.IOException;
import java.nio.file.Paths;

public class SearchTerm {
    public static void main(String[] args) throws IOException {
        String indexPath = "data";
        Directory dir = FSDirectory.open(Paths.get(indexPath));
        Analyzer analyzer = new StandardAnalyzer();

        IndexWriterConfig iwc = new IndexWriterConfig(analyzer);
        iwc.setOpenMode(IndexWriterConfig.OpenMode.CREATE);

        Codec codec = new SimpleTextCodec();
        iwc.setCodec(codec);

        BM25Similarity similarity = new BM25Similarity();
        iwc.setSimilarity(similarity);

        IndexWriter indexWriter = new IndexWriter(dir, iwc);

        Document doc;

        doc = new Document();
        doc.add(new TextField("content", "h", Field.Store.YES));
        doc.add(new TextField("author", "author1", Field.Store.YES));
        indexWriter.addDocument(doc);
        // 1
        doc = new Document();
        doc.add(new TextField("content", "b", Field.Store.YES));
        doc.add(new TextField("author", "author2", Field.Store.YES));
        indexWriter.addDocument(doc);
        // 2
        doc = new Document();
        doc.add(new TextField("content", "a c", Field.Store.YES));
        doc.add(new TextField("author", "author3", Field.Store.YES));

        indexWriter.addDocument(doc);
        // 3
        doc = new Document();
        doc.add(new TextField("content", "a c e", Field.Store.YES));
        doc.add(new TextField("author", "author4", Field.Store.YES));
        indexWriter.addDocument(doc);
        // 4
        doc = new Document();
        doc.add(new TextField("content", "h", Field.Store.YES));
        doc.add(new TextField("author", "author5", Field.Store.YES));
        indexWriter.addDocument(doc);
        // 5
        doc = new Document();
        doc.add(new TextField("content", "c e", Field.Store.YES));
        doc.add(new TextField("author", "author6", Field.Store.YES));
        indexWriter.addDocument(doc);
        // 6
        doc = new Document();
        doc.add(new TextField("content", "c a e", Field.Store.YES));
        doc.add(new TextField("author", "author7", Field.Store.YES));
        indexWriter.addDocument(doc);
        // 7
        doc = new Document();
        doc.add(new TextField("content", "f", Field.Store.YES));
        doc.add(new TextField("author", "author8", Field.Store.YES));
        indexWriter.addDocument(doc);
        // 8
        doc = new Document();
        doc.add(new TextField("content", "b c d e c e", Field.Store.YES));
        doc.add(new TextField("author", "author9", Field.Store.YES));
        indexWriter.addDocument(doc);
        // 9
        doc = new Document();
        doc.add(new TextField("content", "a c e a b c", Field.Store.YES));
        doc.add(new TextField("author", "author10", Field.Store.YES));
        indexWriter.addDocument(doc);
        indexWriter.commit();
        indexWriter.flush();
        indexWriter.close();
    }
}

```

## Search 

```go
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

```