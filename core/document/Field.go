package document

import "github.com/geange/lucene-go/core/analysis"

type Field struct {
	_type      IndexAbleFieldType
	_ftype     FieldValueType
	name       string
	fieldsData interface{}

	// Pre-analyzed tokenStream for indexed fields; this is separate from fieldsData because
	// you are allowed to have both; eg maybe field has a String value but you customize how it's tokenized
	tokenStream analysis.TokenStream
}

func (f *Field) Name() string {
	return f.name
}

func (f *Field) FieldType() IndexAbleFieldType {
	return f._type
}

func (f *Field) FType() FieldValueType {
	return f._ftype
}

func (f *Field) Value() interface{} {
	return f.fieldsData
}
