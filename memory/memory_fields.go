package memory

import (
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/util"
	"github.com/geange/lucene-go/core/util/automaton"
)

type MemoryFields struct {
	fields *treemap.Map
}

func NewMemoryFields(fields *treemap.Map) *MemoryFields {
	return &MemoryFields{fields: fields}
}

func (m *MemoryFields) Iterator() func() string {
	m.fields.Keys()
	//keys := make([]string, 0, len(m.fields))
	//for k := range m.fields {
	//	keys = append(keys, k)
	//}
	//i := 0
	//return func() string {
	//	if i < len(keys) {
	//		res := keys[i]
	//		i++
	//		return res
	//	}
	//	return ""
	//}
	panic("")
}

func (m *MemoryFields) Terms(field string) (index.Terms, error) {
	//info, ok := m.fields[field]
	//if !ok {
	//	return nil, nil
	//}
	panic("")
}

type terms struct {
	info *Info
}

func (t *terms) Iterator() (index.TermsEnum, error) {
	//TODO implement me
	panic("implement me")
}

func (t *terms) Intersect(compiled *automaton.CompiledAutomaton, startTerm []byte) (index.TermsEnum, error) {
	//TODO implement me
	panic("implement me")
}

func (t *terms) Size() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (t *terms) GetSumTotalTermFreq() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (t *terms) GetSumDocFreq() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (t *terms) GetDocCount() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (t *terms) HasFreqs() bool {
	//TODO implement me
	panic("implement me")
}

func (t *terms) HasOffsets() bool {
	//TODO implement me
	panic("implement me")
}

func (t *terms) HasPositions() bool {
	//TODO implement me
	panic("implement me")
}

func (t *terms) HasPayloads() bool {
	//TODO implement me
	panic("implement me")
}

func (t *terms) GetMin() (*util.BytesRef, error) {
	//TODO implement me
	panic("implement me")
}

func (t *terms) GetMax() (*util.BytesRef, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryFields) Size() int {
	return m.fields.Size()
}
