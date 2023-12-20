package index

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/geange/lucene-go/core/store"
	"math"
	"reflect"
)

const (
	STRING_FIRST = "SortField.STRING_FIRST"
	STRING_LAST  = "SortField.STRING_LAST"
)

// SortField
// Stores information about how to sort documents by terms in an individual field.
// Fields must be indexed in order to sort by them.
// Created: Feb 11, 2004 1:25:29 PM
// Since: lucene 1.4
// See Also: Sort
type SortField interface {
	// GetMissingValue Return the item to use for documents that don't have a item.
	// A item of null indicates that default should be used.
	GetMissingValue() any

	// SetMissingValue Set the item to use for documents that don't have a item.
	SetMissingValue(missingValue any) error

	// GetField Returns the name of the field. Could return null if the sort is by SCORE or DOC.
	// Returns: Name of field, possibly null.
	GetField() string

	// GetType Returns the type of contents in the field.
	// Returns: One of the constants SCORE, DOC, STRING, INT or FLOAT.
	GetType() SortFieldType

	// GetReverse Returns whether the sort should be reversed.
	// Returns: True if natural order should be reversed.
	GetReverse() bool

	GetComparatorSource() FieldComparatorSource

	// SetCanUsePoints For numeric sort fields, setting this field, indicates that the same numeric data
	// has been indexed with two fields: doc values and points and that these fields have the same name.
	// This allows to use sort optimization and skip non-competitive documents.
	SetCanUsePoints()

	GetCanUsePoints() bool

	SetBytesComparator(fn BytesComparator)

	GetBytesComparator() BytesComparator

	// GetComparator Returns the FieldComparator to use for sorting.
	//Params: 	numHits – number of top hits the queue will store
	//			sortPos – position of this SortField within Sort. The comparator is primary if
	//			sortPos==0, secondary if sortPos==1, etc. Some comparators can optimize
	//			themselves when they are the primary sort.
	//Returns: FieldComparator to use when sorting
	//lucene.experimental
	GetComparator(numHits, sortPos int) FieldComparator

	//rewrite(searcher search.IndexSearcher)

	GetIndexSorter() IndexSorter

	Serialize(ctx context.Context, out store.DataOutput) error
	Equals(other SortField) bool
	String() string
}

type BytesComparator func(a, b []byte) int

var _ SortField = &SortFieldDefault{}

var (
	FIELD_SCORE = NewSortField("", SCORE)
	FIELD_DOC   = NewSortField("", DOC)
)

// SortFieldDefault Stores information about how to sort documents by terms in an individual field.
// Fields must be indexed in order to sort by them.
// Created: Feb 11, 2004 1:25:29 PM
// Since: lucene 1.4
// See Also: Sort
type SortFieldDefault struct {
	field            string
	_type            SortFieldType
	reverse          bool
	comparatorSource FieldComparatorSource
	canUsePoints     bool
	bytesComparator  BytesComparator
	missingValue     any
}

func NewSortField(field string, _type SortFieldType) *SortFieldDefault {
	s := &SortFieldDefault{}
	s.initFieldType(&field, _type)
	return s
}

func NewSortFieldV1(field string, _type SortFieldType, reverse bool) *SortFieldDefault {
	s := NewSortField(field, _type)
	s.reverse = reverse
	return s
}

func (s *SortFieldDefault) GetComparatorSource() FieldComparatorSource {
	return s.comparatorSource
}

func (s *SortFieldDefault) GetReverse() bool {
	return s.reverse
}

func (s *SortFieldDefault) SetBytesComparator(fn BytesComparator) {
	s.bytesComparator = fn
}

func (s *SortFieldDefault) GetBytesComparator() BytesComparator {
	return s.bytesComparator
}

func (s *SortFieldDefault) GetComparator(numHits, sortPos int) FieldComparator {
	var fieldComparator FieldComparator
	switch s._type {
	// TODO: fix it
	}
	if !s.GetCanUsePoints() {
		fieldComparator.DisableSkipping()
	}
	return fieldComparator
}

