package index

import (
	"github.com/geange/lucene-go/core/tokenattributes"
)

// SortedDocValuesTermsEnum Creates a new TermsEnum over the provided values
type SortedDocValuesTermsEnum struct {
}

func NewSortedDocValuesTermsEnum(values SortedDocValues) *SortedDocValuesTermsEnum {
	return &SortedDocValuesTermsEnum{}
}

func (s *SortedDocValuesTermsEnum) Next() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SortedDocValuesTermsEnum) Attributes() *tokenattributes.AttributeSource {
	//TODO implement me
	panic("implement me")
}

func (s *SortedDocValuesTermsEnum) SeekExact(text []byte) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SortedDocValuesTermsEnum) SeekCeil(text []byte) (SeekStatus, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SortedDocValuesTermsEnum) SeekExactByOrd(ord int64) error {
	//TODO implement me
	panic("implement me")
}

func (s *SortedDocValuesTermsEnum) SeekExactExpert(term []byte, state TermState) error {
	//TODO implement me
	panic("implement me")
}

func (s *SortedDocValuesTermsEnum) Term() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SortedDocValuesTermsEnum) Ord() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SortedDocValuesTermsEnum) DocFreq() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SortedDocValuesTermsEnum) TotalTermFreq() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SortedDocValuesTermsEnum) Postings(reuse PostingsEnum, flags int) (PostingsEnum, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SortedDocValuesTermsEnum) Impacts(flags int) (ImpactsEnum, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SortedDocValuesTermsEnum) TermState() (TermState, error) {
	//TODO implement me
	panic("implement me")
}
