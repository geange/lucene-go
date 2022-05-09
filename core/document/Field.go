package document

import (
	"errors"
	"github.com/geange/lucene-go/core/analysis"
	"github.com/geange/lucene-go/core/index"
	"io"
)

type Field struct {
	_type      index.IndexAbleFieldType
	_fType     FieldValueType
	name       string
	fieldsData interface{}

	// Pre-analyzed tokenStream for indexed fields; this is separate from fieldsData because
	// you are allowed to have both; eg maybe field has a String value but you customize how it's tokenized
	tokenStream analysis.TokenStream
}

func (f *Field) TokenStream(analyzer analysis.Analyzer, reuse analysis.TokenStream) (analysis.TokenStream, error) {
	if f.FieldType().IndexOptions() == index.INDEX_OPTIONS_NONE {
		return nil, nil
	}

	if !f.FieldType().Tokenized() {
		switch f.FType() {
		case FVString:
		case FVBinary:
		default:
			return nil, errors.New("Non-Tokenized Fields must have a String value")
		}
	}

	if f.tokenStream != nil {
		return f.tokenStream, nil
	}

	switch f.FType() {
	case FVString:
		return analyzer.TokenStreamByString(f.name, f.Value().(string))
	case FVReader:
		return analyzer.TokenStreamByReader(f.name, f.Value().(io.Reader))
	default:
		return nil, errors.New("field must have either TokenStream, String, Reader or Number value")
	}
}

func (f *Field) Name() string {
	return f.name
}

func (f *Field) FieldType() index.IndexAbleFieldType {
	return f._type
}

func (f *Field) FType() FieldValueType {
	return f._fType
}

func (f *Field) Value() interface{} {
	return f.fieldsData
}

type StringTokenStream struct {
}