func newSortField(field string, _type SortFieldType, reverse bool) (*SortFieldDefault, error) {
	sortField := &SortFieldDefault{reverse: reverse}
	if err := sortField.initFieldType(&field, _type); err != nil {
		return nil, err
	}
	return sortField, nil
}

// Sets field & type, and ensures field is not NULL unless
// type is SCORE or DOC
func (s *SortFieldDefault) initFieldType(field *string, _type SortFieldType) error {
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

func (s *SortFieldDefault) String() string {
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
func (s *SortFieldDefault) GetField() string {
	return s.field
}

// GetType Returns the type of contents in the field.
// Returns: One of the constants SCORE, DOC, STRING, INT or FLOAT.
func (s *SortFieldDefault) GetType() SortFieldType {
	return s._type
}

// SetCanUsePoints For numeric sort fields, setting this field, indicates that the same numeric
// data has been indexed with two fields: doc values and points and that these fields have the
// same name. This allows to use sort optimization and skip non-competitive documents.
func (s *SortFieldDefault) SetCanUsePoints() {
	s.canUsePoints = true
}

func (s *SortFieldDefault) GetCanUsePoints() bool {
	return s.canUsePoints
}

// NeedsScores Whether the relevance score is needed to sort documents.
func (s *SortFieldDefault) NeedsScores() bool {
	return s._type == SCORE
}

// GetMissingValue Return the item to use for documents that don't have a item. A item of null indicates that default should be used.
func (s *SortFieldDefault) GetMissingValue() any {
	return s.missingValue
}

// SetMissingValue Set the item to use for documents that don't have a item.
func (s *SortFieldDefault) SetMissingValue(missingValue any) error {
	s.missingValue = missingValue
	return nil
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
func (s *SortFieldDefault) GetIndexSorter() IndexSorter {
	sorter := &EmptyNumericDocValuesProvider{
		FnGet: func(reader LeafReader) (NumericDocValues, error) {
			return GetNumeric(reader, s.field)
		},
	}

	switch s._type {
	case STRING:
		sorter := &EmptySortedDocValuesProvider{
			FnGet: func(reader LeafReader) (SortedDocValues, error) {
				return GetSorted(reader, s.field)
			},
		}
		return NewStringSorter(ProviderName, s.missingValue.(string), s.reverse, sorter)
	case INT:
		return NewIntSorter(ProviderName, s.missingValue.(int32), s.reverse, sorter)
	case LONG:
		return NewLongSorter(ProviderName, s.missingValue.(int64), s.reverse, sorter)
	case DOUBLE:
		return NewDoubleSorter(ProviderName, s.missingValue.(float64), s.reverse, sorter)
	case FLOAT:
		return NewFloatSorter(ProviderName, s.missingValue.(float32), s.reverse, sorter)
	default:
		return nil
	}
}

func (s *SortFieldDefault) Serialize(ctx context.Context, out store.DataOutput) error {
	if err := out.WriteString(ctx, s.field); err != nil {
		return err
	}
	if err := out.WriteString(ctx, s._type.String()); err != nil {
		return err
	}

	if err := out.WriteUint32(ctx, func() uint32 {
		if s.reverse {
			return 1
		}
		return 0
	}()); err != nil {
		return err
	}

	if s.missingValue == nil {
		return out.WriteUint32(ctx, 0)
	}

	if err := out.WriteUint32(ctx, 1); err != nil {
		return err
	}
	switch s._type {
	case SCORE:
		switch s.missingValue.(string) {
		case STRING_FIRST:
			if err := out.WriteUint32(ctx, 1); err != nil {
				return err
			}
		case STRING_LAST:
			if err := out.WriteUint32(ctx, 0); err != nil {
				return err
			}
		default:
			return fmt.Errorf("cannot serialize missing item of %v for type STRING", s.missingValue)
		}

		return nil
	case INT:
		return out.WriteUint32(ctx, uint32(s.missingValue.(int32)))
	case LONG:
		return out.WriteUint64(ctx, uint64(s.missingValue.(int64)))
	case FLOAT:
		return out.WriteUint32(ctx, math.Float32bits(s.missingValue.(float32)))
	case DOUBLE:
		return out.WriteUint64(ctx, math.Float64bits(s.missingValue.(float64)))
	default:
		return fmt.Errorf("cannot serialize SortField of type %s", s._type)
	}
}

func (s *SortFieldDefault) Equals(other SortField) bool {
	return s.field == other.GetField() &&
		s._type == other.GetType() &&
		s.reverse == other.GetReverse() &&
		s.canUsePoints == other.GetCanUsePoints() &&
		reflect.DeepEqual(s.missingValue, other.GetMissingValue())

	//return s.field == other.field &&
	//	s._type == other._type &&
	//	s.reverse == other.reverse &&
	//	s.canUsePoints == other.canUsePoints &&
	//	s.missingValue == other.missingValue
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

	// CUSTOM Sort using a custom cmp.
	// Sort values are any Comparable and sorting is done according to natural order.
	CUSTOM

	// STRING_VAL Sort using term values as Strings,
	// but comparing by item (using String.compareTo) for all comparisons.
	// This is typically slower than STRING, which uses ordinals to do the sorting.
	STRING_VAL

	// REWRITEABLE Force rewriting of SortField using rewrite(IndexSearcher) before it can be used for sorting
	REWRITEABLE
)

func init() {
	RegisterSortFieldProvider(newSortFieldProvider())
	RegisterSortFieldProvider(NewSortedSetSortFieldProvider())
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

func (s *sortFieldProvider) ReadSortField(ctx context.Context, in store.DataInput) (SortField, error) {
	field, err := in.ReadString(ctx)
	if err != nil {
		return nil, err
	}

	fieldType, err := readType(ctx, in)
	if err != nil {
		return nil, err
	}

	num, err := in.ReadUint32(ctx)
	if err != nil {
		return nil, err
	}

	reverse := num == 1

	sf, err := newSortField(field, fieldType, reverse)
	if err != nil {
		return nil, err
	}

	num1, err := in.ReadUint32(ctx)
	if err != nil {
		return nil, err
	}

	if num1 == 1 {
		switch sf._type {
		case STRING:
			num, err := in.ReadUint32(ctx)
			if err != nil {
				return nil, err
			}
			if num == 1 {
				if err := sf.SetMissingValue(STRING_FIRST); err != nil {
					return nil, err
				}
			} else {
				if err := sf.SetMissingValue(STRING_LAST); err != nil {
					return nil, err
				}
			}
		case INT:
			num, err := in.ReadUint32(ctx)
			if err != nil {
				return nil, err
			}
			if err := sf.SetMissingValue(num); err != nil {
				return nil, err
			}
		case LONG:
			num, err := in.ReadUint64(ctx)
			if err != nil {
				return nil, err
			}
			if err := sf.SetMissingValue(num); err != nil {
				return nil, err
			}
		case FLOAT:
			num, err := in.ReadUint32(ctx)
			if err != nil {
				return nil, err
			}
			if err := sf.SetMissingValue(math.Float32frombits(num)); err != nil {
				return nil, err
			}
		case DOUBLE:
			num, err := in.ReadUint64(ctx)
			if err != nil {
				return nil, err
			}
			if err := sf.SetMissingValue(math.Float64frombits(num)); err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("cannot deserialize sort of type %s", sf._type)
		}
	}
	return sf, nil
}

func (s *sortFieldProvider) WriteSortField(ctx context.Context, sf SortField, out store.DataOutput) error {
	return sf.Serialize(ctx, out)
}

func readType(ctx context.Context, in store.DataInput) (SortFieldType, error) {
	value, err := in.ReadString(ctx)
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
