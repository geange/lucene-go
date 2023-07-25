# lucene-go

**[中文](README-zh_CN.md)**

## About

> A Go port of Apache Lucene

The original intention of starting this project was because the recipe story of 'Elasticsearch' was too sci-fi. After
understanding the basic knowledge related to search, I hastily started the development of the project.

I originally hoped to achieve a Go version of ES, but now I am still a bit far from this goal.
More importantly, it is necessary to improve the Lucene go project as soon as possible to achieve a fully usable state.
This available state includes but is not limited to the following:

* Improve the code. The original code was to translate Java into Go, but there are many shortcomings. The next goal is
  to make the code more like what Gopher wrote 🐶
* Improving unit testing and single testing is the most ideal solution to ensure code quality
* Improve use cases. The quality of use cases is still relatively low, just some simple cases that I personally used for
  local testing
* Improve the documentation, which will be carried out together with the process of improving the code, making it easier
  for users to obtain the content they want (after all, the Lucene library has a very large amount of code)

The goal of the project has undergone several twists and turns during the development process, encountering far greater
difficulties than expected, language differences, and a lack of theoretical knowledge. After a year of development, the
following major modules have been gradually completed:

* core/store: Lucene's storage module is mainly responsible for data serialization processing
* core/document: Define some search related data structures for Lucene
* core/index: The implementation of Lucene index is also the main package exposed to the public
* core/search: Mainly including the implementation of query (used for data retrieval in indexes)
* memory: Implemented a memory based search engine, a simplified version of Lucene
* util/fst: Implementation of FST (Lucene's important data structure)
* util/automaton: Implementation of Automata
* codes: serialization, currently only supports the format of simpleText (using plain text information to record index
  data)

> It should be noted that the current project is not perfect! Do not use for any project~

## Try

> go1.18+

### Example

[More Example](https://github.com/geange/lucene-go-example)

Using `IndexWriter`

```go
package main

import (
	"context"
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
	defer func() {
		err := writer.Commit(context.Background())
		if err != nil {
			fmt.Println(err)
		}
	}()

	{
		doc := document.NewDocument()
		doc.Add(document.NewStoredFieldAny("a", 74, document.STORED_ONLY))
		doc.Add(document.NewStoredFieldAny("a1", 86, document.STORED_ONLY))
		doc.Add(document.NewStoredFieldAny("a2", 1237, document.STORED_ONLY))
		docID, err := writer.AddDocument(doc)
		if err != nil {
			panic(err)
		}
		fmt.Println(docID)
	}

	{
		doc := document.NewDocument()
		doc.Add(document.NewStoredFieldAny("a", 123, document.STORED_ONLY))
		doc.Add(document.NewStoredFieldAny("a1", 123, document.STORED_ONLY))
		doc.Add(document.NewStoredFieldAny("a2", 789, document.STORED_ONLY))

		docID, err := writer.AddDocument(doc)
		if err != nil {
			panic(err)
		}
		fmt.Println(docID)
	}

	{
		doc := document.NewDocument()
		doc.Add(document.NewStoredFieldAny("a", 741, document.STORED_ONLY))
		doc.Add(document.NewStoredFieldAny("a1", 861, document.STORED_ONLY))
		doc.Add(document.NewStoredFieldAny("a2", 12137, document.STORED_ONLY))
		docID, err := writer.AddDocument(doc)
		if err != nil {
			panic(err)
		}
		fmt.Println(docID)
	}
}

```
