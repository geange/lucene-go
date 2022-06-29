package document

import (
	"github.com/geange/lucene-go/core"
	"github.com/geange/lucene-go/core/index"
)

type DocumentStoredFieldVisitor struct {
	doc         *Document
	fieldsToAdd map[string]struct{}
}

func NewDocumentStoredFieldVisitor() *DocumentStoredFieldVisitor {
	return &DocumentStoredFieldVisitor{fieldsToAdd: make(map[string]struct{})}
}

func NewDocumentStoredFieldVisitorV1(fieldsToAdd map[string]struct{}) *DocumentStoredFieldVisitor {
	return &DocumentStoredFieldVisitor{fieldsToAdd: fieldsToAdd}
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

func (r *DocumentStoredFieldVisitor) BinaryField(fieldInfo *index.FieldInfo, value []byte) error {
	r.doc.Add(core.NewStoredFieldV3(fieldInfo.Name, value))
	return nil
}

func (r *DocumentStoredFieldVisitor) StringField(fieldInfo *index.FieldInfo, value []byte) error {
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
	r.doc.Add(core.NewStoredFieldV5(fieldInfo.Name, string(value), ft))

	return nil
}

func (r *DocumentStoredFieldVisitor) IntField(fieldInfo *index.FieldInfo, value int) error {
	r.doc.Add(core.NewStoredFieldWithInt(fieldInfo.Name, value, core.TYPE))
	return nil
}

func (r *DocumentStoredFieldVisitor) FloatField(fieldInfo *index.FieldInfo, value float64) error {
	r.doc.Add(core.NewStoredFieldWithFloat(fieldInfo.Name, value, core.TYPE))
	return nil
}

func (r *DocumentStoredFieldVisitor) NeedsField(fieldInfo *index.FieldInfo) core.StoredFieldVisitorStatus {
	_, ok := r.fieldsToAdd[fieldInfo.Name]
	if ok {
		return core.SFV_STATUS_YES
	}
	return core.SFV_STATUS_NO
}
