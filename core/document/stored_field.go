package document

import (
	"github.com/geange/lucene-go/core/types"
)

type StoredField struct {
	*Field
}

var (
	TYPE = NewFieldType()
)

func NewStoredFieldV3(name string, value []byte) *StoredField {
	return &StoredField{
		NewFieldV4(name, value, TYPE),
	}
}

func NewStoredFieldV4(name string, value string) *StoredField {
	return &StoredField{
		NewFieldV5(name, value, TYPE),
	}
}

func NewStoredFieldV5(name string, value string, _type types.IndexableFieldType) *StoredField {
	return &StoredField{
		NewFieldV5(name, value, _type),
	}
}

func NewStoredFieldWithInt(name string, value int, _type types.IndexableFieldType) *StoredField {
	return &StoredField{
		NewFieldWithAny(name, _type, value),
	}
}

func NewStoredFieldWithFloat(name string, value float64, _type types.IndexableFieldType) *StoredField {
	return &StoredField{
		NewFieldWithAny(name, _type, value),
	}
}
