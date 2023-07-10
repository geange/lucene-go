package memory

import (
	"github.com/geange/gods-generic/maps/treemap"
	"github.com/geange/lucene-go/core/index"
)

type MemoryFields struct {
	fields *treemap.Map[string, *Info]

	*MemoryIndex
}

func (m *MemoryIndex) NewMemoryFields(fields *treemap.Map[string, *Info]) *MemoryFields {
	return &MemoryFields{
		fields:      fields,
		MemoryIndex: m,
	}
}

func (m *MemoryFields) Iterator() func() string {
	iterator := m.fields.Iterator()

	return func() string {
		if iterator.Next() {
			return iterator.Key()
		}
		return ""
	}
}

func (m *MemoryFields) Names() []string {
	return m.fields.Keys()
}

func (m *MemoryFields) Terms(field string) (index.Terms, error) {
	info, ok := m.fields.Get(field)
	if !ok {
		return nil, nil
	}

	if info.numTokens <= 0 {
		return nil, nil
	}

	return m.NewTerms(info), nil
}

func (m *MemoryFields) Size() int {
	return m.fields.Size()
}
