package document

import (
	"github.com/geange/lucene-go/core/types"
)

var _ StoredFieldVisitor = &DocumentStoredFieldVisitor{}

type DocumentStoredFieldVisitor struct {
	doc         *Document
	fieldsToAdd map[string]struct{}
}

func NewDocumentStoredFieldVisitor() *DocumentStoredFieldVisitor {
	return &DocumentStoredFieldVisitor{
		doc:         NewDocument(),
		fieldsToAdd: nil,
	}
}

func NewDocumentStoredFieldVisitorV1(fieldsToAdd map[string]struct{}) *DocumentStoredFieldVisitor {
	return &DocumentStoredFieldVisitor{
		doc:         NewDocument(),
		fieldsToAdd: fieldsToAdd,
	}
}

func NewDocumentStoredFieldVisitorV2(fields ...string) *DocumentStoredFieldVisitor {
	fieldsToAdd := make(map[string]struct{}, len(fields))
	for _, field := range fields {
		fieldsToAdd[field] = struct{}{}
	}
	return NewDocumentStoredFieldVisitorV1(fieldsToAdd)
}

func (r *DocumentStoredFieldVisitor) GetDocument() *Document {
	return r.doc
}

func (r *DocumentStoredFieldVisitor) BinaryField(fieldInfo *types.FieldInfo, value []byte) error {
	r.doc.Add(NewStoredFieldV3(fieldInfo.Name(), value))
	return nil
}

func (r *DocumentStoredFieldVisitor) StringField(fieldInfo *types.FieldInfo, value []byte) error {
	ft := NewFieldTypeV1(TextFieldStored)
	err := ft.SetStoreTermVectors(fieldInfo.HasVectors())
	if err != nil {
		return err
	}
	err = ft.SetOmitNorms(fieldInfo.OmitsNorms())
	if err != nil {
		return err
	}
	err = ft.SetIndexOptions(fieldInfo.GetIndexOptions())
	if err != nil {
		return err
	}
	r.doc.Add(NewStoredFieldV5(fieldInfo.Name(), string(value), ft))

	return nil
}

func (r *DocumentStoredFieldVisitor) Int32Field(fieldInfo *types.FieldInfo, value int32) error {
	r.doc.Add(NewStoredFieldAny(fieldInfo.Name(), value, STORED_ONLY))
	return nil
}

func (r *DocumentStoredFieldVisitor) Int64Field(fieldInfo *types.FieldInfo, value int64) error {
	r.doc.Add(NewStoredFieldAny(fieldInfo.Name(), value, STORED_ONLY))
	return nil
}

func (r *DocumentStoredFieldVisitor) Float32Field(fieldInfo *types.FieldInfo, value float32) error {
	r.doc.Add(NewStoredFieldAny(fieldInfo.Name(), value, STORED_ONLY))
	return nil
}

func (r *DocumentStoredFieldVisitor) Float64Field(fieldInfo *types.FieldInfo, value float64) error {
	r.doc.Add(NewStoredFieldAny(fieldInfo.Name(), value, STORED_ONLY))
	return nil
}

func (r *DocumentStoredFieldVisitor) NeedsField(fieldInfo *types.FieldInfo) (STORED_FIELD_VISITOR_STATUS, error) {
	if r.fieldsToAdd == nil {
		return STORED_FIELD_VISITOR_YES, nil
	}

	_, ok := r.fieldsToAdd[fieldInfo.Name()]
	if ok {
		return STORED_FIELD_VISITOR_YES, nil
	}
	return STORED_FIELD_VISITOR_NO, nil
}
