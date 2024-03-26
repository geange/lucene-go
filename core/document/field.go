package document

import (
	"errors"
	"io"
	"sync/atomic"

	"github.com/geange/lucene-go/core/analysis"
	"github.com/geange/lucene-go/core/tokenattr"
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
type Field[T any] struct {
	fieldType  IndexableFieldType
	name       string
	fieldsData T
}

// NewField
// Create field with any
func NewField[T any](name string, value T, fieldType IndexableFieldType) *Field[T] {
	field := &Field[T]{
		name:       name,
		fieldType:  fieldType,
		fieldsData: value,
	}
	return field
}

func (r *Field[T]) TokenStream(analyzer analysis.Analyzer, reuse analysis.TokenStream) (analysis.TokenStream, error) {
	if r.FieldType().IndexOptions() == INDEX_OPTIONS_NONE {
		return nil, nil
	}

	if !r.FieldType().Tokenized() {
		switch v := r.Get().(type) {
		case string:
			stream, ok := reuse.(*StringTokenStream)
			if ok {
				stream.SetValue(v)
				return stream, nil
			}

			stream, err := NewStringTokenStream(tokenattr.NewAttributeSource())
			if err != nil {
				return nil, err
			}
			stream.SetValue(v)
			return stream, nil
		case []byte:
			stream, ok := reuse.(*BinaryTokenStream)
			if ok {
				stream.SetValue(v)
				return stream, nil
			}

			stream, err := NewBinaryTokenStream(tokenattr.NewAttributeSource())
			if err != nil {
				return nil, err
			}
			stream.SetValue(v)
			return stream, nil
		default:
			return nil, errors.New("non-tokenized Fields must have a string value")
		}
	}

	fieldValue := r.Get()
	if fieldValue == nil {
		return nil, errors.New("field must have either TokenStream, String, Reader or Number value")
	}

	switch v := fieldValue.(type) {
	case string:
		return analyzer.GetTokenStreamFromText(r.name, v)
	case []rune:
		return analyzer.GetTokenStreamFromText(r.name, string(v))
	case io.Reader:
		return analyzer.GetTokenStreamFromReader(r.name, v)
	default:
		return nil, errors.New("field must have either TokenStream, String, Reader or Number value")
	}
}

func (r *Field[T]) Name() string {
	return r.name
}

func (r *Field[T]) FieldType() IndexableFieldType {
	return r.fieldType
}

func (r *Field[T]) Get() any {
	return r.fieldsData
}

func (r *Field[T]) Set(v any) error {
	if act, ok := v.(T); ok {
		r.fieldsData = act
	}
	return nil
}

func (r *Field[T]) Number() (any, bool) {
	return 0, false
}

var _ analysis.TokenStream = &StringTokenStream{}

func NewStringTokenStream(source *tokenattr.AttributeSource) (*StringTokenStream, error) {
	stream := &StringTokenStream{
		source:          source,
		termAttribute:   source.CharTerm(),
		offsetAttribute: source.Offset(),
		used:            &atomic.Bool{},
		value:           "",
	}
	return stream, nil
}

type StringTokenStream struct {
	source          *tokenattr.AttributeSource
	termAttribute   tokenattr.CharTermAttr
	offsetAttribute tokenattr.OffsetAttr
	used            *atomic.Bool
	value           string
}

func (s *StringTokenStream) AttributeSource() *tokenattr.AttributeSource {
	return s.source
}

func (s *StringTokenStream) IncrementToken() (bool, error) {
	if s.used.Load() {
		return false, nil
	}
	s.used.Store(true)

	if err := s.source.Reset(); err != nil {
		return false, err
	}

	if err := s.termAttribute.AppendString(s.value); err != nil {
		return false, err
	}
	if err := s.offsetAttribute.SetOffset(0, len(s.value)); err != nil {
		return false, err
	}

	return true, nil
}

func (s *StringTokenStream) End() error {
	finalOffset := len(s.value)
	return s.offsetAttribute.SetOffset(finalOffset, finalOffset)
}

func (s *StringTokenStream) Reset() error {
	s.used.Store(false)
	return nil
}

func (s *StringTokenStream) Close() error {
	return nil
}

func (s *StringTokenStream) SetValue(value string) {
	s.value = value
}

var _ analysis.TokenStream = &BinaryTokenStream{}

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
	bytesAtt tokenattr.BytesTermAttr
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

	if err := r.source.Reset(); err != nil {
		return false, err
	}

	if err := r.bytesAtt.SetBytes(r.value); err != nil {
		return false, err
	}
	r.used = true
	return true, nil
}

func (r *BinaryTokenStream) End() error {
	return r.source.Reset()
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
