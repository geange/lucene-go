package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSliceByteStartArray(t *testing.T) {
	startArray := newSliceByteStartArray(10)
	startArray.Init()

	assert.Equal(t, 10, len(startArray.bytesStart))
	assert.Equal(t, 12, len(startArray.start))
	assert.Equal(t, 12, len(startArray.end))
	assert.Equal(t, 12, len(startArray.freq))

	startArray.Grow()
	assert.Equal(t, 11, len(startArray.bytesStart))
	assert.Equal(t, 12, len(startArray.start))
	assert.Equal(t, 12, len(startArray.end))
	assert.Equal(t, 12, len(startArray.freq))

	startArray.Grow()
	assert.Equal(t, 12, len(startArray.bytesStart))
	assert.Equal(t, 12, len(startArray.start))
	assert.Equal(t, 12, len(startArray.end))
	assert.Equal(t, 12, len(startArray.freq))

	startArray.Grow()
	assert.Equal(t, 13, len(startArray.bytesStart))
	assert.Equal(t, 13, len(startArray.start))
	assert.Equal(t, 13, len(startArray.end))
	assert.Equal(t, 13, len(startArray.freq))

	startArray.Clear()
	assert.Equal(t, 0, len(startArray.bytesStart))
	assert.Equal(t, 0, len(startArray.start))
	assert.Equal(t, 0, len(startArray.end))
	assert.Equal(t, 0, len(startArray.freq))
}
