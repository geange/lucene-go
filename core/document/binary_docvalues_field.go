package document

import (
	"github.com/geange/lucene-go/core/types"
)

var (
	BinaryDocValuesFieldType *FieldType
)

func init() {
	BinaryDocValuesFieldType.SetDocValuesType(types.DOC_VALUES_TYPE_BINARY)
	BinaryDocValuesFieldType.Freeze()
}

// BinaryDocValuesField Field that stores a per-document BytesRef value.
// The values are stored directly with no sharing, which is a good fit when the fields don't share (many)
// values, such as a title field. If values may be shared and sorted it's better to use SortedDocValuesField.
// Here's an example usage:
//     document.add(new BinaryDocValuesField(name, new BytesRef("hello")));
//
//If you also need to store the value, you should add a separate StoredField instance.
//See Also:
//BinaryDocValues
type BinaryDocValuesField struct {
	*Field
}

func NewBinaryDocValuesField(name string, value []byte) *BinaryDocValuesField {
	field := &BinaryDocValuesField{NewFieldV1(name, BinaryDocValuesFieldType)}
	field.fieldsData = value
	return field
}
