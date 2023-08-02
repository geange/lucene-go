package attribute

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBytesAttr_CopyTo(t *testing.T) {
	attr := newBytesAttr("x")
	target := newBytesAttr("j", "l")
	err := attr.CopyTo(target)
	assert.Nil(t, err)
	assert.EqualValues(t, attr, target)
}

func TestBytesAttrClone(t *testing.T) {
	attr := newBytesAttr("x")
	newAttr := attr.Clone()
	assert.EqualValues(t, attr, newAttr)
}
