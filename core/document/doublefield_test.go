package document

import (
	"iter"
	"math"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFloat64Point(t *testing.T) {
	doc := NewDocument()

	type KV struct {
		Name  string
		Value []float64
	}

	kvs := []KV{
		{"f1", []float64{1.1, 1.2, 1.4}},
		{"f2", []float64{2.1, 2, 2.5}},
		{"f3", []float64{-1, 2.1, 2.5}},
	}

	next, stop := iter.Pull(slices.Values(kvs))
	defer stop()

	for _, kv := range kvs {
		field, err := NewFloat64Point(kv.Name, kv.Value...)
		assert.Nil(t, err)
		doc.Add(field)
	}

	for field := range doc.GetFields() {
		kv, ok := next()
		assert.True(t, ok)

		points := field.(*Float64Point).Points()

		assert.Equal(t, len(kv.Value), len(points))

		for i, num := range kv.Value {
			assert.Less(t, math.Abs(num-points[i]), 0.000001)
		}
	}
}
