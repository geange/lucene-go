package index

import (
	"bytes"
	"context"
	"io"

	"github.com/geange/lucene-go/core/document"
)

var _ Fields = &DataFields{}

type DataFields struct {
	fields []*FieldData
}

func NewDataFields(fields []*FieldData) *DataFields {
	return &DataFields{fields: fields}
}

func (d *DataFields) Names() []string {
	values := make([]string, 0)
	for _, field := range d.fields {
		values = append(values, field.fieldInfo.Name())
	}
	return values
}

func (d *DataFields) Terms(field string) (Terms, error) {
	for _, fieldData := range d.fields {
		if fieldData.fieldInfo.Name() == field {
			return NewDataTerms(fieldData), nil
		}
	}
	return nil, nil
}

func (d *DataFields) Size() int {
	return len(d.fields)
}

var _ Terms = &DataTerms{}

type DataTerms struct {
	*TermsBase

	fieldData *FieldData
}

func NewDataTerms(fieldData *FieldData) *DataTerms {
	terms := &DataTerms{fieldData: fieldData}
	terms.TermsBase = NewTerms(terms)
	return terms
}

func (d *DataTerms) Iterator() (TermsEnum, error) {
	return NewDataTermsEnum(d.fieldData), nil
}

func (d *DataTerms) Size() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DataTerms) GetSumTotalTermFreq() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DataTerms) GetSumDocFreq() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DataTerms) GetDocCount() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DataTerms) HasFreqs() bool {
	return d.fieldData.fieldInfo.GetIndexOptions() >=
		document.INDEX_OPTIONS_DOCS_AND_FREQS
}

func (d *DataTerms) HasOffsets() bool {
	return d.fieldData.fieldInfo.GetIndexOptions() >=
		document.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS
}

func (d *DataTerms) HasPositions() bool {
	return d.fieldData.fieldInfo.GetIndexOptions() >=
		document.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS
}

func (d *DataTerms) HasPayloads() bool {
	return d.fieldData.fieldInfo.HasPayloads()
}

var _ TermsEnum = &DataTermsEnum{}

type DataTermsEnum struct {
	*BaseTermsEnum

	fieldData *FieldData
	upto      int
}

func NewDataTermsEnum(fieldData *FieldData) *DataTermsEnum {
	termEnum := &DataTermsEnum{
		fieldData: fieldData,
		upto:      -1,
	}
	termEnum.BaseTermsEnum = NewBaseTermsEnum(&BaseTermsEnumConfig{SeekCeil: termEnum.SeekCeil})
	return termEnum
}

func (d *DataTermsEnum) Next(context.Context) ([]byte, error) {
	d.upto++
	if d.upto == len(d.fieldData.terms) {
		return nil, io.EOF
	}
	return d.Term()
}

func (d *DataTermsEnum) SeekCeil(ctx context.Context, text []byte) (SeekStatus, error) {
	// Stupid linear impl:
	for i := 0; i < len(d.fieldData.terms); i++ {
		cmp := bytes.Compare(d.fieldData.terms[i].text, text)
		if cmp == 0 {
			d.upto = i
			return SEEK_STATUS_FOUND, nil
		} else if cmp > 0 {
			d.upto = i
			return SEEK_STATUS_NOT_FOUND, nil
		}
	}
	return SEEK_STATUS_END, nil
}

func (d *DataTermsEnum) SeekExactByOrd(ctx context.Context, ord int64) error {
	//TODO implement me
	panic("implement me")
}

func (d *DataTermsEnum) Term() ([]byte, error) {
	return d.fieldData.terms[d.upto].text, nil
}

func (d *DataTermsEnum) Ord() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DataTermsEnum) DocFreq() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DataTermsEnum) TotalTermFreq() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DataTermsEnum) Postings(reuse PostingsEnum, flags int) (PostingsEnum, error) {
	return newDataPostingsEnum(d.fieldData.terms[d.upto]), nil
}

func (d *DataTermsEnum) Impacts(flags int) (ImpactsEnum, error) {
	//TODO implement me
	panic("implement me")
}

var _ PostingsEnum = &DataPostingsEnum{}

type DataPostingsEnum struct {
	termData *TermData
	docUpto  int
	posUpto  int
}

func newDataPostingsEnum(termData *TermData) *DataPostingsEnum {
	return &DataPostingsEnum{
		termData: termData,
		docUpto:  -1,
		posUpto:  0,
	}
}

func (d *DataPostingsEnum) DocID() int {
	return d.termData.docs[d.docUpto]
}

func (d *DataPostingsEnum) NextDoc() (int, error) {
	d.docUpto++
	if d.docUpto == len(d.termData.docs) {
		return 0, io.EOF
	}
	d.posUpto = -1
	return d.DocID(), nil
}

func (d *DataPostingsEnum) Advance(target int) (int, error) {
	// Slow linear impl:
	if _, err := d.NextDoc(); err != nil {
		return 0, err
	}
	for d.DocID() < target {
		if _, err := d.NextDoc(); err != nil {
			return 0, err
		}
	}

	return d.DocID(), nil
}

func (d *DataPostingsEnum) SlowAdvance(target int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DataPostingsEnum) Cost() int64 {
	//TODO implement me
	panic("implement me")
}

func (d *DataPostingsEnum) Freq() (int, error) {
	return len(d.termData.positions[d.docUpto]), nil
}

func (d *DataPostingsEnum) NextPosition() (int, error) {
	d.posUpto++
	return d.termData.positions[d.docUpto][d.posUpto].Pos, nil
}

func (d *DataPostingsEnum) StartOffset() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DataPostingsEnum) EndOffset() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DataPostingsEnum) GetPayload() ([]byte, error) {
	return d.termData.positions[d.docUpto][d.posUpto].Payload, nil
}
