package document

import (
	"github.com/geange/lucene-go/core/analysis"
)

// IndexableField
// Represents a single field for indexing. IndexWriter consumes
// []IndexableField as a document.
// IndexableField 代表一个可以被索引的field，每一个Document都是由多个IndexAbleField组成
type IndexableField interface {
	Name() string

	FieldType() IndexableFieldType

	// TokenStream
	// Creates the TokenStream used for indexing this field. If appropriate, implementations should
	// use the given Analyzer to create the TokenStreams.
	// analyzer: Analyzer that should be used to create the TokenStreams from
	// reuse: TokenStream for a previous instance of this field name. This allows custom field types
	//	(like StringField and NumericField) that do not use the analyzer to still have good performance.
	//	Note: the passed-in types may be inappropriate, for example if you mix up different types of Fields
	//	for the same field name. So it's the responsibility of the implementation to check.
	// TokenStream value for indexing the document. Should always return a non-null value if the field is to be indexed
	TokenStream(analyzer analysis.Analyzer, reuse analysis.TokenStream) (analysis.TokenStream, error)

	Get() any

	Set(v any) error

	Number() (any, bool)
}

// IndexableFieldType Describes the properties of a field.
type IndexableFieldType interface {
	// Stored
	// True if the field's value should be stored
	Stored() bool

	// Tokenized
	// True if this field's value should be analyzed by the Analyzer.
	// This has no effect if indexOptions() returns IndexOptions.NONE.
	Tokenized() bool

	// StoreTermVectors
	// True if this field's indexed form should be also stored into term vectors.
	// This builds a miniature inverted-index for this field which can be accessed in a document-oriented
	// way from IndexReader.getTermVector(int, String).
	// This option is illegal if indexOptions() returns IndexOptions.NONE.
	StoreTermVectors() bool

	// StoreTermVectorOffsets
	// True if this field's token character offsets should also be stored into term vectors.
	// This option is illegal if term vectors are not enabled for the field (storeTermVectors() is false)
	StoreTermVectorOffsets() bool

	// StoreTermVectorPositions
	// True if this field's token positions should also be stored into the term vectors.
	// This option is illegal if term vectors are not enabled for the field (storeTermVectors() is false).
	StoreTermVectorPositions() bool

	// StoreTermVectorPayloads
	// True if this field's token payloads should also be stored into the term vectors.
	// This option is illegal if term vector positions are not enabled for the field (storeTermVectors() is false).
	StoreTermVectorPayloads() bool

	// OmitNorms
	// True if normalization values should be omitted for the field.
	// This saves memory, but at the expense of scoring quality (length normalization will be disabled),
	// and if you omit norms, you cannot use index-time boosts.
	OmitNorms() bool

	// IndexOptions
	// describing what should be recorded into the inverted index
	IndexOptions() IndexOptions

	// DocValuesType
	// DocValues DocValuesType: how the field's value will be indexed into docValues.
	DocValuesType() DocValuesType

	// PointDimensionCount
	// If this is positive (representing the number of point dimensions),
	// the field is indexed as a point.
	PointDimensionCount() int

	// PointIndexDimensionCount
	// The number of dimensions used for the index key
	PointIndexDimensionCount() int

	// PointNumBytes
	// The number of bytes in each dimension's values.
	PointNumBytes() int

	// GetAttributes
	// Attributes for the field types. Attributes are not thread-safe, user must not add
	// attributes while other threads are indexing documents with this field types.
	GetAttributes() map[string]string
}
