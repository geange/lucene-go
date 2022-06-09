package core

// PointValues Access to indexed numeric values.
// Points represent numeric values and are indexed differently than ordinary text. Instead of an inverted index,
// points are indexed with datastructures such as KD-trees . These structures are optimized for operations such
// as range, distance, nearest-neighbor, and point-in-polygon queries.
// Basic Point Types
// Basic point types in Java and Lucene
// Java type 	Lucene class
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
//     // add year 1970 to document
//     document.add(new IntPoint("year", 1970));
//     // index document
//     writer.addDocument(document);
//     ...
//     // issue range query of 1960-1980
//     Query query = IntPoint.newRangeQuery("year", 1960, 1980);
//     TopDocs docs = searcher.search(query, ...);
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
type PointValues struct {
}
