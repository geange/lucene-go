package index

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"math"
	"reflect"
)

const (
	STRING_FIRST = "SortField.STRING_FIRST"
	STRING_LAST  = "SortField.STRING_LAST"
)

var _ index.SortField = &BaseSortField{}

var (
	FIELD_SCORE = NewSortField("", index.SCORE)
	FIELD_DOC   = NewSortField("", index.DOC)
)

// BaseSortField
// Stores information about how to sort documents by terms in an individual field.
// Fields must be indexed in order to sort by them.
// Created: Feb 11, 2004 1:25:29 PM
// Since: lucene 1.4
// See Also: Sort
type BaseSortField struct {
	field            string
	_type            index.SortFieldType
	reverse          bool
	comparatorSource index.FieldComparatorSource
	canUsePoints     bool
	bytesComparator  index.BytesComparator
	missingValue     any
}

func NewSortField(field string, _type index.SortFieldType) *BaseSortField {
	s := &BaseSortField{}
	s.initFieldType(&field, _type)
	return s
}

func NewSortFieldV1(field string, _type index.SortFieldType, reverse bool) *BaseSortField {
	s := NewSortField(field, _type)
	s.reverse = reverse
	return s
}

func (s *BaseSortField) GetComparatorSource() index.FieldComparatorSource {
	return s.comparatorSource
}

func (s *BaseSortField) GetReverse() bool {
	return s.reverse
}

func (s *BaseSortField) SetBytesComparator(fn index.BytesComparator) {
	s.bytesComparator = fn
}

func (s *BaseSortField) GetBytesComparator() index.BytesComparator {
	return s.bytesComparator
}

func (s *BaseSortField) GetComparator(numHits, sortPos int) index.FieldComparator {
	var fieldComparator index.FieldComparator
	switch s._type {
	// TODO: fix it
	}
	if !s.GetCanUsePoints() {
		fieldComparator.DisableSkipping()
	}
	return fieldComparator
}

func newSortField(field string, _type index.SortFieldType, reverse bool) (*BaseSortField, error) {
	sortField := &BaseSortField{reverse: reverse}
	if err := sortField.initFieldType(&field, _type); err != nil {
		return nil, err
	}
	return sortField, nil
}

// Sets field & type, and ensures field is not NULL unless
// type is SCORE or DOC
func (s *BaseSortField) initFieldType(field *string, _type index.SortFieldType) error {
	s._type = _type
	if field == nil {
		switch _type {
		case index.SCORE, index.DOC:
		default:
			return errors.New("field can only be null when type is SCORE or DOC")
		}
	}
	s.field = *field
	return nil
}

