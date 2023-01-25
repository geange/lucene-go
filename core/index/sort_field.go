package index

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/geange/lucene-go/core/store"
	"math"
	"sync"
)

const (
	STRING_FIRST = "SortField.STRING_FIRST"
	STRING_LAST  = "SortField.STRING_LAST"
)

// SortField Stores information about how to sort documents by terms in an individual field.
// Fields must be indexed in order to sort by them.
// Created: Feb 11, 2004 1:25:29 PM
// Since: lucene 1.4
// See Also: Sort
type SortField struct {
	field        string
	_type        SortFieldType
	reverse      bool
	canUsePoints bool
	missingValue any
}

func NewSortField(field string, _type SortFieldType, reverse bool) (*SortField, error) {
	sortField := &SortField{reverse: reverse}
	if err := sortField.initFieldType(&field, _type); err != nil {
		return nil, err
	}
	return sortField, nil
}

// Sets field & type, and ensures field is not NULL unless
// type is SCORE or DOC
func (s *SortField) initFieldType(field *string, _type SortFieldType) error {
	s._type = _type
	if field == nil {
		switch _type {
		case SCORE, DOC:
		default:
			return errors.New("field can only be null when type is SCORE or DOC")
		}
	}
	s.field = *field
	return nil
}

func (s *SortField) String() string {
	buffer := new(bytes.Buffer)

	switch s._type {
	case SCORE:
		buffer.WriteString("<score>")
	case DOC:
		buffer.WriteString("<doc>")
	case STRING:
		buffer.WriteString(fmt.Sprintf(`<string: "%s">`, s.field))
	case STRING_VAL:
		buffer.WriteString(fmt.Sprintf(`<string_val: "%s">`, s.field))
	case INT:
		buffer.WriteString(fmt.Sprintf(`<int: "%s">`, s.field))
	case LONG:
		buffer.WriteString(fmt.Sprintf(`<long: "%s">`, s.field))
	case FLOAT:
		buffer.WriteString(fmt.Sprintf(`<float: "%s">`, s.field))
	case DOUBLE:
		buffer.WriteString(fmt.Sprintf(`<double: "%s">`, s.field))
	case CUSTOM:
		buffer.WriteString(fmt.Sprintf(`<custom: "%s">`, s.field))
	case REWRITEABLE:
		buffer.WriteString(fmt.Sprintf(`<rewriteable: "%s">`, s.field))
	default:
		buffer.WriteString(fmt.Sprintf(`<???: "%s">`, s.field))
	}

	if s.reverse {
		buffer.WriteString("!")
	}

	if s.missingValue != nil {
		buffer.WriteString(" missingValue=")
		buffer.WriteString(fmt.Sprintf("%v", s.missingValue))
	}

	return buffer.String()
}

// GetField Returns the name of the field. Could return null if the sort is by SCORE or DOC.
// Returns: Name of field, possibly null.
func (s *SortField) GetField() string {
	return s.field
}

// GetType Returns the type of contents in the field.
// Returns: One of the constants SCORE, DOC, STRING, INT or FLOAT.
func (s *SortField) GetType() SortFieldType {
	return s._type
}

// SetCanUsePoints For numeric sort fields, setting this field, indicates that the same numeric
// data has been indexed with two fields: doc values and points and that these fields have the
// same name. This allows to use sort optimization and skip non-competitive documents.
func (s *SortField) SetCanUsePoints() {
	s.canUsePoints = true
}

func (s *SortField) GetCanUsePoints() bool {
	return s.canUsePoints
}

// NeedsScores Whether the relevance score is needed to sort documents.
func (s *SortField) NeedsScores() bool {
	return s._type == SCORE
}

// GetMissingValue Return the value to use for documents that don't have a value. A value of null indicates that default should be used.
func (s *SortField) GetMissingValue() any {
	return s.missingValue
}

// SetMissingValue Set the value to use for documents that don't have a value.
func (s *SortField) SetMissingValue(missingValue any) {
	s.missingValue = missingValue
}

const (
	ProviderName = "SortField"
)

// GetIndexSorter Returns an IndexSorter used for sorting index segments by this SortField.
// If the SortField cannot be used for index sorting (for example, if it uses scores or other
// query-dependent values) then this method should return null SortFields that implement
// this method should also implement a companion SortFieldProvider to serialize and deserialize
// the sort in index segment headers
// lucene.experimental
func (s *SortField) GetIndexSorter() IndexSorter {
	switch s._type {
	case STRING:
		return nil
	case INT:
		return NewIntSorter(ProviderName, s.missingValue.(int32), s.reverse, func(reader LeafReader) (NumericDocValues, error) {
			return nil, nil
		})
	case LONG:
		return NewLongSorter(ProviderName, s.missingValue.(int64), s.reverse, func(reader LeafReader) (NumericDocValues, error) {
			return nil, nil
		})
	case DOUBLE:
		return nil
	case FLOAT:
		return nil
	default:
		return nil
	}
}

