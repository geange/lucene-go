package document

import (
	"errors"
	"github.com/geange/lucene-go/core/analysis"
	"github.com/geange/lucene-go/core/tokenattributes"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util"
	"io"
)

type Field struct {
	_type      types.IndexableFieldType
	name       string
	fieldsData any

	// Pre-analyzed tokenStream for indexed fields; this is separate from fieldsData because
	// you are allowed to have both; eg maybe field has a String value but you customize how it's tokenized
	tokenStream analysis.TokenStream
}

// NewFieldV1 Expert: creates a field with no initial value. Intended only for custom Field subclasses.
// Params: 	name – field name
//			types – field types
// Throws: 	IllegalArgumentException – if either the name or types is null.
func NewFieldV1(name string, _type types.IndexableFieldType) *Field {
	return &Field{
		_type: _type,
		name:  name,
	}
}

func NewFieldWithAny(name string, _type types.IndexableFieldType, data any) *Field {
	field := &Field{
		_type:      _type,
		name:       name,
		fieldsData: data,
	}
	return field
}

// NewFieldV2 Create field with Reader value.
// Params: 	name – field name
//			reader – reader value
//			types – field types
// Throws: 	IllegalArgumentException – if either the name or types is null, or if the field's types is stored(), or if tokenized() is false.
//			NullPointerException – if the reader is null
func NewFieldV2(name string, reader io.Reader, _type types.IndexableFieldType) *Field {
	return &Field{
		_type:       _type,
		name:        name,
		fieldsData:  reader,
		tokenStream: nil,
	}
}

func NewFieldV3(name string, tokenStream analysis.TokenStream, _type types.IndexableFieldType) *Field {
	return &Field{
		_type:       _type,
		name:        name,
		fieldsData:  nil,
		tokenStream: tokenStream,
	}
}

// NewFieldV4 Create field with binary value.
// NOTE: the provided byte[] is not copied so be sure not to change it until you're done with this field.
// Params: 	name – field name
//			value – byte array pointing to binary content (not copied)
//			types – field types
// Throws: 	IllegalArgumentException – if the field name, value or types is null, or the field's types is indexed().
func NewFieldV4(name string, value []byte, _type types.IndexableFieldType) *Field {
	return &Field{
		_type:       _type,
		name:        name,
		fieldsData:  value,
		tokenStream: nil,
	}
}

func NewFieldV5(name string, value string, _type types.IndexableFieldType) *Field {
	return &Field{
		_type:       _type,
		name:        name,
		fieldsData:  value,
		tokenStream: nil,
	}
}

func (r *Field) TokenStream(analyzer analysis.Analyzer, reuse analysis.TokenStream) (analysis.TokenStream, error) {
	if r.FieldType().IndexOptions() == types.INDEX_OPTIONS_NONE {
		return nil, nil
	}

	if !r.FieldType().Tokenized() {
		switch r.Value().(type) {
		case string:
			stream, ok := reuse.(*StringTokenStream)
			if !ok {
				var err error
				stream, err = NewStringTokenStream(util.NewAttributeSource())
				if err != nil {
					return nil, err
				}
			}
			stream.SetValue(r.Value().(string))
			return stream, nil
		case []byte:
			stream, ok := reuse.(*BinaryTokenStream)
			if !ok {
				var err error
				stream, err = NewBinaryTokenStream(util.NewAttributeSource())
				if err != nil {
					return nil, err
				}
			}
			stream.SetValue(r.Value().([]byte))
			return stream, nil
		default:
			return nil, errors.New("Non-Tokenized Fields must have a String value")
		}
	}

	if r.tokenStream != nil {
		return r.tokenStream, nil
	}

	switch r.Value().(type) {
	case string:
		return analyzer.TokenStreamByString(r.name, r.Value().(string))
	case io.Reader:
		return analyzer.TokenStreamByReader(r.name, r.Value().(io.Reader))
	default:
		return nil, errors.New("field must have either TokenStream, String, Reader or Number value")
	}
}

func (r *Field) Name() string {
	return r.name
}

func (r *Field) FieldType() types.IndexableFieldType {
	return r._type
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

func NewStringTokenStream(source *util.AttributeSource) (*StringTokenStream, error) {
	stream := &StringTokenStream{
		source:          source,
		termAttribute:   nil,
		offsetAttribute: nil,
		used:            false,
		value:           "",
	}
	termAttribute, ok := source.Get(tokenattributes.ClassCharTerm)
	if !ok {
		return nil, errors.New("PackedTokenAttribute not exist")
	}
	stream.termAttribute = termAttribute.(*tokenattributes.PackedTokenAttributeImpl)

	offsetAttribute, ok := source.Get(tokenattributes.ClassOffset)
	if !ok {
		return nil, errors.New("PackedTokenAttribute not exist")
	}
	stream.offsetAttribute = offsetAttribute.(*tokenattributes.PackedTokenAttributeImpl)

	return stream, nil
}

type StringTokenStream struct {
	source          *util.AttributeSource
	termAttribute   tokenattributes.CharTermAttribute
	offsetAttribute tokenattributes.OffsetAttribute
	used            bool
	value           string
}

func (s *StringTokenStream) GetAttributeSource() *util.AttributeSource {
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
	_ analysis.TokenStream = &BinaryTokenStream{}
)

func NewBinaryTokenStream(source *util.AttributeSource) (*BinaryTokenStream, error) {
	stream := &BinaryTokenStream{
		source:   source,
		bytesAtt: nil,
		used:     true,
		value:    nil,
	}
	att, ok := source.Get(tokenattributes.ClassBytesTerm)
	if !ok {
		return nil, errors.New("BytesTermAttribute not exist")
	}
	stream.bytesAtt = att.(*tokenattributes.BytesTermAttributeImpl)
	return stream, nil
}

type BinaryTokenStream struct {
	source   *util.AttributeSource
	bytesAtt *tokenattributes.BytesTermAttributeImpl
	used     bool
	value    []byte
}

func (b *BinaryTokenStream) GetAttributeSource() *util.AttributeSource {
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
