package document

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/geange/gods-generic/utils"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/search"
)

type BinaryPoint struct {
}

// NewExactQuery
// Create a query for matching an exact binary value.
// This is for simple one-dimension points, for multidimensional points use newRangeQuery(String, byte[][], byte[][]) instead.
// field: field name. must not be null.
// value: binary value
// Returns: a query matching documents with this exact value
// Throws: IllegalArgumentException – if field is null or value is null
func (b *BinaryPoint) NewExactQuery(field string, value []byte) (search.Query, error) {
	return b.NewRangeQuery(field, value, value)
}

func (b *BinaryPoint) NewRangeQuery(field string, lowerValue, upperValue []byte) (search.Query, error) {
	return b.NewRangeQueryNDim(field, [][]byte{lowerValue}, [][]byte{upperValue})
}

func (b *BinaryPoint) NewRangeQueryNDim(field string, lowerValue, upperValue [][]byte) (search.Query, error) {
	lowerPoint, err := document.BinaryPointPack(lowerValue...)
	if err != nil {
		return nil, err
	}
	upperPoint, err := document.BinaryPointPack(upperValue...)
	if err != nil {
		return nil, err
	}

	dim := len(lowerValue)

	fn := func(dimension int, value []byte) string {
		sb := new(bytes.Buffer)
		sb.WriteString("binary(")
		for i, v := range value {
			if i > 0 {
				sb.WriteString(" ")
			}
			sb.WriteString(fmt.Sprintf("%x", v))
		}
		sb.WriteString(")")
		return sb.String()
	}
	return search.NewPointRangeQuery(field, lowerPoint, upperPoint, dim, fn)
}

func (b *BinaryPoint) NewSetQuery(field string, values ...[]byte) (search.Query, error) {
	// Make sure all byte[] have the same length
	bytesPerDim := -1

	for _, value := range values {
		if bytesPerDim == -1 {
			bytesPerDim = len(value)
		} else if len(value) != bytesPerDim {
			return nil, errors.New("all byte[] must be the same length")
		}

	}

	if bytesPerDim == -1 {
		// There are no points, and we cannot guess the bytesPerDim here, so we return an equivalent query:
		return search.NewMatchNoDocsQuery(), nil
	}

	// Don't unexpectedly change the user's incoming values array:
	sortedValues := make([][]byte, 0, len(values))
	for _, value := range values {
		newArray := make([]byte, len(value))
		copy(newArray, value)
		sortedValues = append(sortedValues, newArray)
	}

	utils.SortGeneric(sortedValues, bytes.Compare)

	return search.NewPointInSetQuery(field, 1, bytesPerDim, sortedValues)
}
