package main

import (
	"fmt"
	"github.com/geange/lucene-go/core/analysis"
	"github.com/geange/lucene-go/core/analysis/standard"
	"github.com/geange/lucene-go/core/document"
	index2 "github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/search"
	"github.com/geange/lucene-go/memory"
)

func main() {
	index, err := memory.NewNewMemoryIndexDefault()
	if err != nil {
		panic(err)
	}

	set := analysis.NewCharArraySet()
	set.Add(" ")
	set.Add("\n")
	set.Add("\t")

	analyzer := standard.NewAnalyzer(set)

	err = index.AddField(document.NewTextFieldByString("f1", "some text", false), analyzer)
	if err != nil {
		panic(err)
	}

	count := index.Search(search.NewTermQuery(index2.NewTerm("f1", []byte("some text"))))
	fmt.Println(count)
}
