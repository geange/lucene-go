package document

import "sync"

var (
	stringFieldTypeOnce      sync.Once
	stringFieldTypeNotStored *FieldType
	stringFieldTypeStored    *FieldType
)

type StringField struct {
	*Field[string]
}

// NewStringField
// Creates a new textual StringField, indexing the provided String value as a single token.
// name: field name
// value: String value
// stored: true if the content should also be stored
func NewStringField(name string, value string, stored bool) *StringField {
	stringFieldTypeOnce.Do(func() {
		stringFieldTypeNotStored = NewFieldType()
		_ = stringFieldTypeNotStored.SetOmitNorms(true)
		_ = stringFieldTypeNotStored.SetIndexOptions(INDEX_OPTIONS_DOCS)
		_ = stringFieldTypeNotStored.SetTokenized(false)
		stringFieldTypeNotStored.Freeze()

		stringFieldTypeStored = NewFieldType()
		_ = stringFieldTypeStored.SetOmitNorms(true)
		_ = stringFieldTypeStored.SetIndexOptions(INDEX_OPTIONS_DOCS)
		_ = stringFieldTypeStored.SetStored(true)
		_ = stringFieldTypeStored.SetTokenized(false)
		stringFieldTypeStored.Freeze()
	})

	fieldType := stringFieldTypeStored
	if !stored {
		fieldType = stringFieldTypeNotStored
	}
	return &StringField{NewField(name, value, fieldType)}
}
