package document

import (
	"bytes"
	"iter"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBinaryPoint(t *testing.T) {
	doc := NewDocument()

	type KV struct {
		Name  string
		Value [][]byte
	}

	kvs := []KV{
		{"f1", [][]byte{[]byte("v1"), []byte("v2")}},
		{"f2", [][]byte{[]byte("v1"), []byte("v2")}},
	}

	next, stop := iter.Pull(slices.Values(kvs))
	defer stop()

	for _, kv := range kvs {
		field, err := NewBinaryPoint(kv.Name, kv.Value...)
		assert.Nil(t, err)
		doc.Add(field)
	}

	for field := range doc.GetFields() {
		kv, ok := next()
		assert.True(t, ok)

		value, ok := field.Get().([]byte)
		assert.True(t, ok)

		assert.EqualValues(t, bytes.Join(kv.Value, []byte{}), value)
	}
}

func TestNewBinaryDocValuesField(t *testing.T) {
	doc := NewDocument()

	type KV struct {
		Name  string
		Value string
	}

	kvs := []KV{
		{"f1", "v1"},
		{"f1", "v2"},
		{"f2", "v2"},
	}

	next, stop := iter.Pull(slices.Values(kvs))
	defer stop()

	for _, kv := range kvs {
		field := NewBinaryDocValuesField(kv.Name, []byte(kv.Value))
		doc.Add(field)
	}

	for field := range doc.GetFields() {
		kv, ok := next()
		assert.True(t, ok)

		value, ok := field.Get().([]byte)
		assert.True(t, ok)

		assert.Equal(t, kv.Value, string(value))
	}
}

func TestNewBinaryRangeDocValuesField(t *testing.T) {
	doc := NewDocument()

	type KV struct {
		Name  string
		Value [][]byte
	}

	kvs := []KV{
		{"f1", [][]byte{[]byte("v1")}},
		{"f2", [][]byte{[]byte("v1")}},
	}

	next, stop := iter.Pull(slices.Values(kvs))
	defer stop()

	for _, kv := range kvs {
		field, err := NewBinaryRangeDocValuesField(kv.Name, kv.Value)
		assert.Nil(t, err)
		doc.Add(field)
	}

	for field := range doc.GetFields() {
		kv, ok := next()
		assert.True(t, ok)

		value, ok := field.Get().([]byte)
		assert.True(t, ok)

		assert.Equal(t, bytes.Join(kv.Value, []byte{}), value)
	}
}
