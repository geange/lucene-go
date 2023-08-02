package store

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRWByte(t *testing.T) {
	source := NewBuffer()
	reader := NewBaseDataInput(source)
	defer reader.Close()
	writer := NewBaseDataOutput(source)

	err := writer.WriteByte('a')
	assert.Nil(t, err)
	err = writer.WriteByte('b')
	assert.Nil(t, err)
	err = writer.WriteByte('c')
	assert.Nil(t, err)

	c1, err := reader.ReadByte()
	assert.Nil(t, err)
	assert.Equal(t, byte('a'), c1)

	c2, err := reader.ReadByte()
	assert.Nil(t, err)
	assert.Equal(t, byte('b'), c2)

	c3, err := reader.ReadByte()
	assert.Nil(t, err)
	assert.Equal(t, byte('c'), c3)
}

func TestRWUint16(t *testing.T) {
	source := NewBuffer()
	reader := NewBaseDataInput(source)
	writer := NewBaseDataOutput(source)

	err := writer.WriteUint16(context.Background(), 1024)
	assert.Nil(t, err)
	num, err := reader.ReadUint16(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, uint16(1024), num)
}

func TestRWUint32(t *testing.T) {
	source := NewBuffer()
	reader := NewBaseDataInput(source)
	writer := NewBaseDataOutput(source)

	err := writer.WriteUint32(context.Background(), 1024)
	assert.Nil(t, err)
	num, err := reader.ReadUint32(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, uint32(1024), num)
}

func TestRWUint64(t *testing.T) {
	source := NewBuffer()
	reader := NewBaseDataInput(source)
	writer := NewBaseDataOutput(source)

	err := writer.WriteUint64(context.Background(), 1024)
	assert.Nil(t, err)
	num, err := reader.ReadUint64(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, uint64(1024), num)
}

func TestRWUvarint(t *testing.T) {
	source := NewBuffer()
	reader := NewBaseDataInput(source)
	writer := NewBaseDataOutput(source)

	err := writer.WriteUvarint(context.Background(), 1024)
	assert.Nil(t, err)
	num, err := reader.ReadUvarint(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, uint64(1024), num)
}

func TestRWString(t *testing.T) {
	source := NewBuffer()
	reader := NewBaseDataInput(source)
	writer := NewBaseDataOutput(source)

	err := writer.WriteString(context.Background(), "1024")
	assert.Nil(t, err)
	num, err := reader.ReadString(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, "1024", num)
}

func TestRWMapOfStrings(t *testing.T) {
	source := NewBuffer()
	reader := NewBaseDataInput(source)
	writer := NewBaseDataOutput(source)

	kvs := map[string]string{
		"1": "1024",
		"2": "1025",
		"3": "1026",
	}

	err := writer.WriteMapOfStrings(context.Background(), kvs)
	assert.Nil(t, err)
	num, err := reader.ReadMapOfStrings(context.Background())
	assert.Nil(t, err)
	assert.EqualValues(t, kvs, num)
}

func TestRWSetOfStrings(t *testing.T) {
	source := NewBuffer()
	reader := NewBaseDataInput(source)
	writer := NewBaseDataOutput(source)

	kvs := map[string]struct{}{
		"1": {},
		"2": {},
		"3": {},
	}

	err := writer.WriteSetOfStrings(context.Background(), kvs)
	assert.Nil(t, err)
	num, err := reader.ReadSetOfStrings(context.Background())
	assert.Nil(t, err)
	assert.EqualValues(t, kvs, num)
}

func TestReader_SkipBytes(t *testing.T) {
	source := NewBuffer()
	source.Write(make([]byte, 104))

	reader := NewBaseDataInput(source)

	err := reader.SkipBytes(context.Background(), 100)
	assert.Nil(t, err)

	assert.EqualValues(t, 4, source.Len())
}

func TestWriter_CopyBytes(t *testing.T) {
	src := new(bytes.Buffer)
	src.Write(make([]byte, 100))
	src.Write([]byte{1, 2, 3, 4})
	input := NewBufferDataInput(src)

	dest := new(bytes.Buffer)
	writer := NewBaseDataOutput(dest)

	err := writer.CopyBytes(context.Background(), input, 104)
	assert.Nil(t, err)

	assert.EqualValues(t, 104, dest.Len())
	assert.EqualValues(t, []byte{1, 2, 3, 4}, dest.Bytes()[100:])
}

func TestReader_Clone(t *testing.T) {
	//io.ReaderAt()
	//
	//source := new(bytes.Buffer)
	//source.Write(make([]byte, 100))
	//source.WriteByte(99)
	//reader := NewReader(source)
	//
	//dest := new(bytes.Buffer)
	//
	//reader.Clone(dest)

}

func TestBuffer_Clone(t *testing.T) {
	source := NewBuffer()
	reader := NewBaseDataInput(source)
	defer reader.Close()
	writer := NewBaseDataOutput(source)

	err := writer.WriteByte('a')
	assert.Nil(t, err)
	err = writer.WriteByte('b')
	assert.Nil(t, err)
	err = writer.WriteByte('c')
	assert.Nil(t, err)

	c1, err := reader.ReadByte()
	assert.Nil(t, err)
	assert.Equal(t, byte('a'), c1)

	readerClone := NewBaseDataInput(source.Clone())
	defer readerClone.Close()

	c2, err := reader.ReadByte()
	assert.Nil(t, err)
	assert.Equal(t, byte('b'), c2)

	{
		c2, err := readerClone.ReadByte()
		assert.Nil(t, err)
		assert.Equal(t, byte('b'), c2)
	}

	c3, err := reader.ReadByte()
	assert.Nil(t, err)
	assert.Equal(t, byte('c'), c3)

	{
		c3, err := readerClone.ReadByte()
		assert.Nil(t, err)
		assert.Equal(t, byte('c'), c3)
	}
}
