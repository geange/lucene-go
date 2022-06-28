package core

import (
	"errors"
	"io"
)

type Field struct {
	_type      IndexableFieldType
	_fType     FieldValueType
	name       string
	fieldsData any

	// Pre-analyzed tokenStream for indexed fields; this is separate from fieldsData because
	// you are allowed to have both; eg maybe field has a String value but you customize how it's tokenized
	tokenStream TokenStream
}

// NewFieldV1 Expert: creates a field with no initial value. Intended only for custom Field subclasses.
// Params: 	name – field name
//			type – field type
// Throws: 	IllegalArgumentException – if either the name or type is null.
func NewFieldV1(name string, _type IndexableFieldType) *Field {
	return &Field{
		_type: _type,
		name:  name,
	}
}

func NewFieldWithAny(name string, _type IndexableFieldType, data any) *Field {
	field := &Field{
		_type:      _type,
		name:       name,
		fieldsData: data,
	}
	switch data.(type) {
	case int, float64:
		field._fType = FVNumeric
	case string:
		field._fType = FVString
	case []byte:
		field._fType = FVBinary
	}
	return field
}

// NewFieldV2 Create field with Reader value.
// Params: 	name – field name
//			reader – reader value
//			type – field type
// Throws: 	IllegalArgumentException – if either the name or type is null, or if the field's type is stored(), or if tokenized() is false.
//			NullPointerException – if the reader is null
func NewFieldV2(name string, reader io.Reader, _type IndexableFieldType) *Field {
	return &Field{
		_type:       _type,
		_fType:      FVReader,
		name:        name,
		fieldsData:  reader,
		tokenStream: nil,
	}
}

func NewFieldV3(name string, tokenStream TokenStream, _type IndexableFieldType) *Field {
	return &Field{
		_type:       _type,
		_fType:      FVTokenStream,
		name:        name,
		fieldsData:  nil,
		tokenStream: tokenStream,
	}
}

// NewFieldV4 Create field with binary value.
// NOTE: the provided byte[] is not copied so be sure not to change it until you're done with this field.
// Params: 	name – field name
//			value – byte array pointing to binary content (not copied)
//			type – field type
// Throws: 	IllegalArgumentException – if the field name, value or type is null, or the field's type is indexed().
func NewFieldV4(name string, value []byte, _type IndexableFieldType) *Field {
	return &Field{
		_type:       _type,
		_fType:      FVBinary,
		name:        name,
		fieldsData:  value,
		tokenStream: nil,
	}
}

func NewFieldV5(name string, value string, _type IndexableFieldType) *Field {
	return &Field{
		_type:       _type,
		_fType:      FVString,
		name:        name,
		fieldsData:  value,
		tokenStream: nil,
	}
}

func (f *Field) TokenStream(analyzer Analyzer, reuse TokenStream) (TokenStream, error) {
	if f.FieldType().IndexOptions() == INDEX_OPTIONS_NONE {
		return nil, nil
	}

	if !f.FieldType().Tokenized() {
		switch f.FType() {
		case FVString:
			stream, ok := reuse.(*StringTokenStream)
			if !ok {
				var err error
				stream, err = NewStringTokenStream(NewAttributeSource())
				if err != nil {
					return nil, err
				}
			}
			stream.SetValue(f.Value().(string))
			return stream, nil
		case FVBinary:
			stream, ok := reuse.(*BinaryTokenStream)
			if !ok {
				var err error
				stream, err = NewBinaryTokenStream(NewAttributeSource())
				if err != nil {
					return nil, err
				}
			}
			stream.SetValue(f.Value().([]byte))
			return stream, nil
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

func (f *Field) FieldType() IndexableFieldType {
	return f._type
}

func (f *Field) FType() FieldValueType {
	return f._fType
}

func (f *Field) Value() interface{} {
	return f.fieldsData
}

var (
	_ TokenStream = &StringTokenStream{}
)

func NewStringTokenStream(source *AttributeSource) (*StringTokenStream, error) {
	stream := &StringTokenStream{
		source:          source,
		termAttribute:   nil,
		offsetAttribute: nil,
		used:            false,
		value:           "",
	}
	termAttribute, ok := source.Get(ClassCharTerm)
	if !ok {
		return nil, errors.New("PackedTokenAttribute not exist")
	}
	stream.termAttribute = termAttribute.(*PackedTokenAttributeImpl)

	offsetAttribute, ok := source.Get(ClassOffset)
	if !ok {
		return nil, errors.New("PackedTokenAttribute not exist")
	}
	stream.offsetAttribute = offsetAttribute.(*PackedTokenAttributeImpl)

	return stream, nil
}

type StringTokenStream struct {
	source          *AttributeSource
	termAttribute   CharTermAttribute
	offsetAttribute OffsetAttribute
	used            bool
	value           string
}

func (s *StringTokenStream) GetAttributeSource() *AttributeSource {
	return s.source
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
	_ TokenStream = &BinaryTokenStream{}
)

func NewBinaryTokenStream(source *AttributeSource) (*BinaryTokenStream, error) {
	stream := &BinaryTokenStream{
		source:   source,
		bytesAtt: nil,
		used:     true,
		value:    nil,
	}
	att, ok := source.Get(ClassBytesTerm)
	if !ok {
		return nil, errors.New("BytesTermAttribute not exist")
	}
	stream.bytesAtt = att.(*BytesTermAttributeImpl)
	return stream, nil
}

type BinaryTokenStream struct {
	source   *AttributeSource
	bytesAtt *BytesTermAttributeImpl
	used     bool
	value    []byte
}

func (b *BinaryTokenStream) GetAttributeSource() *AttributeSource {
	return b.source
}

func (b *BinaryTokenStream) IncrementToken() (bool, error) {
	if b.used {
		return false, nil
	}

	err := b.source.Clear()
	if err != nil {
		return false, err
	}

	if err := b.bytesAtt.SetBytesRef(b.value); err != nil {
		return false, err
	}
	b.used = true
	return true, nil
}

func (b *BinaryTokenStream) End() error {
	return b.source.Clear()
}

func (b *BinaryTokenStream) Reset() error {
	b.used = false
	return nil
}

func (b *BinaryTokenStream) Close() error {
	b.value = nil
	return nil
}

func (b *BinaryTokenStream) SetValue(value []byte) {
	b.value = value
}
