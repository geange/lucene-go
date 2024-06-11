package index

import (
	"context"
	"fmt"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
)

var _ index.SortField = &SortedSetSortField{}

// SortedSetSortField
// SortField for SortedSetDocValues.
// A SortedSetDocValues contains multiple values for a field, so sorting with this technique "selects" a item
// as the representative sort item for the document.
// By default, the minimum item in the set is selected as the sort item, but this can be customized. Selectors
// other than the default do have some limitations to ensure that all selections happen in constant-time for performance.
// Like sorting by string, this also supports sorting missing values as first or last, via setMissingValue(Object).
// See Also: SortedSetSelector
type SortedSetSortField struct {
	*BaseSortField

	selector SortedSetSelectorType
}

func NewSortedSetSortField(field string, reverse bool) *SortedSetSortField {
	return NewSortedSetSortFieldV1(field, reverse, MIN)
}

func (s *SortedSetSortField) serialize(ctx context.Context, out store.DataOutput) error {
	if err := out.WriteString(ctx, s.GetField()); err != nil {
		return err
	}
	reverse := 0
	if s.reverse {
		reverse = 1
	}
	if err := out.WriteUint32(ctx, uint32(reverse)); err != nil {
		return err
	}
	if err := out.WriteUint32(ctx, uint32(s.selector)); err != nil {
		return err
	}
	if s.missingValue == STRING_FIRST {
		return out.WriteUint32(nil, 1)
	}

	if s.missingValue == STRING_LAST {
		return out.WriteUint32(nil, 2)
	}
	return out.WriteUint32(ctx, 0)
}

func NewSortedSetSortFieldV1(field string, reverse bool,
	selector SortedSetSelectorType) *SortedSetSortField {

	return &SortedSetSortField{
		BaseSortField: NewSortFieldV1(field, index.CUSTOM, reverse),
		selector:      selector,
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

func (s *SortedSetSortFieldProvider) ReadSortField(ctx context.Context, in store.DataInput) (index.SortField, error) {
	field, err := in.ReadString(ctx)
	if err != nil {
		return nil, err
	}

	num, err := in.ReadUint32(ctx)
	if err != nil {
		return nil, err
	}

	sType, err := readSelectorType(ctx, in)
	if err != nil {
		return nil, err
	}

	sf := NewSortedSetSortFieldV1(field, num == 1, sType)
	missingValue, err := in.ReadUint32(ctx)
	if err != nil {
		return nil, err
	}

	switch missingValue {
	case 1:
		if err := sf.SetMissingValue(STRING_FIRST); err != nil {
			return nil, err
		}
	case 2:
		if err := sf.SetMissingValue(STRING_LAST); err != nil {
			return nil, err
		}
	}

	return sf, nil
}

func readSelectorType(ctx context.Context, in store.DataInput) (SortedSetSelectorType, error) {
	_type, err := in.ReadUint32(ctx)
	if err != nil {
		return 0, err
	}

	if _type >= 4 {
		return 0, fmt.Errorf("cannot deserialize SortedSetSortField: unknown selector type %d", _type)
	}

	return SortedSetSelectorType(int(_type)), nil
}

func (s *SortedSetSortFieldProvider) WriteSortField(ctx context.Context, sf index.SortField, out store.DataOutput) error {
	v, ok := sf.(*SortedSetSortField)
	if !ok {
		return fmt.Errorf("sf is not *SortedSetSortField")
	}
	return v.serialize(ctx, out)
}
