package document

import (
	"io"
)

var (
	TextFieldStored    *FieldType
	TextFieldNotStored *FieldType
)

func init() {
	TextFieldStored = NewFieldType()
	_ = TextFieldStored.SetIndexOptions(INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS)
	_ = TextFieldStored.SetTokenized(true)
	_ = TextFieldStored.SetStored(true)
	TextFieldStored.Freeze()

	TextFieldNotStored = NewFieldType()
	_ = TextFieldNotStored.SetIndexOptions(INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS)
	_ = TextFieldNotStored.SetTokenized(true)
	TextFieldNotStored.Freeze()
}

type TextField struct {
	*Field
}

func NewTextFieldByReader(name string, reader io.Reader) *TextField {
	return &TextField{NewFieldV2(name, reader, TextFieldNotStored)}
}

func NewTextField[T Value](name string, value T, stored bool) *TextField {
	_type := TextFieldStored
	if !stored {
		_type = TextFieldNotStored
	}
	return &TextField{NewField(name, value, _type)}
}
