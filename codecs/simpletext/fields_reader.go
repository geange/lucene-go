package simpletext

import (
	"bytes"

	"github.com/emirpasic/gods/maps/treemap"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
)

var (
	_ index.FieldsProducer = &FieldsReader{}
	_ index.Fields         = &FieldsReader{}
)

type FieldsReader struct {
	fields     *treemap.Map
	in         store.IndexInput
	fieldInfos *index.FieldInfos
	maxDoc     int
	termsCache map[string]*fieldsReaderTerm
}

func (r *FieldsReader) Names() []string {
	keys := make([]string, 0)
	r.fields.All(func(key interface{}, value interface{}) bool {
		keys = append(keys, key.(string))
		return true
	})
	return keys
}

func (r *FieldsReader) Terms(field string) (index.Terms, error) {
	v, ok := r.termsCache[field]
	if !ok {
		fp, ok := r.fields.Get(field)
		if !ok {
			return nil, nil
		}
		terms, err := r.newFieldsReaderTerm(field, fp.(int64), r.maxDoc)
		if err != nil {
			return nil, err
		}
		r.termsCache[field] = terms
		return terms, nil
	}
	return v, nil
}

func (r *FieldsReader) Size() int {
	return -1
}

func NewFieldsReader(state *index.SegmentReadState) (*FieldsReader, error) {
	maxDoc, err := state.SegmentInfo.MaxDoc()
	if err != nil {
		return nil, err
	}

	name := getPostingsFileName(state.SegmentInfo.Name, state.SegmentSuffix)
	input, err := state.Directory.OpenInput(name, state.Context)
	if err != nil {
		return nil, err
	}

	reader := &FieldsReader{
		fields:     nil,
		in:         input,
		fieldInfos: state.FieldInfos,
		maxDoc:     maxDoc,
		termsCache: make(map[string]*fieldsReaderTerm),
	}

	fields, err := reader.readFields(reader.in.Clone())
	if err != nil {
		_ = input.Close()
		return nil, err
	}
	reader.fields = fields
	return reader, nil
}

func (r *FieldsReader) readFields(in store.IndexInput) (*treemap.Map, error) {
	input := store.NewBufferedChecksumIndexInput(in)
	scratch := new(bytes.Buffer)
	fields := treemap.NewWithStringComparator()

	for {
		if err := ReadLine(input, scratch); err != nil {
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

func (r *FieldsReader) Close() error {
	return r.in.Close()
}

func (r *FieldsReader) CheckIntegrity() error {
	return nil
}

func (r *FieldsReader) GetMergeInstance() index.FieldsProducer {
	return nil
}
