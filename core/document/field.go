package document

import (
	"errors"
	"github.com/geange/lucene-go/core/analysis"
	"github.com/geange/lucene-go/core/tokenattr"
	"io"
)

// Field
// * TextField: Reader or String indexed for full-text search
// * StringField: String indexed verbatim as a single token
// * IntPoint: int indexed for exact/range queries.
// * LongPoint: long indexed for exact/range queries.
// * FloatPoint: float indexed for exact/range queries.
// * DoublePoint: double indexed for exact/range queries.
// * SortedDocValuesField: byte[] indexed column-wise for sorting/faceting
// * SortedSetDocValuesField: SortedSet<byte[]> indexed column-wise for sorting/faceting
// * NumericDocValuesField: long indexed column-wise for sorting/faceting
// * SortedNumericDocValuesField: SortedSet<long> indexed column-wise for sorting/faceting
// * StoredField: Stored-only value for retrieving in summary results
type Field struct {
	iType      IndexableFieldType
	vType      FieldValueType
	name       string
	fieldsData any

	// Pre-analyzed tokenStream for indexed fields; this is separate from fieldsData because
	// you are allowed to have both; eg maybe field has a String value but you customize how it's tokenized
	tokenStream analysis.TokenStream
}

// NewFieldV1
// Expert: creates a field with no initial value. Intended only for custom Field subclasses.
// name: field name
// types: field types
// Throws: 	IllegalArgumentException – if either the name or types is null.
func NewFieldV1(name string, _type IndexableFieldType) *Field {
	return &Field{
		iType: _type,
		name:  name,
	}
}

// NewFieldV2
// Create field with Reader value.
// name: field name
// reader: reader value
// types: field types
//
// Throws: 	IllegalArgumentException – if either the name or types is null,
// or if the field's types is stored(), or if tokenized() is false.
//
// NullPointerException – if the reader is null
func NewFieldV2(name string, reader io.Reader, _type IndexableFieldType) *Field {
	return &Field{
		iType:       _type,
		name:        name,
		fieldsData:  reader,
		tokenStream: nil,
	}
}

func NewFieldV3(name string, tokenStream analysis.TokenStream, _type IndexableFieldType) *Field {
	return &Field{
		iType:       _type,
		name:        name,
		fieldsData:  nil,
		tokenStream: tokenStream,
	}
}

// NewFieldV4 Create field with binary value.
// NOTE: the provided byte[] is not copied so be sure not to change it until you're done with this field.
// Params: 	name – field name
//
//	value – byte array pointing to binary content (not copied)
//	types – field types
//
// Throws: 	IllegalArgumentException – if the field name, value or types is null, or the field's types is indexed().

// NewField Create field with []byte or string
func NewField[T Value](name string, value T, _type IndexableFieldType) *Field {
	return &Field{
		iType:       _type,
		name:        name,
		fieldsData:  value,
		tokenStream: nil,
	}
}

func (r *Field) TokenStream(analyzer analysis.Analyzer, reuse analysis.TokenStream) (analysis.TokenStream, error) {
	if r.FieldType().IndexOptions() == INDEX_OPTIONS_NONE {
		return nil, nil
	}

	if !r.FieldType().Tokenized() {
		switch r.ValueType() {
		case FieldValueString:
			stringValue, _ := r.StringValue()
			stream, ok := reuse.(*StringTokenStream)
			if !ok {
				var err error
				stream, err = NewStringTokenStream(tokenattr.NewAttributeSource())
				if err != nil {
					return nil, err
				}
			}
			stream.SetValue(stringValue)
			return stream, nil
		case FieldValueBytes:
			bytesValue, _ := r.BytesValue()
			stream, ok := reuse.(*BinaryTokenStream)
			if !ok {
				var err error
				stream, err = NewBinaryTokenStream(tokenattr.NewAttributeSource())
				if err != nil {
					return nil, err
				}
			}
			stream.SetValue(bytesValue)
			return stream, nil
		default:
			return nil, errors.New("Non-Tokenized Fields must have a String value")
		}
	}

	if r.tokenStream != nil {
		return r.tokenStream, nil
	}

	switch r.ValueType() {
	case FieldValueString:
		value, _ := r.StringValue()
		return analyzer.TokenStreamByString(r.name, value)
	case FieldValueReader:
		reader, _ := r.ReaderValue()
		return analyzer.TokenStreamByReader(r.name, reader)
	default:
		return nil, errors.New("field must have either TokenStream, String, Reader or Number value")
	}
}

func (r *Field) Name() string {
	return r.name
}

func (r *Field) FieldType() IndexableFieldType {
	return r.iType
}

func (r *Field) I32Value() (int32, error) {
	switch r.fieldsData.(type) {
	case int32:
		return r.fieldsData.(int32), nil
	case int64:
		return int32(r.fieldsData.(int64)), nil
	default:
		return -1, errors.New("fieldsData is not int32")
	}
}

