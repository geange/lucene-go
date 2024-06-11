package simpletext

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	index2 "github.com/geange/lucene-go/core/interface/index"
	"io"
	"strconv"
	"strings"

	"github.com/geange/lucene-go/codecs/utils"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
)

var _ index.DocValuesProducer = &DocValuesReader{}

type DocValuesReader struct {
	maxDoc  int
	data    store.IndexInput
	fields  map[string]*OneField
	scratch *bytes.Buffer
}

type OneField struct {
	dataStartFilePointer int64
	pattern              string
	ordPattern           string
	maxLength            int
	fixedLength          bool
	minValue             int64
	numValues            int64
}

func NewOneField() *OneField {
	return &OneField{}
}

func NewDocValuesReader(ctx context.Context, state *index.SegmentReadState, ext string) (*DocValuesReader, error) {
	r := &DocValuesReader{
		fields:  map[string]*OneField{},
		scratch: new(bytes.Buffer),
	}

	data, err := state.Directory.OpenInput(ctx, store.SegmentFileName(state.SegmentInfo.Name(), state.SegmentSuffix, ext))
	if err != nil {
		return nil, err
	}
	r.data = data

	maxDoc, err := state.SegmentInfo.MaxDoc()
	if err != nil {
		return nil, err
	}
	r.maxDoc = maxDoc

	reader := utils.NewTextReader(r.data, r.scratch)

	for {
		if err := reader.ReadLine(); err != nil {
			return nil, err
		}

		if bytes.Equal(r.scratch.Bytes(), DOC_VALUES_END) {
			break
		}

		if !bytes.HasPrefix(r.scratch.Bytes(), DOC_VALUES_FIELD) {
			return nil, errors.New(r.scratch.String())
		}
		fieldName := r.stripPrefix(DOC_VALUES_FIELD)

		field := NewOneField()
		r.fields[fieldName] = field

		docValuesType, err := reader.ReadLabel(DOC_VALUES_TYPE)
		if err != nil {
			return nil, err
		}

		dvType := document.StringToDocValuesType(docValuesType)
		if dvType == document.DOC_VALUES_TYPE_NONE {
			return nil, errors.New("dvType is NONE")
		}
		switch dvType {
		case document.DOC_VALUES_TYPE_NUMERIC:
			minValue, err := reader.ParseInt64(DOC_VALUES_MINVALUE)
			if err != nil {
				return nil, err
			}
			field.minValue = minValue

			pattern, err := reader.ReadLabel(DOC_VALUES_PATTERN)
			if err != nil {
				return nil, err
			}
			field.pattern = pattern

			field.dataStartFilePointer = r.data.GetFilePointer()
			offset := r.data.GetFilePointer() + int64((1+len(field.pattern)+2)*r.maxDoc)
			if _, err := r.data.Seek(offset, io.SeekStart); err != nil {
				return nil, err
			}
		case document.DOC_VALUES_TYPE_BINARY:
			maxLength, err := reader.ParseInt(DOC_VALUES_MAXLENGTH)
			if err != nil {
				return nil, err
			}
			field.maxLength = maxLength

			pattern, err := reader.ReadLabel(DOC_VALUES_PATTERN)
			if err != nil {
				return nil, err
			}
			field.pattern = pattern

			field.dataStartFilePointer = r.data.GetFilePointer()

			offset := r.data.GetFilePointer() + int64((1+len(field.pattern)+2)*r.maxDoc)

			if _, err := r.data.Seek(offset, io.SeekStart); err != nil {
				return nil, err
			}

		case document.DOC_VALUES_TYPE_SORTED, document.DOC_VALUES_TYPE_SORTED_SET:
			numValues, err := reader.ParseInt64(DOC_VALUES_NUMVALUES)
			if err != nil {
				return nil, err
			}
			field.numValues = numValues

			maxLength, err := reader.ParseInt(DOC_VALUES_MAXLENGTH)
			if err != nil {
				return nil, err
			}
			field.maxLength = maxLength

			pattern, err := reader.ReadLabel(DOC_VALUES_PATTERN)
			if err != nil {
				return nil, err
			}
			field.pattern = pattern

			ordPattern, err := reader.ReadLabel(DOC_VALUES_ORDPATTERN)
			if err != nil {
				return nil, err
			}
			field.ordPattern = ordPattern

			field.dataStartFilePointer = r.data.GetFilePointer()

			offset := r.data.GetFilePointer() +
				int64(9+len(field.pattern)+field.maxLength)*field.numValues +
				int64((1+len(field.ordPattern))*r.maxDoc)

			if _, err = r.data.Seek(offset, io.SeekStart); err != nil {
				return nil, err
			}

		default:
			return nil, errors.New("AssertionError")
		}
	}
	return r, nil
}

