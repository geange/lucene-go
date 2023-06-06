package simpletext

import (
	"bytes"
	"errors"
	"github.com/geange/lucene-go/codecs/utils"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util/fst"
	"io"
	"strconv"
)

var _ index.Terms = &simpleTextTerms{}

type simpleTextTerms struct {
	*index.TermsDefault

	termsStart       int64
	fieldInfo        *types.FieldInfo
	maxDoc           int
	sumTotalTermFreq int64
	sumDocFreq       int64
	docCount         int
	fst              *fst.Fst[*fst.Pair[*fst.Pair[int64, int64], *fst.Pair[int64, int64]]]
	termCount        int

	scratch *bytes.Buffer
}

func (s *SimpleTextFieldsReader) newFieldsReaderTerm(field string, termsStart int64, maxDoc int) (*simpleTextTerms, error) {
	info := s.fieldInfos.FieldInfo(field)
	term := &simpleTextTerms{
		termsStart: termsStart,
		fieldInfo:  info,
		maxDoc:     maxDoc,
		scratch:    new(bytes.Buffer),
	}

	term.TermsDefault = index.NewTermsDefault(&index.TermsDefaultConfig{
		Iterator: term.Iterator,
		Size:     term.Size,
	})

	if err := s.loadTerms(term); err != nil {
		return nil, err
	}
	return term, nil
}

