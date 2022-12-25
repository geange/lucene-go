package simpletext

import (
	"bytes"
	"errors"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/tokenattributes"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util/automaton"
	"github.com/geange/lucene-go/core/util/fst"
	"strconv"
)

var _ index.Terms = &fieldsReaderTerm{}

type fieldsReaderTerm struct {
	termsStart       int64
	fieldInfo        *types.FieldInfo
	maxDoc           int
	sumTotalTermFreq int64
	sumDocFreq       int64
	docCount         int
	fst              *fst.FST[*fst.Pair[*fst.Pair[int64, int64], *fst.Pair[int64, int64]]]
	termCount        int

	scratch *bytes.Buffer
}

func (r *FieldsReader) newFieldsReaderTerm(field string, termsStart int64, maxDoc int) (*fieldsReaderTerm, error) {
	info := r.fieldInfos.FieldInfo(field)
	term := &fieldsReaderTerm{
		termsStart: termsStart,
		fieldInfo:  info,
		maxDoc:     maxDoc,
	}
	if err := r.loadTerms(term); err != nil {
		return nil, err
	}
	return term, nil
}

func (r *FieldsReader) loadTerms(term *fieldsReaderTerm) error {
	posIntOutputs := fst.NewPositiveIntOutputs()
	outputsOuter := fst.NewPairOutputs[int64, int64](posIntOutputs, posIntOutputs)
	outputsInner := fst.NewPairOutputs[int64, int64](posIntOutputs, posIntOutputs)
	outputs := fst.NewPairOutputs[*fst.Pair[int64, int64], *fst.Pair[int64, int64]](outputsOuter, outputsInner)
	fstCompiler := fst.NewBuilder[*fst.Pair[*fst.Pair[int64, int64], *fst.Pair[int64, int64]]](fst.BYTE1, outputs)

	in := r.in.Clone()
	if _, err := in.Seek(term.termsStart, 0); err != nil {
		return err
	}
	lastTerm := new(bytes.Buffer)

	lastDocsStart := int64(-1)
	docFreq := int64(0)
	totalTermFreq := int64(0)
	skipPointer := int64(0)

	visitedDocs := make(map[int]struct{}, term.maxDoc)
	for {
		if err := ReadLine(in, term.scratch); err != nil {
			return err
		}

		text := term.scratch.Bytes()

		if bytes.Equal(text, FIELDS_END) || bytes.HasPrefix(text, FIELDS_FIELD) {
			if lastDocsStart != -1 {
				output := fst.NewPair(fst.NewPair(lastDocsStart, skipPointer), fst.NewPair(docFreq, totalTermFreq))
				if err := fstCompiler.Add(bytes.Runes(lastTerm.Bytes()), output); err != nil {
					return err
				}
				term.sumTotalTermFreq += totalTermFreq
			}
			break
		} else if bytes.HasPrefix(text, FIELDS_DOC) {
			docFreq++
			term.sumDocFreq++
			totalTermFreq++

			docID, err := strconv.Atoi(string(text[len(FIELDS_DOC):]))
			if err != nil {
				return err
			}
			visitedDocs[docID] = struct{}{}

		} else if bytes.HasPrefix(text, FIELDS_FREQ) {
			freq, err := strconv.Atoi(string(text[len(FIELDS_FREQ):]))
			if err != nil {
				return err
			}
			totalTermFreq += int64(freq) - 1
		} else if bytes.HasPrefix(text, SKIP_LIST) {
			skipPointer = in.GetFilePointer()
		} else if bytes.HasPrefix(text, FIELDS_TERM) {
			if lastDocsStart != -1 {
				output := fst.NewPair(
					fst.NewPair(lastDocsStart, skipPointer),
					fst.NewPair(docFreq, totalTermFreq),
				)
				if err := fstCompiler.Add(bytes.Runes(lastTerm.Bytes()), output); err != nil {
					return err
				}
			}

			lastDocsStart = in.GetFilePointer()
			lastTerm.Reset()
			lastTerm.Write(text[len(FIELDS_TERM):])

			docFreq = 0
			term.sumTotalTermFreq += totalTermFreq
			totalTermFreq = 0
			term.termCount++
			skipPointer = 0
		}
	}
	term.termCount = len(visitedDocs)

	var err error
	term.fst, err = fstCompiler.Finish()
	return err
}