func (s *SortField) serialize(out store.DataOutput) error {
	out.WriteString(s.field)
	out.WriteString(s._type.String())

	out.WriteUint32(func() uint32 {
		if s.reverse {
			return 1
		}
		return 0
	}())

	if s.missingValue == nil {
		return out.WriteUint32(0)
	}

	out.WriteUint32(1)
	switch s._type {
	case SCORE:
		switch s.missingValue.(string) {
		case STRING_FIRST:
			out.WriteUint32(1)
		case STRING_LAST:
			out.WriteUint32(0)
		default:
			return fmt.Errorf("cannot serialize missing value of %v for type STRING", s.missingValue)
		}

		return nil
	case INT:
		return out.WriteUint32(uint32(s.missingValue.(int32)))
	case LONG:
		return out.WriteUint64(uint64(s.missingValue.(int64)))
	case FLOAT:
		return out.WriteUint32(math.Float32bits(s.missingValue.(float32)))
	case DOUBLE:
		return out.WriteUint64(math.Float64bits(s.missingValue.(float64)))
	default:
		return fmt.Errorf("cannot serialize SortField of type %s", s._type)
	}
}

// SortFieldType Specifies the type of the terms to be sorted, or special types such as CUSTOM
type SortFieldType int

func (s SortFieldType) String() string {
	switch s {
	case SCORE:
		return "SCORE"
	case DOC:
		return "DOC"
	case STRING:
		return "STRING"
	case INT:
		return "INT"
	case FLOAT:
		return "FLOAT"
	case LONG:
		return "LONG"
	case DOUBLE:
		return "DOUBLE"
	case CUSTOM:
		return "CUSTOM"
	case STRING_VAL:
		return "STRING_VAL"
	case REWRITEABLE:
		return "REWRITEABLE"
	default:
		return ""
	}
}

const (
	// SCORE // Sort by document score (relevance).
	// Sort values are Float and higher values are at the front.
	SCORE = SortFieldType(iota)

	// DOC Sort by document number (index order).
	// Sort values are Integer and lower values are at the front.
	DOC

	// STRING Sort using term values as Strings.
	// Sort values are String and lower values are at the front.
	STRING

	// INT Sort using term values as encoded Integers.
	// Sort values are Integer and lower values are at the front.
	INT

	// FLOAT Sort using term values as encoded Floats.
	// Sort values are Float and lower values are at the front.
	FLOAT

	// LONG Sort using term values as encoded Longs.
	// Sort values are Long and lower values are at the front.
	LONG

	// DOUBLE Sort using term values as encoded Doubles.
	// Sort values are Double and lower values are at the front.
	DOUBLE

	// CUSTOM Sort using a custom Comparator.
	// Sort values are any Comparable and sorting is done according to natural order.
	CUSTOM

	// STRING_VAL Sort using term values as Strings,
	// but comparing by value (using String.compareTo) for all comparisons.
	// This is typically slower than STRING, which uses ordinals to do the sorting.
	STRING_VAL

	// REWRITEABLE Force rewriting of SortField using rewrite(IndexSearcher) before it can be used for sorting
	REWRITEABLE
)

var _ SortFieldProvider = &sortFieldProvider{}

var once sync.Once

func init() {
	once.Do(func() {
		SingleSortFieldProvider.Register(ProviderName, newSortFieldProvider())
	})
}

type sortFieldProvider struct {
	name string
}

func newSortFieldProvider() *sortFieldProvider {
	return &sortFieldProvider{name: ProviderName}
}

func (s *sortFieldProvider) GetName() string {
	return s.name
}

func (s *sortFieldProvider) ReadSortField(in store.DataInput) (*SortField, error) {
	field, err := in.ReadString()
	if err != nil {
		return nil, err
	}

	fieldType, err := readType(in)
	if err != nil {
		return nil, err
	}

	num, err := in.ReadUint32()
	if err != nil {
		return nil, err
	}

	reverse := num == 1

	sf, err := NewSortField(field, fieldType, reverse)
	if err != nil {
		return nil, err
	}

	num1, err := in.ReadUint32()
	if err != nil {
		return nil, err
	}

	if num1 == 1 {
		switch sf._type {
		case STRING:
			num, err := in.ReadUint32()
			if err != nil {
				return nil, err
			}
			if num == 1 {
				sf.SetMissingValue(STRING_FIRST)
			} else {
				sf.SetMissingValue(STRING_LAST)
			}
		case INT:
			num, err := in.ReadUint32()
			if err != nil {
				return nil, err
			}
			sf.SetMissingValue(num)
		case LONG:
			num, err := in.ReadUint64()
			if err != nil {
				return nil, err
			}
			sf.SetMissingValue(num)
		case FLOAT:
			num, err := in.ReadUint32()
			if err != nil {
				return nil, err
			}
			sf.SetMissingValue(math.Float32frombits(num))
		case DOUBLE:
			num, err := in.ReadUint64()
			if err != nil {
				return nil, err
			}
			sf.SetMissingValue(math.Float64frombits(num))
		default:
			return nil, fmt.Errorf("cannot deserialize sort of type %s", sf._type)
		}
	}
	return sf, nil
}

func (s *sortFieldProvider) WriteSortField(sf *SortField, out store.DataOutput) error {
	return sf.serialize(out)
}

func readType(in store.DataInput) (SortFieldType, error) {
	value, err := in.ReadString()
	if err != nil {
		return 0, err
	}
	switch value {
	case "SCORE":
		return SCORE, nil
	case "DOC":
		return DOC, nil
	case "STRING":
		return STRING, nil
	case "INT":
		return INT, nil
	case "FLOAT":
		return FLOAT, nil
	case "LONG":
		return LONG, nil
	case "DOUBLE":
		return DOUBLE, nil
	case "CUSTOM":
		return CUSTOM, nil
	case "STRING_VAL":
		return STRING_VAL, nil
	case "REWRITEABLE":
		return REWRITEABLE, nil
	default:
		return SCORE, errors.New("undefined sort filed type")
	}
}
