package query

import (
	"context"

	"github.com/geange/lucene-go/core/types"
)

// RangeFieldQuery
// Query class for searching RangeField types by a defined PointValues. Relation.
type RangeFieldQuery struct {
	field       string
	queryType   QueryType
	numDims     int
	ranges      []byte
	bytesPerDim int
}

type QueryType interface {
	Compare(ctx context.Context, queryPackedValue, minPackedValue, maxPackedValue []byte, numDims, bytesPerDim, dim int) (types.Relation, error)
	Matches(ctx context.Context, queryPackedValue, packedValue []byte, numDims, bytesPerDim, dim int) bool
}

/*
 boolean matches(byte[] queryPackedValue, byte[] packedValue, int numDims, int bytesPerDim) {
      for (int dim = 0; dim < numDims; ++dim) {
        if (matches(queryPackedValue, packedValue, numDims, bytesPerDim, dim) == false) {
          return false;
        }
      }
      return true;
    }
*/

func matches(ctx context.Context, queryType QueryType, queryPackedValue, packedValue []byte, numDims, bytesPerDim int) bool {
	for dim := 0; dim < numDims; dim++ {
		if queryType.Matches(ctx, queryPackedValue, packedValue, numDims, bytesPerDim, dim) == false {
			return false
		}
	}
	return true
}