func (s *DocValuesReader) GetMergeInstance() index.DocValuesProducer {
	return s
}

func (s *DocValuesReader) Close() error {
	//TODO implement me
	panic("implement me")
}

func (s *DocValuesReader) GetNumeric(ctx context.Context, fieldInfo *document.FieldInfo) (index2.NumericDocValues, error) {
	numFn, err := s.getNumericNonIterator(fieldInfo)
	if err != nil {
		return nil, err
	}

	if numFn == nil {
		return nil, nil
	}

	docsWithField, err := s.getNumericDocsWithField(fieldInfo)
	if err != nil {
		return nil, err
	}

	return &index.NumericDocValuesDefault{
		FnDocID: func() int {
			return docsWithField.DocID()
		},
		FnNextDoc: func() (int, error) {
			return docsWithField.NextDoc()
		},
		FnAdvance: func(target int) (int, error) {
			return docsWithField.Advance(target)
		},
		FnSlowAdvance: func(target int) (int, error) {
			return docsWithField.Advance(target)
		},
		FnCost: func() int64 {
			return docsWithField.Cost()
		},
		FnAdvanceExact: func(target int) (bool, error) {
			return docsWithField.AdvanceExact(target)
		},
		FnLongValue: func() (int64, error) {
			return numFn(docsWithField.DocID())
		},
	}, nil
}

func (s *DocValuesReader) getNumericNonIterator(fieldInfo *document.FieldInfo) (func(value int) (int64, error), error) {

	field, ok := s.fields[fieldInfo.Name()]
	if !ok {
		return nil, fmt.Errorf("%s not found", fieldInfo.Name())
	}

	in := s.data.Clone().(store.IndexInput)
	scratch := new(bytes.Buffer)

	return func(docID int) (int64, error) {
		if docID < 0 || docID >= s.maxDoc {
			return 0, fmt.Errorf("docID must be 0 .. %d; got %d", s.maxDoc-1, docID)
		}

		if _, err := in.Seek(field.dataStartFilePointer+int64((1+(len(field.pattern))+2)*docID), io.SeekStart); err != nil {
			return 0, err
		}

		if err := utils.ReadLine(in, scratch); err != nil {
			return 0, err
		}

		num, err := strconv.Atoi(scratch.String())
		if err != nil {
			return 0, err
		}

		// read the line telling us if it's real or not
		if err := utils.ReadLine(in, scratch); err != nil {
			return 0, err
		}

		return field.minValue + int64(num), nil
	}, nil
}

func (s *DocValuesReader) getNumericDocsWithField(fieldInfo *document.FieldInfo) (DocValuesIterator, error) {
	return &innerDocValuesIterator1{
		field:  s.fields[fieldInfo.Name()],
		in:     s.data.Clone().(store.IndexInput),
		buf:    new(bytes.Buffer),
		reader: s,
		doc:    -1,
	}, nil
}

var _ DocValuesIterator = &innerDocValuesIterator1{}

type innerDocValuesIterator1 struct {
	field  *OneField
	in     store.IndexInput
	buf    *bytes.Buffer
	reader *DocValuesReader
	doc    int
}

func (i *innerDocValuesIterator1) DocID() int {
	return i.doc
}