func (f *fieldsReaderTerm) Iterator() (index.TermsEnum, error) {
	//TODO implement me
	panic("implement me")
}

func (f *fieldsReaderTerm) Intersect(compiled *automaton.CompiledAutomaton, startTerm []byte) (index.TermsEnum, error) {
	//TODO implement me
	panic("implement me")
}

func (f *fieldsReaderTerm) Size() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (f *fieldsReaderTerm) GetSumTotalTermFreq() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (f *fieldsReaderTerm) GetSumDocFreq() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (f *fieldsReaderTerm) GetDocCount() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (f *fieldsReaderTerm) HasFreqs() bool {
	//TODO implement me
	panic("implement me")
}

func (f *fieldsReaderTerm) HasOffsets() bool {
	//TODO implement me
	panic("implement me")
}

func (f *fieldsReaderTerm) HasPositions() bool {
	//TODO implement me
	panic("implement me")
}

func (f *fieldsReaderTerm) HasPayloads() bool {
	//TODO implement me
	panic("implement me")
}

func (f *fieldsReaderTerm) GetMin() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (f *fieldsReaderTerm) GetMax() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

var _ index.BaseTermsEnum = &termsEnum{}

type EnumPair fst.Pair[*fst.Pair[int64, int64], *fst.Pair[int64, int64]]

type BytesEnum fst.BytesRefFSTEnum[*EnumPair]

type termsEnum struct {
	indexOptions  types.IndexOptions
	docFreq       int
	totalTermFreq int64
	docsStart     int64
	skipPointer   int64
	ended         bool
	fstEnum       *fst.BytesRefFSTEnum[*EnumPair]
}

func newTermsEnum(fstInstance *fst.FST[*EnumPair], indexOptions types.IndexOptions) *termsEnum {
	return &termsEnum{
		indexOptions: indexOptions,
		fstEnum:      fst.NewBytesRefFSTEnum(fstInstance),
	}
}

func (t *termsEnum) Attributes() *tokenattributes.AttributeSource {
	//TODO implement me
	panic("implement me")
}

func (t *termsEnum) SeekExact(text []byte) (bool, error) {
	result, err := t.fstEnum.SeekExact(text)
	if err != nil {
		return false, err
	}

	if result != nil {
		pair := result.Output
		pair1 := pair.Output1
		pair2 := pair.Output2

		t.docsStart = pair1.Output1
		t.skipPointer = pair1.Output2
		t.docFreq = int(pair2.Output1)
		t.totalTermFreq = pair2.Output2
		return true, nil
	}
	return false, nil
}

func (t *termsEnum) SeekExactExpert(term []byte, state index.TermState) error {
	//TODO implement me
	panic("implement me")
}

func (t *termsEnum) TermState() (index.TermState, error) {
	//TODO implement me
	panic("implement me")
}

func (t *termsEnum) Next() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (t *termsEnum) SeekCeil(text []byte) (index.SeekStatus, error) {
	//TODO implement me
	panic("implement me")
}

func (t *termsEnum) SeekExactByOrd(ord int64) error {
	return errors.New("UnsupportedOperationException")
}

func (t *termsEnum) Term() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (t *termsEnum) Ord() (int64, error) {
	return 0, errors.New("UnsupportedOperationException")
}

func (t *termsEnum) DocFreq() (int, error) {
	return t.docFreq, nil
}

func (t *termsEnum) TotalTermFreq() (int64, error) {
	if t.indexOptions == types.INDEX_OPTIONS_DOCS {
		return int64(t.docFreq), nil
	}
	return t.totalTermFreq, nil
}

func (t *termsEnum) Postings(reuse index.PostingsEnum, flags int) (index.PostingsEnum, error) {
	//TODO implement me
	panic("implement me")
}

func (t *termsEnum) Impacts(flags int) (index.ImpactsEnum, error) {
	//TODO implement me
	panic("implement me")
}

type BytesRefFSTEnum interface {
}
