package document

type StoredFieldType interface {
	int32 | int64 | float32 | float64 | string | []byte
}

type StoredField[T StoredFieldType] struct {
	*Field[T]
}

var STORED_ONLY = newStoreFieldType()

func newStoreFieldType() *FieldType {
	fieldType := NewFieldType()
	_ = fieldType.SetStored(true)
	fieldType.Freeze()
	return fieldType
}

func NewStoredField[T StoredFieldType](name string, value T) *StoredField[T] {
	return &StoredField[T]{NewField(name, value, STORED_ONLY)}
}

func NewStoredFieldWithType[T StoredFieldType](name string, value T, fieldType IndexableFieldType) *StoredField[T] {
	return &StoredField[T]{NewField(name, value, fieldType)}
}
