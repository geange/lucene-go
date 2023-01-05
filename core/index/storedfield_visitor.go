package index

import "github.com/geange/lucene-go/core/types"

// StoredFieldVisitor Expert: provides a low-level means of accessing the stored field values in an index.
// See IndexReader.document(int, StoredFieldVisitor).
//
// NOTE: a StoredFieldVisitor implementation should not try to load or visit other stored documents in
// the same reader because the implementation of stored fields for most codecs is not reentrant and you
// will see strange exceptions as a result.
//
// See DocumentStoredFieldVisitor, which is a StoredFieldVisitor that builds the Document containing
// all stored fields. This is used by IndexReader.document(int).
// lucene.experimental
type StoredFieldVisitor interface {

	// BinaryField Process a binary field.
	// @param value newly allocated byte array with the binary contents.
	BinaryField(fieldInfo *types.FieldInfo, value []byte) error

	// Process a string field; the provided byte[] value is a UTF-8 encoded string value.
	StringField(fieldInfo *types.FieldInfo, value []byte) error

	// Process a int numeric field.
	Int32Field(fieldInfo *types.FieldInfo, value int32) error

	// Process a long numeric field.
	Int64Field(fieldInfo *types.FieldInfo, value int64) error

	// Process a float numeric field.
	Float32Field(fieldInfo *types.FieldInfo, value float32) error

	// Process a double numeric field.
	Float64Field(fieldInfo *types.FieldInfo, value float64) error

	// NeedsField Hook before processing a field. Before a field is processed, this method is invoked
	//so that subclasses can return a StoredFieldVisitor.Status representing
	// whether they need that particular field or not, or to stop processing entirely.
	NeedsField(fieldInfo *types.FieldInfo) (STORED_FIELD_VISITOR_STATUS, error)
}

type STORED_FIELD_VISITOR_STATUS int

const (
	STORED_FIELD_VISITOR_YES  = STORED_FIELD_VISITOR_STATUS(iota) // YES: the field should be visited.
	STORED_FIELD_VISITOR_NO                                       // NO: don't visit this field, but continue processing fields for this document.
	STORED_FIELD_VISITOR_STOP                                     // STOP: don't visit this field and stop processing any other fields for this document.
)