func (s *SimpleTextFieldsReader) loadTerms(term *simpleTextTerms) error {
	posIntOutputs := fst.NewPositiveIntOutputs()
	outputsOuter := fst.NewPairOutputs[int64, int64](posIntOutputs, posIntOutputs)
	outputsInner := fst.NewPairOutputs[int64, int64](posIntOutputs, posIntOutputs)
	outputs := fst.NewPairOutputs[*fst.Pair[int64, int64], *fst.Pair[int64, int64]](outputsOuter, outputsInner)
	fstCompiler := fst.NewBuilder[*fst.Pair[*fst.Pair[int64, int64], *fst.Pair[int64, int64]]](fst.BYTE1, outputs)

	in := s.in.Clone()
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
		if err := utils.ReadLine(in, term.scratch); err != nil {
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
	term.docCount = len(visitedDocs)

	var err error
	term.fst, err = fstCompiler.Finish()
	return err
}

func (f *simpleTextTerms) Iterator() (index.TermsEnum, error) {
	if f.fst != nil {
		return newSimpleTextTermsEnum(f.fst, f.fieldInfo.GetIndexOptions()), nil
	}
	return nil, io.EOF
}

func (f *simpleTextTerms) Size() (int, error) {
	return f.termCount, nil
}

func (f *simpleTextTerms) GetSumTotalTermFreq() (int64, error) {
	return f.sumTotalTermFreq, nil
}

func (f *simpleTextTerms) GetSumDocFreq() (int64, error) {
	return f.sumDocFreq, nil
}

func (f *simpleTextTerms) GetDocCount() (int, error) {
	return f.docCount, nil
}

func (f *simpleTextTerms) HasFreqs() bool {
	return f.fieldInfo.GetIndexOptions() >= types.INDEX_OPTIONS_DOCS_AND_FREQS
}

func (f *simpleTextTerms) HasOffsets() bool {
	return f.fieldInfo.GetIndexOptions() >= types.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS
}

func (f *simpleTextTerms) HasPositions() bool {
	return f.fieldInfo.GetIndexOptions() >= types.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS
}

func (f *simpleTextTerms) HasPayloads() bool {
	return f.fieldInfo.HasPayloads()
}

var _ index.TermsEnum = &simpleTextTermsEnum{}

type EnumPair fst.Pair[*fst.Pair[int64, int64], *fst.Pair[int64, int64]]

type BytesEnum fst.BytesRefFSTEnum[*EnumPair]

type simpleTextTermsEnum struct {
	r *SimpleTextFieldsReader

	*index.BaseTermsEnum

	indexOptions  types.IndexOptions
	docFreq       int
	totalTermFreq int64
	docsStart     int64
	skipPointer   int64
	ended         bool
	fstEnum       *fst.BytesRefFSTEnum[*fst.Pair[*fst.Pair[int64, int64], *fst.Pair[int64, int64]]]
}

func newSimpleTextTermsEnum(fstInstance *fst.Fst[*fst.Pair[*fst.Pair[int64, int64], *fst.Pair[int64, int64]]],
	indexOptions types.IndexOptions) *simpleTextTermsEnum {
	enum := &simpleTextTermsEnum{
		indexOptions: indexOptions,
		fstEnum:      fst.NewBytesRefFSTEnum(fstInstance),
	}
	enum.BaseTermsEnum = index.NewBaseTermsEnum(&index.BaseTermsEnumConfig{SeekCeil: enum.SeekCeil})
	return enum
}

func (t *simpleTextTermsEnum) SeekExact(text []byte) (bool, error) {
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

func (t *simpleTextTermsEnum) Next() ([]byte, error) {
	result, err := t.fstEnum.Next()
	if err != nil {
		return nil, err
	}
	if result != nil {
		pair := result.Output
		// PairOutputs.Pair<Long, Long> pair1 = pair.output1;
		//        PairOutputs.Pair<Long, Long> pair2 = pair.output2;
		//        docsStart = pair1.output1;
		//        skipPointer = pair1.output2;
		//        docFreq = pair2.output1.intValue();
		//        totalTermFreq = pair2.output2;
		//        return result.input;

		pair1, pair2 := pair.Output1, pair.Output2
		t.docsStart = pair1.Output1
		t.skipPointer = pair1.Output2
		t.docFreq = int(pair2.Output1)
		t.totalTermFreq = pair2.Output2
		return result.Input, nil
	} else {
		return nil, nil
	}
}

func (t *simpleTextTermsEnum) SeekCeil(text []byte) (index.SeekStatus, error) {
	result, err := t.fstEnum.SeekCeil(text)
	if err != nil {
		return 0, err
	}

	if result == nil {
		return index.SEEK_STATUS_END, nil
	}

	pair := result.Output
	pair1 := pair.Output1
	pair2 := pair.Output2

	t.docsStart = pair1.Output1
	t.skipPointer = pair1.Output2
	t.docFreq = int(pair2.Output1)
	t.totalTermFreq = pair2.Output2

	if bytes.Equal(result.Input, text) {
		//System.out.println("  match docsStart=" + docsStart);
		return index.SEEK_STATUS_FOUND, nil
	} else {
		//System.out.println("  not match docsStart=" + docsStart);
		return index.SEEK_STATUS_NOT_FOUND, nil
	}
}

func (t *simpleTextTermsEnum) SeekExactByOrd(ord int64) error {
	return errors.New("ErrUnsupportedOperation")
}

func (t *simpleTextTermsEnum) Term() ([]byte, error) {
	return t.fstEnum.Current().Input, nil
}

func (t *simpleTextTermsEnum) Ord() (int64, error) {
	return 0, errors.New("ErrUnsupportedOperation")
}

func (t *simpleTextTermsEnum) DocFreq() (int, error) {
	return t.docFreq, nil
}

func (t *simpleTextTermsEnum) TotalTermFreq() (int64, error) {
	if t.indexOptions == types.INDEX_OPTIONS_DOCS {
		return int64(t.docFreq), nil
	}
	return t.totalTermFreq, nil
}

func (t *simpleTextTermsEnum) Postings(reuse index.PostingsEnum, flags int) (index.PostingsEnum, error) {
	hasPositions := t.indexOptions >= types.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS
	if hasPositions && index.FeatureRequested(flags, index.POSTINGS_ENUM_POSITIONS) {
		var docsAndPositionsEnum *simpleTextPostingsEnum

		enum, ok := reuse.(*simpleTextPostingsEnum)
		if reuse != nil && ok && enum.CanReuse(t.r.in) {
			docsAndPositionsEnum = reuse.(*simpleTextPostingsEnum)
		} else {
			docsAndPositionsEnum = newSimpleTextPostingsEnum()
		}
		return docsAndPositionsEnum.Reset(t.docsStart, t.indexOptions, t.docFreq, t.skipPointer), nil
	}

	var docsEnum *simpleTextDocsEnum
	enum, ok := reuse.(*simpleTextDocsEnum)
	if reuse != nil && ok && enum.CanReuse(t.r.in) {
		docsEnum = reuse.(*simpleTextDocsEnum)
	} else {
		docsEnum = t.r.newSimpleTextDocsEnum()
	}
	return docsEnum.Reset(t.docsStart, t.indexOptions == types.INDEX_OPTIONS_DOCS, t.docFreq, t.skipPointer)
}

func (t *simpleTextTermsEnum) Impacts(flags int) (index.ImpactsEnum, error) {
	postings, err := t.Postings(nil, flags)
	if err != nil {
		return nil, err
	}

	if t.docFreq <= BLOCK_SIZE {
		// no skip data
		return index.NewSlowImpactsEnum(postings), nil
	}

	return postings.(index.ImpactsEnum), nil
}

var _ index.ImpactsEnum = &simpleTextPostingsEnum{}

type simpleTextPostingsEnum struct {
	inStart store.IndexInput
	in      store.IndexInput
	docID   int
	tf      int

	pos           int
	payload       []byte
	nextDocStart  int64
	readOffsets   bool
	readPositions bool

	startOffset int
	endOffset   int
	cost        int
	skipReader  *SimpleTextSkipReader
	nextSkipDoc int
	seekTo      int
}

func newSimpleTextPostingsEnum() *simpleTextPostingsEnum {
	panic("")
}

func (s *simpleTextPostingsEnum) CanReuse(in store.IndexInput) bool {
	return s.in == in
}

func (s *simpleTextPostingsEnum) DocID() int {
	//TODO implement me
	panic("implement me")
}

func (s *simpleTextPostingsEnum) NextDoc() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *simpleTextPostingsEnum) Advance(target int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *simpleTextPostingsEnum) SlowAdvance(target int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *simpleTextPostingsEnum) Cost() int64 {
	//TODO implement me
	panic("implement me")
}

func (s *simpleTextPostingsEnum) Freq() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *simpleTextPostingsEnum) NextPosition() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *simpleTextPostingsEnum) StartOffset() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *simpleTextPostingsEnum) EndOffset() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *simpleTextPostingsEnum) GetPayload() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (s *simpleTextPostingsEnum) AdvanceShallow(target int) error {
	//TODO implement me
	panic("implement me")
}

