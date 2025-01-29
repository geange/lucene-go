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

// QueryType
// Used by RangeFieldQuery to check how each internal or leaf node relates to the query.
type QueryType interface {
	Compare(ctx context.Context, queryPackedValue, minPackedValue, maxPackedValue []byte, numDims, bytesPerDim, dim int) (types.Relation, error)
	Matches(ctx context.Context, queryPackedValue, packedValue []byte, numDims, bytesPerDim, dim int) bool
}

func matches(ctx context.Context, queryType QueryType, queryPackedValue, packedValue []byte, numDims, bytesPerDim int) bool {
	for dim := 0; dim < numDims; dim++ {
		if queryType.Matches(ctx, queryPackedValue, packedValue, numDims, bytesPerDim, dim) == false {
			return false
		}
	}
	return true
}

var _ QueryType = &INTERSECTS_QueryType{}

type INTERSECTS_QueryType struct {
}

func (*INTERSECTS_QueryType) Compare(ctx context.Context, queryPackedValue, minPackedValue, maxPackedValue []byte, numDims, bytesPerDim, dim int) (types.Relation, error) {
	//TODO implement me
	panic("implement me")
}

func (*INTERSECTS_QueryType) Matches(ctx context.Context, queryPackedValue, packedValue []byte, numDims, bytesPerDim, dim int) bool {
	//TODO implement me
	panic("implement me")
}
