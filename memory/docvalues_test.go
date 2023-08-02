package memory

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewNumericDocValues(t *testing.T) {
	numDocValues := newNumericDocValues(10)
	_, err := numDocValues.NextDoc()
	assert.Nil(t, err)
	id := numDocValues.DocID()
	assert.Equal(t, math.MaxInt32, id)

	advance, err := numDocValues.Advance(1)
	assert.Nil(t, err)
	assert.Equal(t, 1, advance)

	value, err := numDocValues.LongValue()
	assert.Nil(t, err)
	assert.Equal(t, int64(10), value)
}

func TestNewSortedDocValues(t *testing.T) {
	content := []byte("xxxxxxx")

	docValues := newSortedDocValues(content)
	advance, err := docValues.Advance(1)
	assert.Nil(t, err)
	assert.Equal(t, 1, advance)
	assert.Equal(t, 1, docValues.DocID())

	docId, err := docValues.NextDoc()
	assert.Nil(t, err)
	assert.Equal(t, 2, docId)
	assert.Equal(t, 2, docValues.DocID())

	docId, err = docValues.NextDoc()
	assert.Nil(t, err)
	assert.Equal(t, 3, docId)
	assert.Equal(t, 3, docValues.DocID())

	docId, err = docValues.NextDoc()
	assert.Nil(t, err)
	assert.Equal(t, 4, docId)
	assert.Equal(t, 4, docValues.DocID())

	slowAdvance, err := docValues.SlowAdvance(10)
	assert.Nil(t, err)
	assert.Equal(t, 10, slowAdvance)

	cost := docValues.Cost()
	assert.Equal(t, int64(1), cost)

	ordValue, err := docValues.OrdValue()
	assert.Nil(t, err)
	assert.Equal(t, 0, ordValue)

	ord1, err := docValues.LookupOrd(0)
	assert.Nil(t, err)
	assert.Equal(t, content, ord1)

	ord2, err := docValues.LookupOrd(1)
	assert.Nil(t, err)
	assert.Equal(t, content, ord2)

	valueCount := docValues.GetValueCount()
	assert.Equal(t, 1, valueCount)

	advanceExact, err := docValues.AdvanceExact(1)
	assert.Nil(t, err)
	assert.Equal(t, true, advanceExact)
}
