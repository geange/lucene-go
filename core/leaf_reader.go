package core

import "github.com/geange/lucene-go/core/util"

type LeafReader interface {
	IndexReader

	// GetFieldInfos Get the FieldInfos describing all fields in this reader. Note: Implementations should cache the FieldInfos instance returned by this method such that subsequent calls to this method return the same instance.
	GetFieldInfos() *FieldInfos

	// GetLiveDocs Returns the Bits representing live (not deleted) docs. A set bit indicates the doc ID has
	// not been deleted. If this method returns null it means there are no deleted documents
	// (all documents are live). The returned instance has been safely published for use by multiple threads
	// without additional synchronization.
	GetLiveDocs() util.Bits

	// GetPointValues Returns the PointValues used for numeric or spatial searches for the given field, or null
	// if there are no point fields.
	GetPointValues(field string) (PointValues, error)

	// CheckIntegrity Checks consistency of this reader.
	// Note that this may be costly in terms of I/O, e.g. may involve computing a checksum value against large data files.
	CheckIntegrity() error

	// GetMetaData Return metadata about this leaf.
	GetMetaData() *LeafMetaData
}
