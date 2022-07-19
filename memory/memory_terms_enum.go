package memory

import (
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/tokenattributes"
)

var _ index.TermsEnum = &MemoryTermsEnum{}

type MemoryTermsEnum struct {
}

func (m *MemoryTermsEnum) Next() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryTermsEnum) Attributes() *tokenattributes.AttributeSource {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryTermsEnum) SeekExact(text []byte) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryTermsEnum) SeekCeil(text []byte) (index.SeekStatus, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryTermsEnum) SeekExactByOrd(ord int64) error {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryTermsEnum) SeekExactExpert(term []byte, state index.TermState) error {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryTermsEnum) Term() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryTermsEnum) Ord() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryTermsEnum) DocFreq() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryTermsEnum) TotalTermFreq() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryTermsEnum) Postings(reuse index.PostingsEnum, flags int) (index.PostingsEnum, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryTermsEnum) Impacts(flags int) (index.ImpactsEnum, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryTermsEnum) TermState() (index.TermState, error) {
	//TODO implement me
	panic("implement me")
}
