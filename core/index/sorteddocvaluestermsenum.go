package index

import (
	"context"

	"github.com/geange/lucene-go/core/util/attribute"
)

// SortedDocValuesTermsEnum Creates a new TermsEnum over the provided values
type SortedDocValuesTermsEnum struct {
}

func NewSortedDocValuesTermsEnum(values SortedDocValues) *SortedDocValuesTermsEnum {
	return &SortedDocValuesTermsEnum{}
}

func (s *SortedDocValuesTermsEnum) Next(context.Context) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SortedDocValuesTermsEnum) Attributes() *attribute.Source {
	//TODO implement me
	panic("implement me")
}

func (s *SortedDocValuesTermsEnum) SeekExact(ctx context.Context, text []byte) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SortedDocValuesTermsEnum) SeekCeil(ctx context.Context, text []byte) (SeekStatus, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SortedDocValuesTermsEnum) SeekExactByOrd(ctx context.Context, ord int64) error {
	//TODO implement me
	panic("implement me")
}

func (s *SortedDocValuesTermsEnum) SeekExactExpert(ctx context.Context, term []byte, state TermState) error {
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
