package index

import (
	"bytes"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util/automaton"
	"io"
)

var _ Fields = &FreqProxFields{}

type FreqProxFields struct {
	fields map[string]*FreqProxTermsWriterPerField
}

func NewFreqProxFields(fieldList []*FreqProxTermsWriterPerField) *FreqProxFields {
	fields := make(map[string]*FreqProxTermsWriterPerField)
	for _, field := range fieldList {
		fields[field.getFieldName()] = field
	}
	return &FreqProxFields{fields: fields}
}

func (f *FreqProxFields) Names() []string {
	//TODO implement me
	panic("implement me")
}

func (f *FreqProxFields) Terms(field string) (Terms, error) {
	perField, ok := f.fields[field]
	if !ok {
		return nil, nil
	}
	return NewFreqProxTerms(perField), nil
}

func (f *FreqProxFields) Size() int {
	return len(f.fields)
}

var _ Terms = &FreqProxTerms{}

type FreqProxTerms struct {
	terms *FreqProxTermsWriterPerField
}

func NewFreqProxTerms(terms *FreqProxTermsWriterPerField) *FreqProxTerms {
	return &FreqProxTerms{terms: terms}
}

func (f *FreqProxTerms) Iterator() (TermsEnum, error) {
	termsEnum := NewFreqProxTermsEnum(f.terms)
	termsEnum.reset()
	return termsEnum, nil
}

func (f *FreqProxTerms) Intersect(compiled *automaton.CompiledAutomaton, startTerm []byte) (TermsEnum, error) {
	//TODO implement me
	panic("implement me")
}

func (f *FreqProxTerms) Size() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (f *FreqProxTerms) GetSumTotalTermFreq() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (f *FreqProxTerms) GetSumDocFreq() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (f *FreqProxTerms) GetDocCount() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (f *FreqProxTerms) HasFreqs() bool {
	return f.terms.indexOptions >=
		types.INDEX_OPTIONS_DOCS_AND_FREQS
}

func (f *FreqProxTerms) HasOffsets() bool {
	return f.terms.indexOptions >=
		types.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS
}

func (f *FreqProxTerms) HasPositions() bool {
	return f.terms.indexOptions >=
		types.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS
}

func (f *FreqProxTerms) HasPayloads() bool {
	return f.terms.sawPayloads
}

func (f *FreqProxTerms) GetMin() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (f *FreqProxTerms) GetMax() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

var _ TermsEnum = &FreqProxTermsEnum{}

type FreqProxTermsEnum struct {
	*BaseTermsEnum

	terms         *FreqProxTermsWriterPerField
	sortedTermIDs []int
	postingsArray *FreqProxPostingsArray
	numTerms      int
	ord           int
	scratch       []byte
}

func NewFreqProxTermsEnum(terms *FreqProxTermsWriterPerField) *FreqProxTermsEnum {
	termEnum := &FreqProxTermsEnum{
		terms:         terms,
		sortedTermIDs: terms.getSortedTermIDs(),
		postingsArray: terms.postingsArray.(*FreqProxPostingsArray),
		numTerms:      terms.getNumTerms(),
		ord:           0,
	}
	termEnum.BaseTermsEnum = NewBaseTermsEnum(&BaseTermsEnumConfig{
		SeekCeil: termEnum.SeekCeil,
	})
	return termEnum
}

func (f *FreqProxTermsEnum) reset() {
	f.ord = -1
}

func (f *FreqProxTermsEnum) Next() ([]byte, error) {
	f.ord++
	if f.ord >= f.numTerms {
		return nil, io.EOF
	}
	textStart := f.postingsArray.textStarts[f.sortedTermIDs[f.ord]]
	var err error
	f.scratch, err = f.terms.bytePool.GetAddress(textStart)
	return f.scratch, err
}

func (f *FreqProxTermsEnum) SeekCeil(text []byte) (SeekStatus, error) {
	// binary search:
	lo := 0
	hi := f.numTerms - 1
	var err error
	for hi >= lo {
		mid := (lo + hi) >> 1
		textStart := f.postingsArray.textStarts[f.sortedTermIDs[mid]]
		f.scratch, err = f.terms.bytePool.GetAddress(textStart)
		if err != nil {
			return 0, err
		}
		cmp := bytes.Compare(f.scratch, text)
		if cmp < 0 {
			lo = mid + 1
		} else if cmp > 0 {
			hi = mid - 1
		} else {
			// found:
			f.ord = mid
			//assert term().compareTo(text) == 0;
			return SEEK_STATUS_FOUND, nil
		}
	}

	// not found:
	f.ord = lo
	if f.ord >= f.numTerms {
		return SEEK_STATUS_END, nil
	} else {
		textStart := f.postingsArray.textStarts[f.sortedTermIDs[f.ord]]
		f.scratch, _ = f.terms.bytePool.GetAddress(textStart)
		//assert term().compareTo(text) > 0;
		return SEEK_STATUS_NOT_FOUND, nil
	}
}

func (f *FreqProxTermsEnum) SeekExactByOrd(ord int64) error {
	f.ord = int(ord)
	textStart := f.postingsArray.textStarts[f.sortedTermIDs[f.ord]]
	var err error
	f.scratch, err = f.terms.bytePool.GetAddress(textStart)
	return err
}

func (f *FreqProxTermsEnum) Term() ([]byte, error) {
	return f.scratch, nil
}

func (f *FreqProxTermsEnum) Ord() (int64, error) {
	return int64(f.ord), nil
}

func (f *FreqProxTermsEnum) DocFreq() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (f *FreqProxTermsEnum) TotalTermFreq() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (f *FreqProxTermsEnum) Postings(reuse PostingsEnum, flags int) (PostingsEnum, error) {
	//TODO implement me
	panic("implement me")
}

func (f *FreqProxTermsEnum) Impacts(flags int) (ImpactsEnum, error) {
	//TODO implement me
	panic("implement me")
}
