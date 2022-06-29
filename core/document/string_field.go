package document

import "github.com/geange/lucene-go/core/types"

var (
	StringFieldTypeNotStored *FieldType
	StringFieldTypeStored    *FieldType
)

func init() {
	StringFieldTypeNotStored = NewFieldType()
	StringFieldTypeNotStored.SetOmitNorms(true)
	StringFieldTypeNotStored.SetIndexOptions(types.INDEX_OPTIONS_DOCS)
	StringFieldTypeNotStored.SetTokenized(false)
	StringFieldTypeNotStored.Freeze()

	StringFieldTypeStored = NewFieldType()
	StringFieldTypeStored.SetOmitNorms(true)
	StringFieldTypeStored.SetIndexOptions(types.INDEX_OPTIONS_DOCS)
	StringFieldTypeStored.SetStored(true)
	StringFieldTypeStored.SetTokenized(false)
	StringFieldTypeStored.Freeze()
}

type StringField struct {
	*Field
}

// NewStringFieldByString Creates a new textual StringField, indexing the provided String value as a single token.
// Params: 	name – field name
//			value – String value
//			stored – Store.YES if the content should also be stored
// Throws: 	IllegalArgumentException – if the field name or value is null.
func NewStringFieldByString(name string, value string, stored bool) *StringField {
	_type := StringFieldTypeStored
	if !stored {
		_type = StringFieldTypeNotStored
	}
	return &StringField{NewFieldV5(name, value, _type)}
}

// NewStringFieldByBytes Creates a new binary StringField, indexing the provided binary (BytesRef) value as a
// single token.
// Params: 	name – field name
//			value – BytesRef value. The provided value is not cloned so you must not change it until the
//	 		document(s) holding it have been indexed.
//			stored – Store.YES if the content should also be stored
// Throws: 	IllegalArgumentException – if the field name or value is null.
func NewStringFieldByBytes(name string, value []byte, stored bool) *StringField {
	_type := StringFieldTypeStored
	if !stored {
		_type = StringFieldTypeNotStored
	}
	return &StringField{NewFieldV4(name, value, _type)}
}
