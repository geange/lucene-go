package types

// IndexableFieldType Describes the properties of a field.
// 描述一个field的属性
type IndexableFieldType interface {
	// Stored True if the field's value should be stored
	Stored() bool

	// Tokenized True if this field's value should be analyzed by the Analyzer.
	//This has no effect if indexOptions() returns IndexOptions.NONE.
	Tokenized() bool

	// StoreTermVectors True if this field's indexed form should be also stored into term vectors.
	// This builds a miniature inverted-index for this field which can be accessed in a document-oriented
	// way from IndexReader.getTermVector(int, String).
	// This option is illegal if indexOptions() returns IndexOptions.NONE.
	StoreTermVectors() bool

	// StoreTermVectorOffsets True if this field's token character offsets should also be stored into term vectors.
	// This option is illegal if term vectors are not enabled for the field (storeTermVectors() is false)
	StoreTermVectorOffsets() bool

	// StoreTermVectorPositions True if this field's token positions should also be stored into the term vectors.
	// This option is illegal if term vectors are not enabled for the field (storeTermVectors() is false).
	StoreTermVectorPositions() bool

	// StoreTermVectorPayloads True if this field's token payloads should also be stored into the term vectors.
	// This option is illegal if term vector positions are not enabled for the field (storeTermVectors() is false).
	StoreTermVectorPayloads() bool

	// OmitNorms True if normalization values should be omitted for the field.
	// This saves memory, but at the expense of scoring quality (length normalization will be disabled),
	// and if you omit norms, you cannot use index-time boosts.
	// 忽略标准化值，节省内存？？？
	OmitNorms() bool

	// IndexOptions describing what should be recorded into the inverted index
	IndexOptions() IndexOptions

	// DocValuesType DocValues DocValuesType: how the field's value will be indexed into docValues.
	DocValuesType() DocValuesType

	// PointDimensionCount If this is positive (representing the number of point dimensions),
	// the field is indexed as a point.
	PointDimensionCount() int

	// PointIndexDimensionCount The number of dimensions used for the index key
	PointIndexDimensionCount() int

	// PointNumBytes The number of bytes in each dimension's values.
	PointNumBytes() int

	// GetAttributes Attributes for the field types. Attributes are not thread-safe, user must not add
	// attributes while other threads are indexing documents with this field types.
	//Returns: Map
	GetAttributes() map[string]string
}
