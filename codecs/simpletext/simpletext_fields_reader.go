package simpletext

import (
	"bytes"

	"github.com/geange/gods-generic/maps/treemap"
	"github.com/geange/lucene-go/codecs/utils"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
)

var (
	_ index.FieldsProducer = &SimpleTextFieldsReader{}
	_ index.Fields         = &SimpleTextFieldsReader{}
)

type SimpleTextFieldsReader struct {
	fields     *treemap.Map[string, int64]
	in         store.IndexInput
	fieldInfos *index.FieldInfos
	maxDoc     int
	termsCache map[string]*simpleTextTerms
}

func (s *SimpleTextFieldsReader) Names() []string {
	return s.fields.Keys()
}

func (s *SimpleTextFieldsReader) Terms(field string) (index.Terms, error) {
	v, ok := s.termsCache[field]
	if !ok {
		fp, ok := s.fields.Get(field)
		if !ok {
			return nil, nil
		}
		terms, err := s.newFieldsReaderTerm(field, fp, s.maxDoc)
		if err != nil {
			return nil, err
		}
		s.termsCache[field] = terms
		return terms, nil
	}
	return v, nil
}

func (s *SimpleTextFieldsReader) Size() int {
	return -1
}

func NewSimpleTextFieldsReader(state *index.SegmentReadState) (*SimpleTextFieldsReader, error) {
	maxDoc, err := state.SegmentInfo.MaxDoc()
	if err != nil {
		return nil, err
	}

	name := getPostingsFileName(state.SegmentInfo.Name(), state.SegmentSuffix)
	input, err := state.Directory.OpenInput(name, state.Context)
	if err != nil {
		return nil, err
	}

	reader := &SimpleTextFieldsReader{
		fields:     nil,
		in:         input,
		fieldInfos: state.FieldInfos,
		maxDoc:     maxDoc,
		termsCache: make(map[string]*simpleTextTerms),
	}

	fields, err := reader.readFields(reader.in.Clone())
	if err != nil {
		_ = input.Close()
		return nil, err
	}
	reader.fields = fields
	return reader, nil
}

func (s *SimpleTextFieldsReader) readFields(in store.IndexInput) (*treemap.Map[string, int64], error) {
	input := store.NewBufferedChecksumIndexInput(in)
	scratch := new(bytes.Buffer)
	fields := treemap.New[string, int64]()

	for {
		if err := utils.ReadLine(input, scratch); err != nil {
			return nil, err
		}

		text := scratch.Bytes()

		if bytes.Equal(text, FIELDS_END) {
			return fields, nil
		} else if bytes.HasPrefix(text, FIELDS_FIELD) {
			fieldName := string(text[len(FIELDS_FIELD):])
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
	return s
}
