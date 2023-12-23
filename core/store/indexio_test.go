package store

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndexInputBase(t *testing.T) {

}

func TestIndexOutputBase(t *testing.T) {
	outputWrap := NewBaseIndexOutput("x", new(bytes.Buffer))
	assert.Equal(t, "x", outputWrap.GetName())
}
