package index

import (
	"fmt"
	"github.com/geange/lucene-go/core/store"
)

var _ SortField = &SortedSetSortField{}

// SortedSetSortField
// SortField for SortedSetDocValues.
// A SortedSetDocValues contains multiple values for a field, so sorting with this technique "selects" a item
// as the representative sort item for the document.
// By default, the minimum item in the set is selected as the sort item, but this can be customized. Selectors
// other than the default do have some limitations to ensure that all selections happen in constant-time for performance.
// Like sorting by string, this also supports sorting missing values as first or last, via setMissingValue(Object).
// See Also: SortedSetSelector
type SortedSetSortField struct {
	*SortFieldDefault

	selector SortedSetSelectorType
}

func NewSortedSetSortField(field string, reverse bool) *SortedSetSortField {
	return NewSortedSetSortFieldV1(field, reverse, MIN)
}

func (s *SortedSetSortField) serialize(out store.DataOutput) error {
	err := out.WriteString(s.GetField())
	if err != nil {
		return err
	}
	reverse := 0
	if s.reverse {
		reverse = 1
	}
	err = out.WriteUint32(uint32(reverse))
	if err != nil {
		return err
	}
	err = out.WriteUint32(uint32(s.selector))
	if err != nil {
		return err
	}
	if s.missingValue == STRING_FIRST {
		return out.WriteUint32(1)
	}

	if s.missingValue == STRING_LAST {
		return out.WriteUint32(2)
	}
	return out.WriteUint32(0)
}

func NewSortedSetSortFieldV1(field string, reverse bool,
	selector SortedSetSelectorType) *SortedSetSortField {

	return &SortedSetSortField{
		SortFieldDefault: NewSortFieldV1(field, CUSTOM, reverse),
		selector:         selector,
	}
}

var _ SortFieldProvider = &SortedSetSortFieldProvider{}

type SortedSetSortFieldProvider struct {
}

func NewSortedSetSortFieldProvider() *SortedSetSortFieldProvider {
	return &SortedSetSortFieldProvider{}
}

func (s *SortedSetSortFieldProvider) GetName() string {
	return "SortedSetSortField"
}

func (s *SortedSetSortFieldProvider) ReadSortField(in store.DataInput) (SortField, error) {
	field, err := in.ReadString()
	if err != nil {
		return nil, err
	}

	num, err := in.ReadUint32()
	if err != nil {
		return nil, err
	}

	_type, err := readSelectorType(in)
	if err != nil {
		return nil, err
	}

	sf := NewSortedSetSortFieldV1(field, num == 1, _type)
	missingValue, err := in.ReadUint32()
	if missingValue == 1 {
		err := sf.SetMissingValue(STRING_FIRST)
		if err != nil {
			return nil, err
		}
	} else if missingValue == 2 {
		err := sf.SetMissingValue(STRING_LAST)
		if err != nil {
			return nil, err
		}
	}
	return sf, nil
}

func readSelectorType(in store.DataInput) (SortedSetSelectorType, error) {
	_type, err := in.ReadUint32()
	if err != nil {
		return 0, err
	}

	if _type >= 4 {
		return 0, fmt.Errorf("cannot deserialize SortedSetSortField: unknown selector type %d", _type)
	}

	return SortedSetSelectorType(int(_type)), nil
}

func (s *SortedSetSortFieldProvider) WriteSortField(sf SortField, out store.DataOutput) error {
	v, ok := sf.(*SortedSetSortField)
	if !ok {
		return fmt.Errorf("sf is not *SortedSetSortField")
	}
	return v.serialize(out)
}
