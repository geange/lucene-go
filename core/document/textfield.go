package document

var (
	textFieldStored    *FieldType
	textFieldNotStored *FieldType
)

func init() {
	textFieldStored = NewFieldType()
	_ = textFieldStored.SetIndexOptions(INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS)
	_ = textFieldStored.SetTokenized(true)
	_ = textFieldStored.SetStored(true)
	textFieldStored.Freeze()

	textFieldNotStored = NewFieldType()
	_ = textFieldNotStored.SetIndexOptions(INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS)
	_ = textFieldNotStored.SetTokenized(true)
	textFieldNotStored.Freeze()
}

type TextField struct {
	*Field[string]
}

func NewTextField(name string, value string, stored bool) *TextField {
	fieldType := textFieldStored
	if !stored {
		fieldType = textFieldNotStored
	}
	return &TextField{NewField(name, value, fieldType)}
}
