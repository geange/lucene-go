package document

import (
	"github.com/geange/lucene-go/core/types"
	"io"
)

var (
	TextFieldStored    *FieldType
	TextFieldNotStored *FieldType
)

func init() {
	TextFieldStored = NewFieldType()
	TextFieldStored.SetIndexOptions(types.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS)
	TextFieldStored.SetTokenized(true)
	TextFieldStored.SetStored(true)
	TextFieldStored.Freeze()

	TextFieldNotStored = NewFieldType()
	TextFieldNotStored.SetIndexOptions(types.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS)
	TextFieldNotStored.SetTokenized(true)
	TextFieldNotStored.Freeze()
}

type TextField struct {
	*Field
}

func NewTextFieldByReader(name string, reader io.Reader) *TextField {
	return &TextField{NewFieldV2(name, reader, TextFieldNotStored)}
}

func NewTextFieldByString(name string, value string, stored bool) *TextField {
	_type := TextFieldStored
	if !stored {
		_type = TextFieldNotStored
	}
	return &TextField{NewFieldV5(name, value, _type)}
}

func NewTextFieldByBytes(name string, value []byte, stored bool) *TextField {
	_type := TextFieldStored
	if !stored {
		_type = TextFieldNotStored
	}
	return &TextField{NewFieldV4(name, value, _type)}
}
