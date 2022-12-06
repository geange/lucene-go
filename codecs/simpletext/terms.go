package simpletext

import (
	"bytes"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util/automaton"
	"github.com/geange/lucene-go/core/util/fst"
	"strconv"
)

var _ index.Terms = &textTerms{}

type textTerms struct {
	reader *FieldsReader

	termsStart       int64
	fieldInfo        *types.FieldInfo
	maxDoc           int
	sumTotalTermFreq int64
	sumDocFreq       int64
	docCount         int
	fst              *fst.FST[*fst.Pair[*fst.Pair[int64, int64], *fst.Pair[int64, int64]]]
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
	return terms
}

func (s *textTerms) loadTerms() error {
	posIntOutputs := fst.NewPositiveIntOutputs[int64]()

	outputsOuter := fst.NewPairOutputs[int64, int64](posIntOutputs, posIntOutputs)
	outputsInner := fst.NewPairOutputs[int64, int64](posIntOutputs, posIntOutputs)

	outputs := fst.NewPairOutputs[*fst.Pair[int64, int64], *fst.Pair[int64, int64]](outputsOuter, outputsInner)

	fstCompiler := fst.NewBuilder[*fst.Pair[*fst.Pair[int64, int64], *fst.Pair[int64, int64]]](fst.BYTE1, outputs)

	in := s.reader.in.Clone()
	if err := in.Seek(s.termsStart); err != nil {
		return err
	}

	lastTerm := new(bytes.Buffer)

	lastDocsStart := int64(-1)
	docFreq := int64(0)
	totalTermFreq := int64(0)
	skipPointer := int64(0)
	visitedDocs := make(map[int]struct{})

	for {
		s.scratch.Reset()

		if err := ReadLine(in, s.scratch); err != nil {
			return err
		}

		text := s.scratch.Bytes()

		if bytes.Equal(text, FIELDS_END) || bytes.HasPrefix(text, FIELDS_FIELD) {
			if lastDocsStart != -1 {
				if err := fstCompiler.Add(bytes.Runes(lastTerm.Bytes()),
					fst.NewPair(fst.NewPair(lastDocsStart, skipPointer),
						fst.NewPair(docFreq, totalTermFreq),
					)); err != nil {
					return err
				}
				s.sumTotalTermFreq += totalTermFreq
			}
			break
		} else if bytes.HasSuffix(text, FIELDS_DOC) {
			docFreq++
			s.sumDocFreq++
			totalTermFreq++

			docID, err := strconv.Atoi(string(text[len(FIELDS_DOC):]))
			if err != nil {
				return err
			}
			visitedDocs[docID] = struct{}{}
		} else if bytes.HasPrefix(text, FIELDS_FREQ) {
			value, err := strconv.Atoi(string(text[len(FIELDS_FREQ):]))
			if err != nil {
				return err
			}
			totalTermFreq += int64(value)
		} else if bytes.HasPrefix(text, SKIP_LIST) {
			skipPointer = in.GetFilePointer()
		} else if bytes.HasPrefix(text, FIELDS_TERM) {
			if lastDocsStart != -1 {
				if err := fstCompiler.Add(bytes.Runes(lastTerm.Bytes()),
					fst.NewPair(
						fst.NewPair(lastDocsStart, skipPointer),
						fst.NewPair(docFreq, totalTermFreq),
					)); err != nil {
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
	compiled, err := fstCompiler.Finish()
	if err != nil {
		return err
	}
	s.fst = compiled
	return nil
}

func (s *textTerms) Iterator() (index.TermsEnum, error) {
	if s.fst != nil {
		panic("")
	}
	panic("")
}

func (s *textTerms) Intersect(compiled *automaton.CompiledAutomaton, startTerm []byte) (index.TermsEnum, error) {
	//TODO implement me
	panic("implement me")
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

func (s *textTerms) GetMin() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (s *textTerms) GetMax() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}
