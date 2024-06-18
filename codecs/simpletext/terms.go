package simpletext

import (
	"bytes"
	"context"
	index2 "github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"io"
	"strconv"

	"github.com/geange/lucene-go/codecs/utils"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/util/fst"
)

var _ index2.Terms = &textTerms{}

type textTerms struct {
	*index.BaseTerms

	reader           *FieldsReader
	termsStart       int64
	fieldInfo        *document.FieldInfo
	maxDoc           int
	sumTotalTermFreq int64
	sumDocFreq       int64
	docCount         int
	fst              *fst.FST
	termCount        int
	scratch          *bytes.Buffer
}

func (s *FieldsReader) newSimpleTextTerms(field string, termsStart int64, maxDoc int) *textTerms {
	terms := &textTerms{
		reader:           s,
		termsStart:       termsStart,
		fieldInfo:        s.fieldInfos.FieldInfo(field),
		maxDoc:           maxDoc,
		sumTotalTermFreq: 0,
		sumDocFreq:       0,
		docCount:         0,
		fst:              nil,
		termCount:        0,
		scratch:          new(bytes.Buffer),
	}
	terms.BaseTerms = index.NewTerms(terms)
	return terms
}

func (s *textTerms) loadTerms(ctx context.Context) error {
	fstCompiler, err := fst.NewBuilder(fst.BYTE1, fst.NewPostingOutputManager())
	if err != nil {
		return err
	}

	in := s.reader.in.Clone().(store.IndexInput)
	if _, err := in.Seek(s.termsStart, io.SeekStart); err != nil {
		return err
	}

	lastTerm := new(bytes.Buffer)

	lastDocsStart := int64(-1)
	docFreq := int64(0)
	totalTermFreq := int64(0)
	skipPointer := int64(0)
	visitedDocs := make(map[int]struct{})

OUTER:
	for {
		s.scratch.Reset()

		if err := utils.ReadLine(in, s.scratch); err != nil {
			return err
		}

		text := s.scratch.Bytes()

		switch {
		case bytes.Equal(text, FIELDS_END) || bytes.HasPrefix(text, FIELDS_FIELD):
			if lastDocsStart != -1 {
				value := fst.NewPostingOutput(lastDocsStart, skipPointer, docFreq, totalTermFreq)
				if err := fstCompiler.Add(ctx, bytes.Runes(lastTerm.Bytes()), value); err != nil {
					return err
				}
				s.sumTotalTermFreq += totalTermFreq
			}
			break OUTER
		case bytes.HasSuffix(text, FIELDS_DOC):
			docFreq++
			s.sumDocFreq++
			totalTermFreq++

			docID, err := strconv.Atoi(string(text[len(FIELDS_DOC):]))
			if err != nil {
				return err
			}
			visitedDocs[docID] = struct{}{}

		case bytes.HasPrefix(text, FIELDS_FREQ):
			value, err := strconv.Atoi(string(text[len(FIELDS_FREQ):]))
			if err != nil {
				return err
			}
			totalTermFreq += int64(value)

		case bytes.HasPrefix(text, SKIP_LIST):
			skipPointer = in.GetFilePointer()

		case bytes.HasPrefix(text, FIELDS_TERM):
			if lastDocsStart != -1 {
				value := fst.NewPostingOutput(lastDocsStart, skipPointer, docFreq, totalTermFreq)
				if err := fstCompiler.Add(ctx, bytes.Runes(lastTerm.Bytes()), value); err != nil {
					return err
				}
			}
			lastDocsStart = in.GetFilePointer()
			lastTerm.Write(text)
			docFreq = 0
			s.sumTotalTermFreq += totalTermFreq
			totalTermFreq = 0
			s.termCount++
			skipPointer = 0
		}
	}

	s.docCount = len(visitedDocs)
	compiled, err := fstCompiler.Finish(ctx)
	if err != nil {
		return err
	}
	s.fst = compiled
	return nil
}

func (s *textTerms) Iterator() (index2.TermsEnum, error) {
	if s.fst != nil {
		panic("")
	}
	panic("")
}

func (s *textTerms) Size() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *textTerms) GetSumTotalTermFreq() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (s *textTerms) GetSumDocFreq() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (s *textTerms) GetDocCount() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *textTerms) HasFreqs() bool {
	//TODO implement me
	panic("implement me")
}

func (s *textTerms) HasOffsets() bool {
	//TODO implement me
	panic("implement me")
}

func (s *textTerms) HasPositions() bool {
	//TODO implement me
	panic("implement me")
}

func (s *textTerms) HasPayloads() bool {
	//TODO implement me
	panic("implement me")
}
