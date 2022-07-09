package memory

import (
	"github.com/geange/lucene-go/core/analysis"
	"github.com/geange/lucene-go/core/analysis/standard"
	"github.com/geange/lucene-go/core/document"
	"testing"
)

func TestNewMemoryIndex(t *testing.T) {
	set := analysis.NewCharArraySet()
	set.Add(" ")
	set.Add("\n")
	set.Add("\t")
	newAnalyzer := standard.NewAnalyzer(set)
	analyzer := analysis.NewAnalyzerImp(newAnalyzer)

	doc := document.NewDocument()
	doc.Add(document.NewTextFieldByString("name", "name1", false))
	doc.Add(document.NewTextFieldByString("address", "address1", false))
	doc.Add(document.NewTextFieldByString("other", "other1", false))

	index, err := NewNewMemoryIndexDefault()
	if err != nil {
		return
	}
	index.AddField(document.NewTextFieldByString("name", "name1", false), analyzer)
}
