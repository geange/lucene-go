package simpletext

import (
	"errors"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/types"
)

var _ index.BaseTermsEnum = &SimpleTextTermsEnum{}

type SimpleTextTermsEnum struct {
	*index.BaseTermsEnumImp

	indexOptions  types.IndexOptions
	docFreq       int
	totalTermFreq int64
	docsStart     int64
	skipPointer   int64
	ended         bool

	fstEnum interface{}
}

func (s *SimpleTextTermsEnum) Next() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextTermsEnum) SeekCeil(text []byte) (index.SeekStatus, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextTermsEnum) SeekExactByOrd(ord int64) error {
	return errors.New("UnsupportedOperationException")
}

func (s *SimpleTextTermsEnum) Term() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextTermsEnum) Ord() (int64, error) {
	return 0, errors.New("UnsupportedOperationException")
}

func (s *SimpleTextTermsEnum) DocFreq() (int, error) {
	return s.docFreq, nil
}

func (s *SimpleTextTermsEnum) TotalTermFreq() (int64, error) {
	if s.indexOptions == types.INDEX_OPTIONS_DOCS {
		return int64(s.docFreq), nil
	}
	return s.totalTermFreq, nil
}

func (s *SimpleTextTermsEnum) Postings(reuse index.PostingsEnum, flags int) (index.PostingsEnum, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextTermsEnum) Impacts(flags int) (index.ImpactsEnum, error) {
	//TODO implement me
	panic("implement me")
}

type BytesRefFSTEnum interface {
}
