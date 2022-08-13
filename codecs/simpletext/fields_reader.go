package simpletext

import (
	"bytes"

	"github.com/emirpasic/gods/maps/hashmap"
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util"
	"github.com/geange/lucene-go/core/util/automaton"
)

var (
	_ index.FieldsProducer = &SimpleTextFieldsReader{}
	_ index.Fields         = &SimpleTextFieldsReader{}
)

type SimpleTextFieldsReader struct {
	fields     *treemap.Map
	in         store.IndexInput
	fieldInfos *index.FieldInfos
	maxDoc     int
	termsCache *hashmap.Map
}

func (s *SimpleTextFieldsReader) Names() []string {
	keys := make([]string, 0)
	s.fields.All(func(key interface{}, value interface{}) bool {
		keys = append(keys, key.(string))
		return true
	})
	return keys
}

func (s *SimpleTextFieldsReader) Terms(field string) (index.Terms, error) {
	v, ok := s.termsCache.Get(field)
	if !ok {
		fp, ok := s.fields.Get(field)
		if !ok {
			return nil, nil
		}
		terms := s.NewSimpleTextTerms(field, fp.(int64), s.maxDoc)
		s.termsCache.Put(field, terms)
		return terms, nil
	}
	return v.(*SimpleTextTerms), nil
}

func (s *SimpleTextFieldsReader) Size() int {
	return -1
}

func (s *SimpleTextFieldsReader) readFields(in store.IndexInput) (*treemap.Map, error) {
	input := store.NewBufferedChecksumIndexInput(in)
	scratch := new(bytes.Buffer)
	fields := treemap.NewWithStringComparator()

	for {
		if err := ReadLine(input, scratch); err != nil {
			return nil, err
		}

		text := scratch.Bytes()

		if bytes.Equal(text, FieldsToken.END) {
			return fields, nil
		} else if bytes.HasPrefix(text, FieldsToken.FIELD) {
			fieldName := string(text[len(FieldsToken.FIELD):])
			fields.Put(fieldName, input.GetFilePointer())
		}
	}
}

func (s *SimpleTextFieldsReader) Close() error {
	return s.in.Close()
}

func (s *SimpleTextFieldsReader) CheckIntegrity() error {
	return nil
}

func (s *SimpleTextFieldsReader) GetMergeInstance() index.FieldsProducer {
	return nil
}

type SimpleTextTermsEnum struct {
	index.BaseTermsEnum

	indexOptions  types.IndexOptions
	docFreq       int
	totalTermFreq int64
	docsStart     int64
	skipPointer   int64
	ended         bool
}

//var _ index.ImpactsEnum = &SimpleTextDocsEnum{}

type SimpleTextDocsEnum struct {
	inStart      store.IndexInput
	in           store.IndexInput
	omitTF       bool
	docID        int
	tf           int
	scratch      *util.BytesRefBuilder
	scratchUTF16 *util.CharsRefBuilder
	cost         int
}

//var _ index.ImpactsEnum = &SimpleTextPostingsEnum{}

type SimpleTextPostingsEnum struct {
	inStart        store.IndexInput
	in             store.IndexInput
	docID          int
	tf             int
	scratch        *util.BytesRefBuilder
	scratch2       *util.BytesRefBuilder
	scratchUTF16   *util.CharsRefBuilder
	scratchUTF16_2 *util.CharsRefBuilder
	pos            int
	payload        []byte
	nextDocStart   int64
	readOffsets    bool
	readPositions  bool
	startOffset    int
	endOffset      int
	cost           int
	skipReader     *SimpleTextSkipReader
	nextSkipDoc    int
	seekTo         int64
}

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