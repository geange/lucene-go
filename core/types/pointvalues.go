package types

import (
	"bytes"
	"context"
	"errors"
	"io"
	"math"
)

type Relation int

const (
	CELL_INSIDE_QUERY  = Relation(iota) // Return this if the cell is fully contained by the query
	CELL_OUTSIDE_QUERY                  // Return this if the cell and query do not overlap
	CELL_CROSSES_QUERY                  // Return this if the cell partially overlaps the query
)

// PointValues Access to indexed numeric values.
// Points represent numeric values and are indexed differently than ordinary text. Instead of an inverted index,
// points are indexed with datastructures such as KD-trees . These structures are optimized for operations such
// as range, distance, nearest-neighbor, and point-in-polygon queries.
// Basic Point Types
// Basic point types in Java and Lucene
// Java types 	Lucene class
// int 			IntPoint
// long 		LongPoint
// float 		FloatPoint
// double 		DoublePoint
// byte[] 		BinaryPoint
// BigInteger 	BigIntegerPoint*
// InetAddress 	InetAddressPoint*
// * in the lucene-sandbox jar
// Basic Lucene point types behave like their java peers: for example IntPoint represents a signed 32-bit Integer,
// supporting values ranging from Integer.MIN_VALUE to Integer.MAX_VALUE, ordered consistent with
// Integer.compareTo(Integer). In addition to indexing support, point classes also contain static methods
// (such as IntPoint.newRangeQuery(String, int, int)) for creating common queries. For example:
//
//	// add year 1970 to document
//	document.add(new IntPoint("year", 1970));
//	// index document
//	writer.addDocument(document);
//	...
//	// issue range query of 1960-1980
//	Query query = IntPoint.newRangeQuery("year", 1960, 1980);
//	TopDocs docs = searcher.search(query, ...);
//
// Geospatial Point Types
// Although basic point types such as DoublePoint support points in multi-dimensional space too, Lucene has
// specialized classes for location data. These classes are optimized for location data: they are more
// space-efficient and support special operations such as distance and polygon queries. There are currently
// two implementations:
// LatLonPoint: indexes (latitude,longitude) as (x,y) in two-dimensional space.
// Geo3DPoint* in lucene-spatial3d: indexes (latitude,longitude) as (x,y,z) in three-dimensional space.
// * does not support altitude, 3D here means "uses three dimensions under-the-hood"
// Advanced usage
// Custom structures can be created on top of single- or multi- dimensional basic types, on top of BinaryPoint
// for more flexibility, or via custom Field subclasses.
type PointValues interface {
	// Intersect Finds all documents and points matching the provided visitor.
	// This method does not enforce live documents,
	// so it's up to the caller to test whether each document is deleted, if necessary.
	Intersect(ctx context.Context, visitor IntersectVisitor) error

	// EstimatePointCount Estimate the number of points that would be visited
	// by intersect with the given PointValues.BytesVisitor.
	// This should run many times faster than intersect(PointValues.BytesVisitor).
	EstimatePointCount(ctx context.Context, visitor IntersectVisitor) (int, error)

	// EstimateDocCount
	// Estimate the number of documents that would be matched by intersect with the given
	// PointValues.IntersectVisitor. This should run many times faster than
	// intersect(PointValues.IntersectVisitor).
	// See Also: DocIdSetIterator.cost
	EstimateDocCount(ctx context.Context, visitor IntersectVisitor) (int, error)

	// GetMinPackedValue Returns minimum item for each dimension, packed, or null if size is 0
	GetMinPackedValue() ([]byte, error)

	// GetMaxPackedValue Returns maximum item for each dimension, packed, or null if size is 0
	GetMaxPackedValue() ([]byte, error)

	// GetNumDimensions Returns how many dimensions are represented in the values
	GetNumDimensions() (int, error)

	// GetNumIndexDimensions Returns how many dimensions are used for the index
	GetNumIndexDimensions() (int, error)

	// GetBytesPerDimension Returns the number of bytes per dimension
	GetBytesPerDimension() (int, error)

	// Size Returns the total number of indexed points across all documents.
	Size() int

	// GetDocCount Returns the total number of documents that have indexed at least one point.
	GetDocCount() int
}

type EstimateDocCountSPI interface {
	EstimatePointCount(ctx context.Context, visitor IntersectVisitor) (int, error)
	Size() int
	GetDocCount() int
}

