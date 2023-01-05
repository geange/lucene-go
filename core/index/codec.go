package index

// Codec Encodes/decodes an inverted index segment.
// Note, when extending this class, the name (getName) is written into the index. In order for the segment to be read, the name must resolve to your implementation via forName(String). This method uses Java's Service Provider Interface (SPI) to resolve codec names.
// If you implement your own codec, make sure that it has a no-arg constructor so SPI can load it.
// See Also: ServiceLoader
type Codec interface {

	// PostingsFormat Encodes/decodes postings
	PostingsFormat() PostingsFormat

	// DocValuesFormat Encodes/decodes docvalues
	DocValuesFormat() DocValuesFormat

	// StoredFieldsFormat Encodes/decodes stored fields
	StoredFieldsFormat() StoredFieldsFormat

	// TermVectorsFormat Encodes/decodes term vectors
	TermVectorsFormat() TermVectorsFormat

	// FieldInfosFormat Encodes/decodes field infos file
	FieldInfosFormat() FieldInfosFormat

	// SegmentInfoFormat Encodes/decodes segment info file
	SegmentInfoFormat() SegmentInfoFormat

	// NormsFormat Encodes/decodes document normalization values
	NormsFormat() NormsFormat

	// LiveDocsFormat Encodes/decodes live docs
	LiveDocsFormat() LiveDocsFormat

	// CompoundFormat Encodes/decodes compound files
	CompoundFormat() CompoundFormat

	// PointsFormat Encodes/decodes points index
	PointsFormat() PointsFormat
}

type NamedSPI interface {
	GetName() string
}
