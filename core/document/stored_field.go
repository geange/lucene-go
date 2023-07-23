package document

type StoredField struct {
	*Field
}

var STORED_ONLY = newStoreFieldType()

func newStoreFieldType() *FieldType {
	fieldType := NewFieldType()
	_ = fieldType.SetStored(true)
	fieldType.Freeze()
	return fieldType
}

func NewStoredField[T Value](name string, value T) *StoredField {
	return &StoredField{
		NewField(name, value, STORED_ONLY),
	}
}

func NewStoredFieldWithType[T Value](name string, value T, _type IndexableFieldType) *StoredField {
	return &StoredField{
		NewField(name, value, _type),
	}
}
