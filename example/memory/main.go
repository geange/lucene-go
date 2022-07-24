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

	fmt.Println("---------")

	index, err := memory.NewNewMemoryIndexDefault()
	if err != nil {
		panic(err)
	}

	set := analysis.NewCharArraySet()
	set.Add(" ")
	set.Add("\n")
	set.Add("\t")

	analyzer := standard.NewAnalyzer(set)

	err = index.AddField(document.NewTextFieldByString("name", "cnn abcd", false), analyzer)
	if err != nil {
		panic(err)
	}

	count := index.Search(search.NewTermQuery(index2.NewTerm("name", []byte("cnn"))))
	fmt.Println(count)
}
