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