func (s *simpleTextPostingsEnum) GetImpacts() (index.Impacts, error) {
	//TODO implement me
	panic("implement me")
}

func (s *simpleTextPostingsEnum) Reset(fp int64, indexOptions types.IndexOptions, docFreq int, skipPointer int64) index.PostingsEnum {

	s.nextDocStart = fp
	s.docID = -1
	s.readPositions = indexOptions >= (types.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS)
	s.readOffsets = indexOptions >= (types.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS)
	if !s.readOffsets {
		s.startOffset = -1
		s.endOffset = -1
	}
	s.cost = docFreq
	s.skipReader.Reset(skipPointer, docFreq)
	s.nextSkipDoc = 0
	s.seekTo = -1
	return s
}

var _ index.ImpactsEnum = &simpleTextDocsEnum{}

type simpleTextDocsEnum struct {
	inStart     store.IndexInput
	in          store.IndexInput
	omitTF      bool
	docID       int
	tf          int
	cost        int64
	skipReader  *SimpleTextSkipReader
	nextSkipDoc int
	seekTo      int64
}

func (s *SimpleTextFieldsReader) newSimpleTextDocsEnum() *simpleTextDocsEnum {
	return &simpleTextDocsEnum{
		inStart:     s.in,
		in:          s.in.Clone(),
		omitTF:      false,
		docID:       -1,
		tf:          0,
		cost:        0,
		skipReader:  nil,
		nextSkipDoc: 0,
		seekTo:      -1,
	}
}

func (s *simpleTextDocsEnum) DocID() int {
	return s.docID
}

func (s *simpleTextDocsEnum) NextDoc() (int, error) {
	return s.Advance(s.docID + 1)
}

func (s *simpleTextDocsEnum) Advance(target int) (int, error) {
	err := s.AdvanceShallow(target)
	if err != nil {
		return 0, err
	}
	return s.advanceTarget(target)
}

func (s *simpleTextDocsEnum) SlowAdvance(target int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *simpleTextDocsEnum) Cost() int64 {
	return s.cost
}

