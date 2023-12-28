package document

import "sync"

var (
	numericDocValuesFieldTypeOnce sync.Once
	numericDocValuesFieldType     *FieldType
)

// NumericDocValuesField
// Field that stores a per-document long value for scoring, sorting or value
// retrieval. Here's an example usage:
//
//	document.Add(NewNumericDocValuesField(name, 22));
//
// If you also need to store the value, you should add a separate StoredField instance.
type NumericDocValuesField struct {
	*Field[int64]
}

// NewNumericDocValuesField
// Creates a new DocValues field with the specified 64-bit long value
// name: field name
// value: 64-bit long value or null if the existing fields value should be removed on update
func NewNumericDocValuesField(name string, value int64) NumericDocValuesField {
	numericDocValuesFieldTypeOnce.Do(func() {
		numericDocValuesFieldType = NewFieldType()
		_ = numericDocValuesFieldType.SetDocValuesType(DOC_VALUES_TYPE_NUMERIC)
		numericDocValuesFieldType.Freeze()
	})
	return NumericDocValuesField{NewField(name, value, numericDocValuesFieldType)}
}
