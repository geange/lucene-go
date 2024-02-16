package builder

import (
	"bytes"

	"github.com/geange/lucene-go/core/search"
)

type Binary struct{}

// NewExactQuery
// Create a query for matching an exact binary value.
// This is for simple one-dimension points, for multidimensional points use
// NewRangeQuery(String, []byte, []byte) instead.
// field: field name. must not be null.
// value: binary value
func (b *Binary) NewExactQuery(field string, value []byte) (search.Query, error) {
	return b.NewRangeQuery(field, value, value)
}

func (b *Binary) NewRangeQuery(field string, lower, upper []byte) (search.Query, error) {
	return b.NewRangeQueryNDim(field, [][]byte{lower}, [][]byte{upper})
}

func (b *Binary) NewRangeQueryNDim(field string, lower, upper [][]byte) (search.Query, error) {
	packLower := bytes.Join(lower, []byte{})
	packUpper := bytes.Join(upper, []byte{})
	return search.NewPointRangeQuery(field, packLower, packUpper, len(lower))
}