func (i *innerDocValuesIterator1) NextDoc() (int, error) {
	return i.Advance(i.DocID() + 1)
}

func (i *innerDocValuesIterator1) Advance(target int) (int, error) {
	for idx := target; idx < i.reader.maxDoc; idx++ {
		offset := i.field.dataStartFilePointer + int64((1+len(i.field.pattern)+2)*idx)
		if _, err := i.in.Seek(offset, io.SeekStart); err != nil {
			return 0, err
		}

		// data
		if err := utils.ReadLine(i.in, i.buf); err != nil {
			return 0, err
		}
		// 'T' or 'F'
		if err := utils.ReadLine(i.in, i.buf); err != nil {
			return 0, err
		}

		if i.buf.Bytes()[0] == 'T' {
			i.doc = idx
			return i.doc, nil
		}
	}
	i.doc = types.NO_MORE_DOCS
	return i.doc, nil
}

func (i *innerDocValuesIterator1) SlowAdvance(target int) (int, error) {
	return i.Advance(target)
}

func (i *innerDocValuesIterator1) Cost() int64 {
	return int64(i.reader.maxDoc)
}

func (i *innerDocValuesIterator1) AdvanceExact(target int) (bool, error) {
	i.doc = target
	offset := i.field.dataStartFilePointer + int64((1+len(i.field.pattern)+2)*target)
	if _, err := i.in.Seek(offset, io.SeekStart); err != nil {
		return false, err
	}

	// data
	if err := utils.ReadLine(i.in, i.buf); err != nil {
		return false, err
	}
	// 'T' or 'F'
	if err := utils.ReadLine(i.in, i.buf); err != nil {
		return false, err
	}

	return i.buf.Bytes()[0] == 'T', nil

}

func (s *DocValuesReader) GetBinary(ctx context.Context, fieldInfo *document.FieldInfo) (index2.BinaryDocValues, error) {
	field, ok := s.fields[fieldInfo.Name()]
	if !ok {
		return nil, fmt.Errorf("%s not found", fieldInfo.Name())
	}

	in := s.data.Clone().(store.IndexInput)
	scratch := new(bytes.Buffer)

	docsWithField, err := s.getBinaryDocsWithField(fieldInfo)
	if err != nil {
		return nil, err
	}

	intFunc := func(docID int) ([]byte, error) {
		if docID < 0 || docID >= s.maxDoc {
			return nil, fmt.Errorf("docID must be 0 .. %d; got %d", s.maxDoc-1, docID)
		}

		offset := field.dataStartFilePointer + int64((9+len(field.pattern)+field.maxLength+2)*docID)
		if _, err := in.Seek(offset, io.SeekStart); err != nil {
			return nil, err
		}
		if err := utils.ReadLine(in, scratch); err != nil {
			return nil, err
		}
		if !bytes.HasPrefix(scratch.Bytes(), DOC_VALUES_LENGTH) {
			return nil, fmt.Errorf("%s", scratch.String())
		}

		scratch.Next(len(DOC_VALUES_LENGTH))
		size, err := strconv.Atoi(scratch.String())
		if err != nil {
			return nil, err
		}

		bs := make([]byte, size)
		if _, err = in.Read(bs); err != nil {
			return nil, err
		}
		return bs, nil
	}

	return &index.BaseBinaryDocValues{
		FnDocID:        docsWithField.DocID,
		FnNextDoc:      docsWithField.NextDoc,
		FnAdvance:      docsWithField.Advance,
		FnSlowAdvance:  docsWithField.Advance,
		FnCost:         docsWithField.Cost,
		FnAdvanceExact: docsWithField.AdvanceExact,
		FnBinaryValue: func() ([]byte, error) {
			return intFunc(docsWithField.DocID())
		},
	}, nil
}

func (s *DocValuesReader) getBinaryDocsWithField(fieldInfo *document.FieldInfo) (DocValuesIterator, error) {
	field := s.fields[fieldInfo.Name()]

	return &innerDocValuesIterator2{
		field:   field,
		in:      s.data.Clone().(store.IndexInput),
		scratch: new(bytes.Buffer),
		doc:     -1,
		reader:  s,
	}, nil
}

