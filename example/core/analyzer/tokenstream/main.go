package main

import (
	"fmt"
	"github.com/geange/lucene-go/core/analysis"
	"github.com/geange/lucene-go/core/analysis/standard"
)

func main() {
	set := analysis.NewCharArraySet()
	set.Add(" ")
	set.Add("\n")
	set.Add("\t")

	analyzer := standard.NewAnalyzer(set)

	imp := analysis.NewAnalyzerImp(analyzer)

	tokenstream, err := imp.TokenStreamByString("xxxx", "aaaa BBBFFDs cccc dddd")
	if err != nil {
		panic(err)
	}

	tokenstream.IncrementToken()
	fmt.Println(string(tokenstream.AttributeSource().CharTerm().Buffer()))

	tokenstream.IncrementToken()
	fmt.Println(string(tokenstream.AttributeSource().CharTerm().Buffer()))

	tokenstream.IncrementToken()
	fmt.Println(string(tokenstream.AttributeSource().CharTerm().Buffer()))
}
