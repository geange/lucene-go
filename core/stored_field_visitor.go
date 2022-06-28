package core

// StoredFieldVisitor Expert: provides a low-level means of accessing the stored field values in an index.
// See IndexReader.document(int, StoredFieldVisitor).
// NOTE: a StoredFieldVisitor implementation should not try to load or visit other stored documents in the
// same reader because the implementation of stored fields for most codecs is not reentrant and you will see
// strange exceptions as a result.
// See DocumentStoredFieldVisitor, which is a StoredFieldVisitor that builds the Document containing all
// stored fields. This is used by IndexReader.document(int).
type StoredFieldVisitor interface {

	// BinaryField Process a binary field.
	// Params: value – newly allocated byte array with the binary contents.
	BinaryField(fieldInfo *FieldInfo, value []byte) error

	// StringField Process a string field; the provided byte[] value is a UTF-8 encoded string value.
	StringField(fieldInfo *FieldInfo, value []byte) error

	// IntField Process a int numeric field.
	IntField(fieldInfo *FieldInfo, value int) error

	// LongField Process a long numeric field.
	//LongField(fieldInfo *FieldInfo, value int64) error

	// FloatField Process a float numeric field.
	FloatField(fieldInfo *FieldInfo, value float64) error

	// DoubleField Process a double numeric field.
	//DoubleField(fieldInfo *FieldInfo, value []byte) error

	// NeedsField Hook before processing a field. Before a field is processed, this method is invoked so that
	// subclasses can return a StoredFieldVisitor.Status representing whether they need that particular field
	// or not, or to stop processing entirely.
	NeedsField(fieldInfo *FieldInfo) StoredFieldVisitorStatus
}

type StoredFieldVisitorStatus int

const (
	SFV_STATUS_YES = StoredFieldVisitorStatus(iota)
	SFV_STATUS_NO
	SFV_STATUS_STOP
)