var _ DocValuesIterator = &innerDocValuesIterator2{}

type innerDocValuesIterator2 struct {
	field   *OneField
	in      store.IndexInput
	scratch *bytes.Buffer
	doc     int
	reader  *DocValuesReader
}

func (i *innerDocValuesIterator2) DocID() int {
	return i.doc
}

func (i *innerDocValuesIterator2) NextDoc() (int, error) {
	return i.Advance(i.DocID() + 1)
}

func (i *innerDocValuesIterator2) Advance(target int) (int, error) {
	for idx := target; idx < i.reader.maxDoc; idx++ {
		position := i.field.dataStartFilePointer + int64((9+len(i.field.pattern)+i.field.maxLength+2)*idx)
		if _, err := i.in.Seek(position, io.SeekStart); err != nil {
			return 0, err
		}
		if err := utils.ReadLine(i.in, i.scratch); err != nil {
			return 0, err
		}

		if bytes.HasPrefix(i.scratch.Bytes(), DOC_VALUES_LENGTH) {
			i.scratch.Next(len(DOC_VALUES_LENGTH))
		} else {
			return 0, errors.New(i.scratch.String())
		}

		size, err := strconv.Atoi(i.scratch.String())
		if err != nil {
			return 0, err
		}

		// skip past bytes
		if err := i.in.SkipBytes(nil, size); err != nil {
			return 0, err
		}

		// newline
		if err := utils.ReadLine(i.in, i.scratch); err != nil {
			return 0, err
		}
		// 'T' or 'F'
		if err := utils.ReadLine(i.in, i.scratch); err != nil {
			return 0, err
		}

		if i.scratch.Bytes()[0] == 'T' {
			i.doc = idx
			return i.doc, nil
		}
	}
	i.doc = types.NO_MORE_DOCS
	return i.doc, nil
}

func (i *innerDocValuesIterator2) SlowAdvance(target int) (int, error) {
	return i.Advance(target)
}

func (i *innerDocValuesIterator2) Cost() int64 {
	return int64(i.reader.maxDoc)
}

func (i *innerDocValuesIterator2) AdvanceExact(target int) (bool, error) {
	i.doc = target
	offset := i.field.dataStartFilePointer + int64((9+len(i.field.pattern)+i.field.maxLength+2)*target)
	if _, err := i.in.Seek(offset, io.SeekStart); err != nil {
		return false, err
	}
	value, err := utils.ReadValue(i.in, DOC_VALUES_LENGTH, i.scratch)
	if err != nil {
		return false, err
	}

	size, err := strconv.Atoi(value)
	if err != nil {
		return false, err
	}
	if err := i.in.SkipBytes(nil, size); err != nil {
		return false, err
	}
	if err := utils.ReadLine(i.in, i.scratch); err != nil {
		return false, err
	}
	if err := utils.ReadLine(i.in, i.scratch); err != nil {
		return false, err
	}
	return i.scratch.Bytes()[0] == 'T', nil
}

func (s *DocValuesReader) GetSorted(ctx context.Context, fieldInfo *document.FieldInfo) (index2.SortedDocValues, error) {
	field, ok := s.fields[fieldInfo.Name()]
	if !ok {
		return nil, fmt.Errorf("%s not found", fieldInfo.Name())
	}

	return newInnerSortedDocValues(field, s.data.Clone().(store.IndexInput), s), nil
}

var _ index2.SortedDocValues = &innerSortedDocValues{}

type innerSortedDocValues struct {
	*index.BaseSortedDocValues

	field   *OneField
	in      store.IndexInput
	doc     int
	ord     int
	reader  *DocValuesReader
	term    *bytes.Buffer
	scratch *bytes.Buffer
}

