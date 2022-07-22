package main

import (
	"github.com/geange/lucene-go/core/analysis"
	"github.com/geange/lucene-go/core/analysis/standard"
	"github.com/geange/lucene-go/core/document"
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
	imp := analysis.NewAnalyzerImp(analyzer)

	err = index.AddField(document.NewTextFieldByString("name", "chenhualin", true), imp)
	if err != nil {
		panic(err)
	}
}
