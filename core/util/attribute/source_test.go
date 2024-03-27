package attribute

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAttributeSource(t *testing.T) {
	source := NewSource()

	assert.ElementsMatch(t, []string{ClassBytesTerm, ClassTermToBytesRef}, source.BytesTerm().Interfaces())
	assert.ElementsMatch(t, []string{ClassCharTerm, ClassTermToBytesRef}, source.CharTerm().Interfaces())
	assert.ElementsMatch(t, []string{ClassPayload}, source.Payload().Interfaces())

	err := source.BytesTerm().SetBytes([]byte("hello"))
	assert.Nil(t, err)
	assert.Equal(t, []byte("hello"), source.BytesTerm().GetBytes())
	assert.Equal(t, []byte("hello"), source.Term2Bytes().GetBytes())
	err = source.BytesTerm().Reset()
	assert.Nil(t, err)
	assert.Equal(t, []byte{}, source.BytesTerm().GetBytes())

	err = source.CharTerm().AppendString("hello")
	assert.Nil(t, err)
	assert.Equal(t, "hello", source.CharTerm().GetString())

	err = source.CharTerm().AppendString(" word")
	assert.Nil(t, err)
	assert.Equal(t, "hello word", source.CharTerm().GetString())

	err = source.CharTerm().AppendRune('!')
	assert.Nil(t, err)
	assert.Equal(t, "hello word!", source.CharTerm().GetString())

	err = source.CharTerm().Reset()
	assert.Nil(t, err)
	assert.Equal(t, "", source.CharTerm().GetString())

	err = source.Payload().SetPayload([]byte{1, 2, 3, 4})
	assert.Nil(t, err)
	assert.Equal(t, []byte{1, 2, 3, 4}, source.Payload().GetPayload())

	source.Type().SetType("x")
	assert.Equal(t, "x", source.Type().Type())

	err = source.Offset().SetOffset(1, 9)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, source.Offset().StartOffset())
	assert.EqualValues(t, 9, source.Offset().EndOffset())

	err = source.Offset().SetOffset(100, 9)
	assert.NotNil(t, err)

	err = source.PositionIncrement().SetPositionIncrement(98)
	assert.Nil(t, err)
	assert.EqualValues(t, 98, source.PositionIncrement().GetPositionIncrement())

	err = source.PositionLength().SetPositionLength(192)
	assert.Nil(t, err)
	assert.EqualValues(t, 192, source.PositionLength().GetPositionLength())

	err = source.TermFrequency().SetTermFrequency(25)
	assert.Nil(t, err)
	assert.EqualValues(t, 25, source.TermFrequency().GetTermFrequency())

	err = source.Reset()
	assert.Nil(t, err)
	assert.EqualValues(t, 1, source.TermFrequency().GetTermFrequency())
}
