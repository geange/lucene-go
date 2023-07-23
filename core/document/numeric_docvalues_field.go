package document

var (
	NumericDocValuesFieldType *FieldType
)

func init() {
	NumericDocValuesFieldType = NewFieldType()
	NumericDocValuesFieldType.SetDocValuesType(DOC_VALUES_TYPE_NUMERIC)
	NumericDocValuesFieldType.Freeze()
}

// NumericDocValuesField Field that stores a per-document long value for scoring, sorting or value
// retrieval. Here's an example usage:
//
//	document.add(new NumericDocValuesField(name, 22L));
//
// If you also need to store the value, you should add a separate StoredField instance.
type NumericDocValuesField struct {
	*Field
}

// NewNumericDocValuesField Creates a new DocValues field with the specified 64-bit long value
// Params: 	name – field name
//
//	value – 64-bit long value or null if the existing fields value should be removed on update
//
// Throws: 	IllegalArgumentException – if the field name is null
func NewNumericDocValuesField(name string, value int) *NumericDocValuesField {
	field := &NumericDocValuesField{
		NewFieldV1(name, NumericDocValuesFieldType),
	}
	field.fieldsData = value
	return field
}

func (r *NumericDocValuesField) SetDoubleValue() {

}