func newInnerSortedDocValues(field *OneField, in store.IndexInput, reader *DocValuesReader) *innerSortedDocValues {
	values := &innerSortedDocValues{
		BaseSortedDocValues: nil,
		field:               field,
		in:                  in,
		doc:                 -1,
		ord:                 0,
		reader:              reader,
		term:                new(bytes.Buffer),
		scratch:             new(bytes.Buffer),
	}

	values.BaseSortedDocValues = index.NewBaseSortedDocValues(&index.SortedDocValuesDefaultConfig{
		OrdValue:      values.OrdValue,
		LookupOrd:     values.LookupOrd,
		GetValueCount: values.GetValueCount,
	})
	return values
}

func (i *innerSortedDocValues) OrdValue() (int, error) {
	return i.ord, nil
}

func (i *innerSortedDocValues) LookupOrd(ord int) ([]byte, error) {
	if ord < 0 || ord >= int(i.field.numValues) {
		return nil, fmt.Errorf("ord must be 0 .. %d; got %d", i.field.numValues-1, ord)
	}

	offset := i.field.dataStartFilePointer + int64(ord*(9+len(i.field.pattern)+i.field.maxLength))

	if _, err := i.in.Seek(offset, io.SeekStart); err != nil {
		return nil, err
	}

	value, err := utils.ReadValue(i.in, DOC_VALUES_LENGTH, i.scratch)
	if err != nil {
		return nil, err
	}

	size, err := strconv.Atoi(value)
	if err != nil {
		return nil, err
	}

	bs := make([]byte, size)
	if _, err = i.in.Read(bs); err != nil {
		return nil, err
	}
	return bs, nil
}

func (i *innerSortedDocValues) GetValueCount() int {
	return int(i.field.numValues)
}

func (i *innerSortedDocValues) DocID() int {
	return i.doc
}

func (i *innerSortedDocValues) NextDoc() (int, error) {
	return i.Advance(i.DocID() + 1)
}

func (i *innerSortedDocValues) Advance(target int) (int, error) {
	for idx := target; idx < i.reader.maxDoc; idx++ {
		offset := i.field.dataStartFilePointer +
			i.field.numValues*int64(9+len(i.field.pattern)+i.field.maxLength) +
			int64(idx*(1+len(i.field.ordPattern)))
		if _, err := i.in.Seek(offset, io.SeekStart); err != nil {
			return 0, err
		}
		if err := utils.ReadLine(i.in, i.scratch); err != nil {
			return 0, err
		}

		ord, err := strconv.Atoi(i.scratch.String())
		if err != nil {
			return 0, err
		}

		if ord >= 0 {
			i.doc = idx
			return i.doc, nil
		}
	}
	i.doc = types.NO_MORE_DOCS
	return i.doc, nil
}

func (i *innerSortedDocValues) SlowAdvance(target int) (int, error) {
	return i.Advance(target)
}

func (i *innerSortedDocValues) Cost() int64 {
	return int64(i.reader.maxDoc)
}

func (i *innerSortedDocValues) AdvanceExact(target int) (bool, error) {
	i.doc = target
	offset := i.field.dataStartFilePointer +
		i.field.numValues*int64(9+len(i.field.pattern)+i.field.maxLength) +
		int64(target*(1+len(i.field.ordPattern)))
	if _, err := i.in.Seek(offset, io.SeekStart); err != nil {
		return false, err
	}
	if err := utils.ReadLine(i.in, i.scratch); err != nil {
		return false, err
	}

	ord, err := strconv.Atoi(i.scratch.String())
	if err != nil {
		return false, err
	}
	return ord >= 0, nil
}

func (i *innerSortedDocValues) TermsEnum() (index2.TermsEnum, error) {
	return index.NewSortedDocValuesTermsEnum(i), nil
}

func (s *DocValuesReader) GetSortedNumeric(ctx context.Context, fieldInfo *document.FieldInfo) (index2.SortedNumericDocValues, error) {
	binary, err := s.GetBinary(ctx, fieldInfo)
	if err != nil {
		return nil, err
	}
	return newInnerSortedNumericDocValues(binary), nil
}

var _ index2.SortedNumericDocValues = &innerSortedNumericDocValues{}