func EstimateDocCount(ctx context.Context, spi EstimateDocCountSPI, visitor IntersectVisitor) (int, error) {

	estimatedPointCount, err := spi.EstimatePointCount(ctx, visitor)
	if err != nil {
		return 0, err
	}
	docCount := spi.GetDocCount()
	size := spi.Size()
	if estimatedPointCount >= size {
		// math all docs
		return docCount, nil
	} else if size == docCount || estimatedPointCount == 0 {
		// if the point count estimate is 0 or we have only single values
		// return this estimate
		return estimatedPointCount, nil
	} else {
		// in case of multi values estimate the number of docs using the solution provided in
		// https://math.stackexchange.com/questions/1175295/urn-problem-probability-of-drawing-balls-of-k-unique-colors
		// then approximate the solution for points per doc << size() which results in the expression
		// D * (1 - ((N - n) / N)^(N/D))
		// where D is the total number of docs, N the total number of points and n the estimated point count
		f64Size := float64(size)
		f64EstimatedPointCount := float64(estimatedPointCount)
		f64DocCount := float64(docCount)

		docEstimate := f64DocCount * (1 - math.Pow((f64Size-f64EstimatedPointCount)/f64Size, f64Size/f64DocCount))

		if docEstimate == 0 {
			return 1, nil
		}
		return int(docEstimate), nil
	}
}

type IntersectVisitor interface {
	Visit(ctx context.Context, docID int) error
	VisitLeaf(ctx context.Context, docID int, packedValue []byte) error
	Compare(minPackedValue, maxPackedValue []byte) Relation
	Grow(count int)
}

var _ IntersectVisitor = &BytesVisitor{}

type BytesVisitor struct {
	// VisitFn Called for all documents in a leaf cell that's fully contained by the query. The consumer
	// should blindly accept the docID.
	VisitFn func(docID int) error

	// VisitLeafFn Called for all documents in a leaf cell that crosses the query. The consumer should scrutinize the
	// packedValue to decide whether to accept it. In the 1D case, values are visited in increasing order,
	// and in the case of ties, in increasing docID order.
	VisitLeafFn func(ctx context.Context, docID int, packedValue []byte) error

	// CompareFn Called for non-leaf cells to test how the cell relates to the query,
	// to determine how to further recurse down the tree.
	CompareFn func(minPackedValue, maxPackedValue []byte) Relation

	// GrowFn Notifies the caller that this many documents are about to be visited
	GrowFn func(count int)
}

func (r *BytesVisitor) Visit(ctx context.Context, docID int) error {
	return r.VisitFn(docID)
}

func (r *BytesVisitor) VisitLeaf(ctx context.Context, docID int, packedValue []byte) error {
	return r.VisitLeafFn(ctx, docID, packedValue)
}

func (r *BytesVisitor) Compare(minPackedValue, maxPackedValue []byte) Relation {
	return r.CompareFn(minPackedValue, maxPackedValue)
}

func (r *BytesVisitor) Grow(count int) {
	r.GrowFn(count)
}

func Visit(ctx context.Context, visitor IntersectVisitor, iterator DocIdSetIterator, packedValue []byte) error {
	for {
		docID, err := iterator.NextDoc(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		if err := visitor.VisitLeaf(ctx, docID, packedValue); err != nil {
			return err
		}
	}
}

func (r *BytesVisitor) VisitIterator(ctx context.Context, iterator DocValuesIterator, packedValue []byte) error {
	for {
		docID, err := iterator.NextDoc(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		if err := r.VisitLeafFn(ctx, docID, packedValue); err != nil {
			return err
		}
	}
}

type MutablePointValues interface {
	PointValues

	// GetValue Set packedValue with a reference to the packed bytes of the i-th item.
	GetValue(i int, packedValue *bytes.Buffer)

	// GetByteAt Get the k-th byte of the i-th item.
	GetByteAt(i, k int) byte

	// GetDocID Return the doc ID of the i-th item.
	GetDocID(i int) int

	// Swap the i-th and j-th values.
	Swap(i, j int)

	// Save the i-th item into the j-th position in temporary storage.
	Save(i, j int)

	// Restore values between i-th and j-th(excluding) in temporary storage into original storage.
	Restore(i, j int)
}
