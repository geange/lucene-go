package attribute

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultAttributeFactory_CreateAttributeInstance(t *testing.T) {
	classes := []string{
		ClassBytesTerm,
		ClassCharTerm,
		ClassOffset,
		ClassPositionIncrement,
		ClassPayload,
		ClassPositionLength,
		ClassTermFrequency,
		ClassTermToBytesRef,
	}

	for _, class := range classes {
		_, err := DEFAULT_ATTRIBUTE_FACTORY.CreateAttributeInstance(class)
		assert.Nil(t, err)
	}

	_, err := DEFAULT_ATTRIBUTE_FACTORY.CreateAttributeInstance("")
	assert.NotNil(t, err)
}