func (s *simpleTextDocsEnum) Freq() (int, error) {
	return s.tf, nil
}

func (s *simpleTextDocsEnum) NextPosition() (int, error) {
	return -1, nil
}

func (s *simpleTextDocsEnum) StartOffset() (int, error) {
	return -1, nil
}

func (s *simpleTextDocsEnum) EndOffset() (int, error) {
	return -1, nil
}

func (s *simpleTextDocsEnum) GetPayload() ([]byte, error) {
	return nil, nil
}

func (s *simpleTextDocsEnum) AdvanceShallow(target int) error {
	if target > s.nextSkipDoc {
		_, err := s.skipReader.SkipTo(target)
		if err != nil {
			return err
		}
		if s.skipReader.getNextSkipDoc() != index.NO_MORE_DOCS {
			s.seekTo = s.skipReader.getNextSkipDocFP()
		}
		s.nextSkipDoc = s.skipReader.getNextSkipDoc()
	}
	return nil
}

func (s *simpleTextDocsEnum) GetImpacts() (index.Impacts, error) {
	err := s.AdvanceShallow(s.docID)
	if err != nil {
		return nil, err
	}
	return s.skipReader.GetImpacts(), nil
}

func (s *simpleTextDocsEnum) CanReuse(in store.IndexInput) bool {
	return in == s.inStart
}

func (s *simpleTextDocsEnum) Reset(fp int64, omitTF bool, docFreq int, skipPointer int64) (index.PostingsEnum, error) {
	_, err := s.in.Seek(fp, 0)
	if err != nil {
		return nil, err
	}
	s.omitTF = omitTF
	s.docID = -1
	s.tf = 1
	s.cost = int64(docFreq)
	err = s.skipReader.reset(skipPointer, docFreq)
	if err != nil {
		return nil, err
	}
	s.nextSkipDoc = 0
	s.seekTo = -1
	return s, nil
}

func (s *simpleTextDocsEnum) readDoc() (int, error) {
	if s.docID == index.NO_MORE_DOCS {
		return s.docID, nil
	}
	first := true
	termFreq := 0

	scratch := new(bytes.Buffer)

	for {
		lineStart := s.in.GetFilePointer()
		err := utils.ReadLine(s.in, scratch)
		if err != nil {
			return 0, err
		}
		if bytes.HasPrefix(scratch.Bytes(), FIELDS_DOC) {
			if !first {
				_, err := s.in.Seek(lineStart, 0)
				if err != nil {
					return 0, err
				}
				if !s.omitTF {
					s.tf = termFreq
				}
				return s.docID, nil
			}
			scratch.Next(len(FIELDS_DOC))
			s.docID, err = strconv.Atoi(scratch.String())
			if err != nil {
				return 0, err
			}
			termFreq = 0
			first = false
		} else if bytes.HasPrefix(scratch.Bytes(), FIELDS_FREQ) {
			scratch.Next(len(FIELDS_FREQ))
			termFreq, err = strconv.Atoi(scratch.String())
			if err != nil {
				return 0, err
			}
		} else if bytes.HasPrefix(scratch.Bytes(), FIELDS_POS) {
			// skip termFreq++;
		} else if bytes.HasPrefix(scratch.Bytes(), FIELDS_START_OFFSET) {
			// skip
		} else if bytes.HasPrefix(scratch.Bytes(), FIELDS_END_OFFSET) {
			// skip
		} else if bytes.HasPrefix(scratch.Bytes(), FIELDS_PAYLOAD) {
			// skip
		} else {
			if !first {
				_, err := s.in.Seek(lineStart, 0)
				if err != nil {
					return 0, err
				}
				if !s.omitTF {
					s.tf = termFreq
				}
				return s.docID, nil
			}
		}
	}
}

func (s *simpleTextDocsEnum) advanceTarget(target int) (int, error) {
	if s.seekTo > 0 {
		_, err := s.in.Seek(s.seekTo, 0)
		if err != nil {
			return 0, err
		}
		s.seekTo = -1
	}

	doc, err := s.readDoc()
	if err != nil {
		return 0, err
	}

	for doc >= target {
		break
	}
	return doc, nil
}
