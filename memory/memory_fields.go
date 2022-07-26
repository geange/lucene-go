package memory

import (
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/geange/lucene-go/core/index"
)

type MemoryFields struct {
	fields *treemap.Map

	*MemoryIndex
}

func (m *MemoryIndex) NewMemoryFields(fields *treemap.Map) *MemoryFields {
	return &MemoryFields{
		fields:      fields,
		MemoryIndex: m,
	}
}

func (m *MemoryFields) Iterator() func() string {
	m.fields.Keys()
	keys := make([]string, 0)

	m.fields.Each(func(key interface{}, value interface{}) {
		if value.(*Info).numTokens > 0 {
			keys = append(keys, value.(string))
		}
	})

	for _, v := range m.fields.Keys() {
		keys = append(keys, v.(string))
	}

	i := 0

	return func() string {
		if i < len(keys) {
			res := keys[i]
			i++
			return res
		}
		return ""
	}
}

func (m *MemoryFields) Terms(field string) (index.Terms, error) {
	v, ok := m.fields.Get(field)
	if !ok {
		return nil, nil
	}
	info := v.(*Info)
	if info.numTokens <= 0 {
		return nil, nil
	}

	return m.NewTerms(info), nil
}

func (m *MemoryFields) Size() int {
	return m.fields.Size()
}
