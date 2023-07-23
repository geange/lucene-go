package memory

import (
	"github.com/geange/lucene-go/core/analysis"
	"github.com/geange/lucene-go/core/analysis/standard"
	"github.com/geange/lucene-go/core/document"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewMemoryIndex(t *testing.T) {
	set := analysis.NewCharArraySet()
	set.Add(" ")
	set.Add("\n")
	set.Add("\t")
	newAnalyzer := standard.NewStandardAnalyzer(set)
	analyzer := analysis.NewAnalyzerImp(newAnalyzer)

	doc := document.NewDocument()
	doc.Add(document.NewTextField("name", "name1", false))
	doc.Add(document.NewTextField("address", "address1", false))
	doc.Add(document.NewTextField("other", "other1", false))

	index, err := NewNewMemoryIndexDefault()
	assert.Nil(t, err)
	err = index.AddField(document.NewTextField("name", "name1", false), analyzer)
	assert.Nil(t, err)
}
