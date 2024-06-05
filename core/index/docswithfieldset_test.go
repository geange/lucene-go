package index

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDocsWithFieldSet(t *testing.T) {
	fieldSet := NewDocsWithFieldSet()

	for i := 0; i < 100; i++ {
		err := fieldSet.Add(i)
		assert.Nil(t, err)
	}

	iterator, err := fieldSet.Iterator()
	assert.Nil(t, err)

	for i := 0; i < 100; i++ {
		docId, err := iterator.NextDoc()
		assert.Nil(t, err)
		assert.EqualValues(t, i, docId)

		curDocID := iterator.DocID()
		assert.EqualValues(t, i, curDocID)
	}
}
