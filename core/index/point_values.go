package index

import (
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
	Intersect(visitor IntersectVisitor) error

	// EstimatePointCount Estimate the number of points that would be visited
	// by intersect with the given PointValues.BytesVisitor.
	// This should run many times faster than intersect(PointValues.BytesVisitor).
	EstimatePointCount(visitor IntersectVisitor) int64

	// EstimateDocCount
	// Estimate the number of documents that would be matched by intersect with the given
	// PointValues.IntersectVisitor. This should run many times faster than
	// intersect(PointValues.IntersectVisitor).
	// See Also: DocIdSetIterator.cost
	EstimateDocCount(visitor IntersectVisitor) int64

	// GetMinPackedValue Returns minimum value for each dimension, packed, or null if size is 0
	GetMinPackedValue() ([]byte, error)

	// GetMaxPackedValue Returns maximum value for each dimension, packed, or null if size is 0
	GetMaxPackedValue() ([]byte, error)

	// GetNumDimensions Returns how many dimensions are represented in the values
	GetNumDimensions() (int, error)

	// GetNumIndexDimensions Returns how many dimensions are used for the index
	GetNumIndexDimensions() (int, error)

	// GetBytesPerDimension Returns the number of bytes per dimension
	GetBytesPerDimension() (int, error)

	// Size Returns the total number of indexed points across all documents.
	Size() int64

	// GetDocCount Returns the total number of documents that have indexed at least one point.
	GetDocCount() int
}

type EstimateDocCountSPI interface {
	EstimatePointCount(visitor IntersectVisitor) int64
	Size() int64
	GetDocCount() int
}

func EstimateDocCount(spi EstimateDocCountSPI, visitor IntersectVisitor) int64 {

	estimatedPointCount := spi.EstimatePointCount(visitor)
	docCount := spi.GetDocCount()
	size := spi.Size()
	if estimatedPointCount >= size {
		// math all docs
		return int64(docCount)
	} else if size == int64(docCount) || estimatedPointCount == 0 {
		// if the point count estimate is 0 or we have only single values
		// return this estimate
		return estimatedPointCount
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
			return 1
		}
		return int64(docEstimate)
	}
}

type IntersectVisitor interface {
	Visit(docID int) error
	VisitLeaf(docID int, packedValue []byte) error
	VisitIterator(iterator DocValuesIterator, packedValue []byte) error
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
	VisitLeafFn func(docID int, packedValue []byte) error

	// CompareFn Called for non-leaf cells to test how the cell relates to the query,
	// to determine how to further recurse down the tree.
	CompareFn func(minPackedValue, maxPackedValue []byte) Relation

	// GrowFn Notifies the caller that this many documents are about to be visited
	GrowFn func(count int)
}

func (r *BytesVisitor) Visit(docID int) error {
	return r.VisitFn(docID)
}

func (r *BytesVisitor) VisitLeaf(docID int, packedValue []byte) error {
	return r.VisitLeafFn(docID, packedValue)
}

func (r *BytesVisitor) Compare(minPackedValue, maxPackedValue []byte) Relation {
	return r.CompareFn(minPackedValue, maxPackedValue)
}

func (r *BytesVisitor) Grow(count int) {
	r.GrowFn(count)
}

func (r *BytesVisitor) VisitIterator(iterator DocValuesIterator, packedValue []byte) error {
	for {
		docID, err := iterator.NextDoc()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		if err := r.VisitLeafFn(docID, packedValue); err != nil {
			return err
		}
	}
}