func (r *Field) I64Value() (int64, error) {
	switch r.fieldsData.(type) {
	case int32:
		return int64(r.fieldsData.(int32)), nil
	default:
		return -1, errors.New("fieldsData is not int32")
	}
}

func (r *Field) F32Value() (float32, error) {
	switch r.fieldsData.(type) {
	case float32:
		return r.fieldsData.(float32), nil
	default:
		return -1, errors.New("fieldsData is not float32")
	}
}

func (r *Field) F64Value() (float64, error) {
	switch r.fieldsData.(type) {
	case float64:
		return r.fieldsData.(float64), nil
	default:
		return -1, errors.New("fieldsData is not float64")
	}
}

func (r *Field) StringValue() (string, error) {
	switch r.fieldsData.(type) {
	case string:
		return r.fieldsData.(string), nil
	case []byte:
		return string(r.fieldsData.([]byte)), nil
	default:
		return "", errors.New("fieldsData is not string")
	}
}

func (r *Field) BytesValue() ([]byte, error) {
	switch r.fieldsData.(type) {
	case string:
		return []byte(r.fieldsData.(string)), nil
	case []byte:
		return r.fieldsData.([]byte), nil
	default:
		return nil, errors.New("fieldsData is not []byte")
	}
}

func (r *Field) ReaderValue() (io.Reader, error) {
	reader, ok := r.fieldsData.(io.Reader)
	if !ok {
		return nil, errors.New("fieldsData is not io.Reader")
	}
	return reader, nil
}

func (r *Field) ValueType() FieldValueType {
	switch r.fieldsData.(type) {
	case int32:
		return FieldValueI32
	case int64:
		return FieldValueI64
	case float32:
		return FieldValueF32
	case float64:
		return FieldValueF64
	case string:
		return FieldValueString
	case []byte:
		return FieldValueBytes
	case io.Reader:
		return FieldValueReader
	default:
		return FieldValueOther
	}
}

func (r *Field) Value() any {
	return r.fieldsData
}

func (r *Field) SetFloat64(value float64) {
	r.fieldsData = value
}

func (r *Field) SetIntValue(value int) {
	r.fieldsData = value
}

var (
	_ analysis.TokenStream = &StringTokenStream{}
)

func NewStringTokenStream(source *tokenattr.AttributeSource) (*StringTokenStream, error) {
	stream := &StringTokenStream{
		source:          source,
		termAttribute:   source.CharTerm(),
		offsetAttribute: source.Offset(),
		used:            false,
		value:           "",
	}
	return stream, nil
}

type StringTokenStream struct {
	source          *tokenattr.AttributeSource
	termAttribute   tokenattr.CharTermAttribute
	offsetAttribute tokenattr.OffsetAttribute
	used            bool
	value           string
}

func (t *StringTokenStream) AttributeSource() *tokenattr.AttributeSource {
	return t.source
}

func (s *StringTokenStream) IncrementToken() (bool, error) {
	if s.used {
		return false, nil
	}

	err := s.source.Clear()
	if err != nil {
		return false, err
	}

	s.termAttribute.Append(s.value)
	if err := s.offsetAttribute.SetOffset(0, len(s.value)); err != nil {
		return false, err
	}
	s.used = true
	return true, nil
}

func (s *StringTokenStream) End() error {
	finalOffset := len(s.value)
	return s.offsetAttribute.SetOffset(finalOffset, finalOffset)
}

func (s *StringTokenStream) Reset() error {
	s.used = false
	return nil
}

func (s *StringTokenStream) Close() error {
	return nil
}

func (s *StringTokenStream) SetValue(value string) {
	s.value = value
}

var (
	_ analysis.TokenStream = &BinaryTokenStream{}
)

func NewBinaryTokenStream(source *tokenattr.AttributeSource) (*BinaryTokenStream, error) {
	stream := &BinaryTokenStream{
		source:   source,
		bytesAtt: source.BytesTerm(),
		used:     true,
		value:    nil,
	}
	return stream, nil
}

type BinaryTokenStream struct {
	source   *tokenattr.AttributeSource
	bytesAtt tokenattr.BytesTermAttribute
	used     bool
	value    []byte
}

func (r *BinaryTokenStream) AttributeSource() *tokenattr.AttributeSource {
	return r.source
}

func (r *BinaryTokenStream) IncrementToken() (bool, error) {
	if r.used {
		return false, nil
	}

	err := r.source.Clear()
	if err != nil {
		return false, err
	}

	if err := r.bytesAtt.SetBytesRef(r.value); err != nil {
		return false, err
	}
	r.used = true
	return true, nil
}

func (r *BinaryTokenStream) End() error {
	return r.source.Clear()
}

func (r *BinaryTokenStream) Reset() error {
	r.used = false
	return nil
}

func (r *BinaryTokenStream) Close() error {
	r.value = nil
	return nil
}

func (r *BinaryTokenStream) SetValue(value []byte) {
	r.value = value
}
