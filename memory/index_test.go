package memory

import (
	"testing"

	"github.com/geange/lucene-go/core/analysis"
	"github.com/geange/lucene-go/core/analysis/standard"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/search"
	"github.com/stretchr/testify/assert"
)

func TestNewMemoryIndex(t *testing.T) {
	set := analysis.NewCharArraySet()
	set.Add(" ")
	set.Add("\n")
	set.Add("\t")
	analyzer := standard.NewAnalyzer(set)

	doc := document.NewDocument()
	doc.Add(document.NewTextField("name", "name1", false))
	doc.Add(document.NewTextField("address", "address1", false))
	doc.Add(document.NewTextField("other", "other1", false))

	memIndex, err := NewIndex()
	assert.Nil(t, err)
	err = memIndex.AddIndexAbleField(document.NewTextField("name", "name1", false), analyzer)
	assert.Nil(t, err)
}

func TestMemoryIndex(t *testing.T) {
	memIndex, err := NewIndex()
	if err != nil {
		panic(err)
	}

	set := analysis.NewCharArraySet()
	set.Add(" ")
	set.Add("\n")
	set.Add("\t")

	analyzer := standard.NewAnalyzer(set)

	err = memIndex.AddIndexAbleField(document.NewTextField("f1", "some text", false), analyzer)
	if err != nil {
		panic(err)
	}

	score := memIndex.Search(search.NewTermQuery(index.NewTerm("f1", []byte("text"))))
	assert.InDelta(t, 0.13076457, score, 0.00000001)

	score1 := memIndex.Search(search.NewTermQuery(index.NewTerm("f1", []byte("some"))))
	assert.InDelta(t, 0.13076457, score1, 0.00000001)

	score2 := memIndex.Search(search.NewTermQuery(index.NewTerm("f1", []byte("some text"))))
	assert.InDelta(t, 0, score2, 0.00000001)
}

func TestSeekByTermOrd(t *testing.T) {
	fieldName := "text"

	set := analysis.NewCharArraySet()
	set.Add(" ")
	set.Add("\n")
	set.Add("\t")
	analyzer := standard.NewAnalyzer(set)

	memIndex, err := NewIndex()
	assert.Nil(t, err)

	err = memIndex.AddFieldString(fieldName, "la la", analyzer)
	assert.Nil(t, err)
	err = memIndex.AddFieldString(fieldName, "foo bar foo bar foo", analyzer)
	assert.Nil(t, err)

	memIndexReader := memIndex.CreateSearcher().GetIndexReader()

	memIndexReader.GetRefCount()
}

func checkReader(t *testing.T, reader index.IndexReader) {
	leaves, err := reader.Leaves()
	assert.Nil(t, err)

	for _, ctx := range leaves {
		checkReaderDoSlowChecks(t, ctx.Reader(), true)
	}
}

func checkReaderDoSlowChecks(t *testing.T, reader index.IndexReader, doSlowChecks bool) {

}
