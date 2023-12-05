package memory

import (
	"testing"

	"github.com/geange/lucene-go/core/analysis"
	"github.com/geange/lucene-go/core/analysis/standard"
	"github.com/geange/lucene-go/core/document"
	"github.com/stretchr/testify/assert"
)

func Test_fields_Terms(t *testing.T) {
	set := analysis.NewCharArraySet()
	set.Add(" ")
	set.Add("\n")
	set.Add("\t")
	analyzer := standard.NewAnalyzer(set)

	memIndex, err := NewIndex()
	assert.Nil(t, err)
	err = memIndex.AddIndexAbleField(document.NewTextField("name", "k1 k2 k3", false), analyzer)
	//analyzer = standard.NewAnalyzer(set)
	err = memIndex.AddIndexAbleField(document.NewTextField("age", "k1 k2 k3 k4", false), analyzer)
	memIndex.Freeze()

	mFields := memIndex.newFields(memIndex.fields)

	{
		terms, err := mFields.Terms("name")
		assert.Nil(t, err)

		size, err := terms.Size()
		assert.Nil(t, err)
		assert.Equal(t, 3, size)
	}

	{
		terms, err := mFields.Terms("age")
		assert.Nil(t, err)

		size, err := terms.Size()
		assert.Nil(t, err)
		assert.Equal(t, 4, size)
	}
}
