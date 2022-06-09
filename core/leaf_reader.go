package core

type LeafReader interface {
	IndexReader

	// Returns the PointValues used for numeric or spatial searches for the given field, or null if there are no point fields.
	getPointValues(field string)

	// CheckIntegrity Checks consistency of this reader.
	// Note that this may be costly in terms of I/O, e.g. may involve computing a checksum value against large data files.
	CheckIntegrity() error

	// GetMetaData Return metadata about this leaf.
	GetMetaData() *LeafMetaData
}