type innerSortedNumericDocValues struct {
	values []int64
	index  int
	binary index2.BinaryDocValues
}

func newInnerSortedNumericDocValues(binary index2.BinaryDocValues) *innerSortedNumericDocValues {
	return &innerSortedNumericDocValues{
		values: make([]int64, 0),
		index:  0,
		binary: binary,
	}
}

func (i *innerSortedNumericDocValues) DocID() int {
	return i.binary.DocID()
}

func (i *innerSortedNumericDocValues) NextDoc() (int, error) {
	doc, err := i.binary.NextDoc()
	if err != nil {
		return 0, err
	}

	if err := i.setCurrentDoc(); err != nil {
		return 0, err
	}
	return doc, nil
}

func (i *innerSortedNumericDocValues) Advance(target int) (int, error) {
	doc, err := i.binary.Advance(target)
	if err != nil {
		return 0, err
	}

	if err := i.setCurrentDoc(); err != nil {
		return 0, err
	}
	return doc, nil
}

func (i *innerSortedNumericDocValues) SlowAdvance(target int) (int, error) {
	return i.Advance(target)
}

func (i *innerSortedNumericDocValues) Cost() int64 {
	return i.binary.Cost()
}

func (i *innerSortedNumericDocValues) AdvanceExact(target int) (bool, error) {
	ok, err := i.binary.AdvanceExact(target)
	if err != nil {
		return false, err
	}

	if ok {
		if err := i.setCurrentDoc(); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func (i *innerSortedNumericDocValues) NextValue() (int64, error) {
	value := i.values[i.index]
	i.index++
	return value, nil
}

func (i *innerSortedNumericDocValues) DocValueCount() int {
	return len(i.values)
}

func (i *innerSortedNumericDocValues) setCurrentDoc() error {
	if i.DocID() == types.NO_MORE_DOCS {
		return nil
	}
	bs, err := i.binary.BinaryValue()
	if err != nil {
		return err
	}
	csv := string(bs)

	if len(csv) == 0 {
		i.values = make([]int64, 0)
	} else {
		s := strings.Split(csv, ",")

		i.values = i.values[:0]

		for _, v := range s {
			num, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			i.values = append(i.values, int64(num))
		}
	}
	i.index = 0
	return nil
}

func (s *DocValuesReader) GetSortedSet(ctx context.Context, fieldInfo *document.FieldInfo) (index2.SortedSetDocValues, error) {
	field, ok := s.fields[fieldInfo.Name()]
	if !ok {
		return nil, fmt.Errorf("%s not found", fieldInfo.Name())
	}

	return &innerSortedSetDocValues{
		field:        field,
		in:           s.data.Clone().(store.IndexInput),
		reader:       s,
		currentOrds:  []string{},
		currentIndex: 0,
		term:         new(bytes.Buffer),
		scratch:      new(bytes.Buffer),
		doc:          -1,
	}, nil
}

var _ index2.SortedSetDocValues = &innerSortedSetDocValues{}

type innerSortedSetDocValues struct {
	field        *OneField
	in           store.IndexInput
	reader       *DocValuesReader
	currentOrds  []string
	currentIndex int
	term         *bytes.Buffer
	scratch      *bytes.Buffer
	doc          int
}

func (i *innerSortedSetDocValues) DocID() int {
	return i.doc
}

func (i *innerSortedSetDocValues) NextDoc() (int, error) {
	return i.Advance(i.doc + 1)
}

func (i *innerSortedSetDocValues) Advance(target int) (int, error) {
	for idx := target; idx < i.reader.maxDoc; idx++ {
		offset := i.field.dataStartFilePointer + i.field.numValues*
			int64(9+len(i.field.pattern)+i.field.maxLength) +
			int64(idx*(1+len(i.field.ordPattern)))

		if _, err := i.in.Seek(offset, io.SeekStart); err != nil {
			return 0, err
		}

		if err := utils.ReadLine(i.in, i.scratch); err != nil {
			return 0, err
		}

		ordList := strings.Trim(i.scratch.String(), " ")

		if len(ordList) > 0 {
			i.currentOrds = strings.Split(ordList, ",")
			i.currentIndex = 0
			i.doc = idx
			return i.doc, nil
		}
	}
	i.doc = types.NO_MORE_DOCS
	return i.doc, nil
}

func (i *innerSortedSetDocValues) SlowAdvance(target int) (int, error) {
	return i.Advance(target)
}

func (i *innerSortedSetDocValues) Cost() int64 {
	return int64(i.reader.maxDoc)
}

func (i *innerSortedSetDocValues) AdvanceExact(target int) (bool, error) {
	offset := i.field.dataStartFilePointer +
		i.field.numValues*int64(9+len(i.field.pattern)+i.field.maxLength) +
		int64(target*(1+len(i.field.ordPattern)))

	if _, err := i.in.Seek(offset, io.SeekStart); err != nil {
		return false, err
	}

	if err := utils.ReadLine(i.in, i.scratch); err != nil {
		return false, err
	}

	ordList := strings.Trim(i.scratch.String(), " ")
	i.doc = target

	if len(ordList) != 0 {
		i.currentOrds = strings.Split(ordList, ",")
		i.currentIndex = 0
		return true, nil
	}
	return false, nil
}

func (i *innerSortedSetDocValues) NextOrd() (int64, error) {
	if i.currentIndex == len(i.currentOrds) {
		return index.NO_MORE_ORDS, nil
	} else {
		num, err := strconv.Atoi(i.currentOrds[i.currentIndex])
		if err != nil {
			return 0, err
		}
		i.currentIndex++

		return int64(num), nil
	}
}

func (i *innerSortedSetDocValues) LookupOrd(ord int64) ([]byte, error) {
	if ord < 0 || ord >= i.field.numValues {
		return nil, fmt.Errorf("ord must be 0 .. %d; git %d", i.field.numValues-1, ord)
	}

	offset := i.field.dataStartFilePointer + ord*int64(9+len(i.field.pattern)+i.field.maxLength)
	if _, err := i.in.Seek(offset, io.SeekStart); err != nil {
		return nil, err
	}

	value, err := utils.ReadValue(i.in, DOC_VALUES_LENGTH, i.scratch)
	if err != nil {
		return nil, err
	}
	size, err := strconv.Atoi(value)
	if err != nil {
		return nil, err
	}
	bs := make([]byte, size)
	if _, err = i.in.Read(bs); err != nil {
		return nil, err
	}
	return bs, nil
}

func (i *innerSortedSetDocValues) GetValueCount() int64 {
	return i.field.numValues
}

func (s *DocValuesReader) CheckIntegrity() error {
	scratch := new(bytes.Buffer)
	clone := s.data.Clone().(store.IndexInput)

	if _, err := clone.Seek(0, io.SeekStart); err != nil {
		return err
	}

	// checksum is fixed-width encoded with 20 bytes,
	// plus 1 byte for newline (the space is included in SimpleTextUtil.CHECKSUM):
	footerStartPos := s.data.Length() - int64(len(utils.CHECKSUM)+21)

	input := store.NewBufferedChecksumIndexInput(clone)

	for {
		if err := utils.ReadLine(input, scratch); err != nil {
			return err
		}
		if input.GetFilePointer() >= footerStartPos {
			// Make sure we landed at precisely the right location:
			if input.GetFilePointer() != footerStartPos {
				return fmt.Errorf("SimpleText failure: "+
					"footer does not start at expected position current=%d vs expected=%d",
					input.GetFilePointer(), footerStartPos)
			}
			if err := utils.CheckFooter(input); err != nil {
				return err
			}
			break
		}
	}
	return nil
}

func (s *DocValuesReader) readLine() error {
	return utils.ReadLine(s.data, s.scratch)
}

func (s *DocValuesReader) stripPrefix(field []byte) string {
	return string(s.scratch.Bytes()[len(field):])
}

type DocValuesIterator interface {
	types.DocIdSetIterator

	AdvanceExact(target int) (bool, error)
}
