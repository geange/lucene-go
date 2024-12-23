package document

// StoredFieldVisitor
// Expert: provides a low-level means of accessing the stored field values in an index.
// See IndexReader.document(int, StoredFieldVisitor).
//
// NOTE: a StoredFieldVisitor implementation should not try to load or visit other stored documents in
// the same reader because the implementation of stored fields for most codecs is not reentrant and you
// will see strange exceptions as a result.
//
// See DocStoredFieldVisitor, which is a StoredFieldVisitor that builds the Document containing
// all stored fields. This is used by IndexReader.document(int).
// lucene.experimental
type StoredFieldVisitor interface {

	// BinaryField Process a binary field.
	// @param value newly allocated byte array with the binary contents.
	BinaryField(fieldInfo *FieldInfo, value []byte) error

	// StringField Process a string field; the provided byte[] value is a UTF-8 encoded string value.
	StringField(fieldInfo *FieldInfo, value []byte) error

	// Int32Field Process a int numeric field.
	Int32Field(fieldInfo *FieldInfo, value int32) error

	// Int64Field Process a long numeric field.
	Int64Field(fieldInfo *FieldInfo, value int64) error

	// Float32Field Process a float numeric field.
	Float32Field(fieldInfo *FieldInfo, value float32) error

	// Float64Field Process a double numeric field.
	Float64Field(fieldInfo *FieldInfo, value float64) error

	// NeedsField Hook before processing a field. Before a field is processed, this method is invoked
	//so that subclasses can return a StoredFieldVisitor.Status representing
	// whether they need that particular field or not, or to stop processing entirely.
	NeedsField(fieldInfo *FieldInfo) (STORED_FIELD_VISITOR_STATUS, error)
}

type STORED_FIELD_VISITOR_STATUS int

const (
	STORED_FIELD_VISITOR_YES  = STORED_FIELD_VISITOR_STATUS(iota) // YES: the field should be visited.
	STORED_FIELD_VISITOR_NO                                       // NO: don't visit this field, but continue processing fields for this document.
	STORED_FIELD_VISITOR_STOP                                     // STOP: don't visit this field and stop processing any other fields for this document.
)

var _ StoredFieldVisitor = &DocStoredFieldVisitor{}

type DocStoredFieldVisitor struct {
	doc         *Document
	fieldsToAdd map[string]struct{}
}

func NewDocumentStoredFieldVisitor(fields ...string) *DocStoredFieldVisitor {
	fieldsToAdd := make(map[string]struct{}, len(fields))
	for _, field := range fields {
		fieldsToAdd[field] = struct{}{}
	}
	return newDocumentStoredFieldVisitor(fieldsToAdd)
}

func newDocumentStoredFieldVisitor(fieldsToAdd map[string]struct{}) *DocStoredFieldVisitor {
	return &DocStoredFieldVisitor{
		doc:         NewDocument(),
		fieldsToAdd: fieldsToAdd,
	}
}

func (r *DocStoredFieldVisitor) GetDocument() *Document {
	return r.doc
}

func (r *DocStoredFieldVisitor) BinaryField(fieldInfo *FieldInfo, value []byte) error {
	field := NewStoredField[[]byte](fieldInfo.Name(), value)
	r.doc.Add(field)
	return nil
}

func (r *DocStoredFieldVisitor) StringField(fieldInfo *FieldInfo, value []byte) error {
	ft := NewFieldTypeFrom(textFieldStored)
	if err := ft.SetStoreTermVectors(fieldInfo.HasVectors()); err != nil {
		return err
	}
	if err := ft.SetOmitNorms(fieldInfo.OmitsNorms()); err != nil {
		return err
	}
	if err := ft.SetIndexOptions(fieldInfo.GetIndexOptions()); err != nil {
		return err
	}

	field := NewStoredFieldWithType(fieldInfo.Name(), string(value), ft)
	r.doc.Add(field)
	return nil
}

func (r *DocStoredFieldVisitor) Int32Field(fieldInfo *FieldInfo, value int32) error {
	r.doc.Add(NewStoredFieldWithType(fieldInfo.Name(), value, STORED_ONLY))
	return nil
}

func (r *DocStoredFieldVisitor) Int64Field(fieldInfo *FieldInfo, value int64) error {
	r.doc.Add(NewStoredFieldWithType(fieldInfo.Name(), value, STORED_ONLY))
	return nil
}

func (r *DocStoredFieldVisitor) Float32Field(fieldInfo *FieldInfo, value float32) error {
	r.doc.Add(NewStoredFieldWithType(fieldInfo.Name(), value, STORED_ONLY))
	return nil
}

func (r *DocStoredFieldVisitor) Float64Field(fieldInfo *FieldInfo, value float64) error {
	r.doc.Add(NewStoredFieldWithType(fieldInfo.Name(), value, STORED_ONLY))
	return nil
}

func (r *DocStoredFieldVisitor) NeedsField(fieldInfo *FieldInfo) (STORED_FIELD_VISITOR_STATUS, error) {
	if len(r.fieldsToAdd) == 0 {
		return STORED_FIELD_VISITOR_YES, nil
	}

	if _, ok := r.fieldsToAdd[fieldInfo.Name()]; ok {
		return STORED_FIELD_VISITOR_YES, nil
	}
	return STORED_FIELD_VISITOR_NO, nil
}
