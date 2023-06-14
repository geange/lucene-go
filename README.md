# lucene-go

> A Go port of Apache Lucene 8.11.2

I am learning the code of lucene, so I am using golang to re-implement the version of lucene in golang.
At present, this is only a learning project. I try to implement the contents of lucene/core package. 
Due to the complexity of Lucene, the readability of the project code needs to be improved.

The current project is not fully operational, and there is still a lot of work to be improved.
Only a small number of class libraries can run independently, but there may be some bugs.

## example

### IndexWriter

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

### IndexSearch

use indexSearch to get TopN docs

> IndexSearch has not been developed yet～

[more detail](example/core/search/query/README.md)


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

output

```shell
result0: 文档2
result1: 文档9
result2: 文档3
result3: 文档6
```

### memory

```go
package main

import (
	"fmt"
	"github.com/geange/lucene-go/core/analysis"
	"github.com/geange/lucene-go/core/analysis/standard"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/search"
	"github.com/geange/lucene-go/memory"
)

func main() {
	mi, err := memory.NewNewMemoryIndexDefault()
	if err != nil {
		panic(err)
	}

	set := analysis.NewCharArraySet()
	set.Add(" ")
	set.Add("\n")
	set.Add("\t")

	analyzer := standard.NewStandardAnalyzer(set)

	err = mi.AddField(document.NewTextFieldByString("f1", "some text", false), analyzer)
	if err != nil {
		panic(err)
	}

	count := mi.Search(search.NewTermQuery(index.NewTerm("f1", []byte("text"))))
	fmt.Println(count)
}
```