func (s *BaseSortField) String() string {
	buffer := new(bytes.Buffer)

	switch s._type {
	case index.SCORE:
		buffer.WriteString("<score>")
	case index.DOC:
		buffer.WriteString("<doc>")
	case index.STRING:
		buffer.WriteString(fmt.Sprintf(`<string: "%s">`, s.field))
	case index.STRING_VAL:
		buffer.WriteString(fmt.Sprintf(`<string_val: "%s">`, s.field))
	case index.INT:
		buffer.WriteString(fmt.Sprintf(`<int: "%s">`, s.field))
	case index.LONG:
		buffer.WriteString(fmt.Sprintf(`<long: "%s">`, s.field))
	case index.FLOAT:
		buffer.WriteString(fmt.Sprintf(`<float: "%s">`, s.field))
	case index.DOUBLE:
		buffer.WriteString(fmt.Sprintf(`<double: "%s">`, s.field))
	case index.CUSTOM:
		buffer.WriteString(fmt.Sprintf(`<custom: "%s">`, s.field))
	case index.REWRITEABLE:
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
func (s *BaseSortField) GetField() string {
	return s.field
}

// GetType Returns the type of contents in the field.
// Returns: One of the constants SCORE, DOC, STRING, INT or FLOAT.
func (s *BaseSortField) GetType() index.SortFieldType {
	return s._type
}

// SetCanUsePoints For numeric sort fields, setting this field, indicates that the same numeric
// data has been indexed with two fields: doc values and points and that these fields have the
// same name. This allows to use sort optimization and skip non-competitive documents.
func (s *BaseSortField) SetCanUsePoints() {
	s.canUsePoints = true
}

func (s *BaseSortField) GetCanUsePoints() bool {
	return s.canUsePoints
}

// NeedsScores Whether the relevance score is needed to sort documents.
func (s *BaseSortField) NeedsScores() bool {
	return s._type == index.SCORE
}

// GetMissingValue Return the item to use for documents that don't have a item. A item of null indicates that default should be used.
func (s *BaseSortField) GetMissingValue() any {
	return s.missingValue
}

// SetMissingValue Set the item to use for documents that don't have a item.
func (s *BaseSortField) SetMissingValue(missingValue any) error {
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
func (s *BaseSortField) GetIndexSorter() index.IndexSorter {
	sorter := &EmptyNumericDocValuesProvider{
		FnGet: func(reader index.LeafReader) (index.NumericDocValues, error) {
			return GetNumeric(reader, s.field)
		},
	}

	switch s._type {
	case index.STRING:
		sorter := &EmptySortedDocValuesProvider{
			FnGet: func(reader index.LeafReader) (index.SortedDocValues, error) {
				return GetSorted(reader, s.field)
			},
		}
		return NewStringSorter(ProviderName, s.missingValue.(string), s.reverse, sorter)
	case index.INT:
		return NewIntSorter(ProviderName, s.missingValue.(int32), s.reverse, sorter)
	case index.LONG:
		return NewLongSorter(ProviderName, s.missingValue.(int64), s.reverse, sorter)
	case index.DOUBLE:
		return NewDoubleSorter(ProviderName, s.missingValue.(float64), s.reverse, sorter)
	case index.FLOAT:
		return NewFloatSorter(ProviderName, s.missingValue.(float32), s.reverse, sorter)
	default:
		return nil
	}
}

func (s *BaseSortField) Serialize(ctx context.Context, out store.DataOutput) error {
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
	case index.SCORE:
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
	case index.INT:
		return out.WriteUint32(ctx, uint32(s.missingValue.(int32)))
	case index.LONG:
		return out.WriteUint64(ctx, uint64(s.missingValue.(int64)))
	case index.FLOAT:
		return out.WriteUint32(ctx, math.Float32bits(s.missingValue.(float32)))
	case index.DOUBLE:
		return out.WriteUint64(ctx, math.Float64bits(s.missingValue.(float64)))
	default:
		return fmt.Errorf("cannot serialize SortField of type %s", s._type)
	}
}

func (s *BaseSortField) Equals(other index.SortField) bool {
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

func (s *sortFieldProvider) ReadSortField(ctx context.Context, in store.DataInput) (index.SortField, error) {
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
		case index.STRING:
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
		case index.INT:
			num, err := in.ReadUint32(ctx)
			if err != nil {
				return nil, err
			}
			if err := sf.SetMissingValue(num); err != nil {
				return nil, err
			}
		case index.LONG:
			num, err := in.ReadUint64(ctx)
			if err != nil {
				return nil, err
			}
			if err := sf.SetMissingValue(num); err != nil {
				return nil, err
			}
		case index.FLOAT:
			num, err := in.ReadUint32(ctx)
			if err != nil {
				return nil, err
			}
			if err := sf.SetMissingValue(math.Float32frombits(num)); err != nil {
				return nil, err
			}
		case index.DOUBLE:
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

func (s *sortFieldProvider) WriteSortField(ctx context.Context, sf index.SortField, out store.DataOutput) error {
	return sf.Serialize(ctx, out)
}

func readType(ctx context.Context, in store.DataInput) (index.SortFieldType, error) {
	value, err := in.ReadString(ctx)
	if err != nil {
		return 0, err
	}
	switch value {
	case "SCORE":
		return index.SCORE, nil
	case "DOC":
		return index.DOC, nil
	case "STRING":
		return index.STRING, nil
	case "INT":
		return index.INT, nil
	case "FLOAT":
		return index.FLOAT, nil
	case "LONG":
		return index.LONG, nil
	case "DOUBLE":
		return index.DOUBLE, nil
	case "CUSTOM":
		return index.CUSTOM, nil
	case "STRING_VAL":
		return index.STRING_VAL, nil
	case "REWRITEABLE":
		return index.REWRITEABLE, nil
	default:
		return index.SCORE, errors.New("undefined sort filed type")
	}
}
