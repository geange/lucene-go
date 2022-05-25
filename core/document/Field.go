package document

import (
	"errors"
	"github.com/geange/lucene-go/core"
	"github.com/geange/lucene-go/core/analysis/tokenattributes"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/util"
	"io"
)

type Field struct {
	_type      index.IndexAbleFieldType
	_fType     FieldValueType
	name       string
	fieldsData interface{}

	// Pre-analyzed tokenStream for indexed fields; this is separate from fieldsData because
	// you are allowed to have both; eg maybe field has a String value but you customize how it's tokenized
	tokenStream core.TokenStream
}

func (f *Field) TokenStream(analyzer core.Analyzer, reuse core.TokenStream) (core.TokenStream, error) {
	if f.FieldType().IndexOptions() == index.INDEX_OPTIONS_NONE {
		return nil, nil
	}

	if !f.FieldType().Tokenized() {
		switch f.FType() {
		case FVString:
			stream, ok := reuse.(*StringTokenStream)
			if !ok {
				var err error
				stream, err = NewStringTokenStream(util.NewAttributeSource())
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
				stream, err = NewBinaryTokenStream(util.NewAttributeSource())
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

func (f *Field) FieldType() index.IndexAbleFieldType {
	return f._type
}

func (f *Field) FType() FieldValueType {
	return f._fType
}

func (f *Field) Value() interface{} {
	return f.fieldsData
}

var (
	_ core.TokenStream = &StringTokenStream{}
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
	stream.termAttribute = termAttribute.(*core.PackedTokenAttributeImpl)

	offsetAttribute, ok := source.Get(tokenattributes.ClassOffset)
	if !ok {
		return nil, errors.New("PackedTokenAttribute not exist")
	}
	stream.offsetAttribute = offsetAttribute.(*core.PackedTokenAttributeImpl)

	return stream, nil
}

type StringTokenStream struct {
	source          *util.AttributeSource
	termAttribute   tokenattributes.CharTermAttribute
	offsetAttribute tokenattributes.OffsetAttribute
	used            bool
	value           string
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
	_ core.TokenStream = &BinaryTokenStream{}
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
	stream.bytesAtt = att.(*core.BytesTermAttributeImpl)
	return stream, nil
}

type BinaryTokenStream struct {
	source   *util.AttributeSource
	bytesAtt *core.BytesTermAttributeImpl
	used     bool
	value    []byte
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
