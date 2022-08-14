package simpletext

import (
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util"
	"github.com/geange/lucene-go/core/util/automaton"
)

var _ index.Terms = &SimpleTextTerms{}

type SimpleTextTerms struct {
	termsStart       int64
	fieldInfo        *types.FieldInfo
	maxDoc           int
	sumTotalTermFreq int64
	sumDocFreq       int64
	docCount         int
	fst              interface{}
	termCount        int
	scratch          *util.BytesRefBuilder
	scratchUTF16     *util.CharsRefBuilder
}

func (s *SimpleTextFieldsReader) NewSimpleTextTerms(field string, termsStart int64, maxDoc int) *SimpleTextTerms {
	panic("")
}

func (s *SimpleTextTerms) loadTerms() error {
	panic("")
}

func (s *SimpleTextTerms) Iterator() (index.TermsEnum, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextTerms) Intersect(compiled *automaton.CompiledAutomaton, startTerm []byte) (index.TermsEnum, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextTerms) Size() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextTerms) GetSumTotalTermFreq() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextTerms) GetSumDocFreq() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextTerms) GetDocCount() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextTerms) HasFreqs() bool {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextTerms) HasOffsets() bool {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextTerms) HasPositions() bool {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextTerms) HasPayloads() bool {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextTerms) GetMin() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextTerms) GetMax() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}
