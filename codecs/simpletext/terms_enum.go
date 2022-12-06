package simpletext

import (
	"errors"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/types"
)

var _ index.BaseTermsEnum = &TermsEnum{}

type TermsEnum struct {
	*index.BaseTermsEnumImp

	indexOptions  types.IndexOptions
	docFreq       int
	totalTermFreq int64
	docsStart     int64
	skipPointer   int64
	ended         bool

	fstEnum interface{}
}

func (s *TermsEnum) Next() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (s *TermsEnum) SeekCeil(text []byte) (index.SeekStatus, error) {
	//TODO implement me
	panic("implement me")
}

func (s *TermsEnum) SeekExactByOrd(ord int64) error {
	return errors.New("UnsupportedOperationException")
}

func (s *TermsEnum) Term() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (s *TermsEnum) Ord() (int64, error) {
	return 0, errors.New("UnsupportedOperationException")
}

func (s *TermsEnum) DocFreq() (int, error) {
	return s.docFreq, nil
}

func (s *TermsEnum) TotalTermFreq() (int64, error) {
	if s.indexOptions == types.INDEX_OPTIONS_DOCS {
		return int64(s.docFreq), nil
	}
	return s.totalTermFreq, nil
}

func (s *TermsEnum) Postings(reuse index.PostingsEnum, flags int) (index.PostingsEnum, error) {
	//TODO implement me
	panic("implement me")
}

func (s *TermsEnum) Impacts(flags int) (index.ImpactsEnum, error) {
	//TODO implement me
	panic("implement me")
}

type BytesRefFSTEnum interface {
}
