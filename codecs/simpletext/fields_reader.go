package simpletext

import (
	"bytes"

	"github.com/emirpasic/gods/maps/hashmap"
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
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
