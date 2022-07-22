package memory

import (
	"github.com/geange/lucene-go/core/index"
)

type Terms struct {
	*index.TermsImp

	info          *Info
	storeOffsets  bool
	storePayloads bool
}

func NewTerms(info *Info) *Terms {
	terms := &Terms{info: info}
	terms.TermsImp = index.NewTermsImp(terms)
	return terms
}

func (t *Terms) Iterator() (index.TermsEnum, error) {
	return NewMemoryTermsEnum(t.info), nil
}

func (t *Terms) Size() (int, error) {
	return t.info.terms.Size(), nil
}

func (t *Terms) GetSumTotalTermFreq() (int64, error) {
	return t.info.sumTotalTermFreq, nil
}

func (t *Terms) GetSumDocFreq() (int64, error) {
	return int64(t.info.terms.Size()), nil
}

func (t *Terms) GetDocCount() (int, error) {
	size, err := t.Size()
	if err != nil {
		return 0, err
	}
	if size > 0 {
		return 1, nil
	}
	return 0, nil
}

func (t *Terms) HasFreqs() bool {
	return true
}

func (t *Terms) HasOffsets() bool {
	return t.storeOffsets
}

func (t *Terms) HasPositions() bool {
	return true
}

func (t *Terms) HasPayloads() bool {
	return t.storePayloads
}
