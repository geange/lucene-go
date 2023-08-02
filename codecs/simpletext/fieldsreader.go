package simpletext

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strconv"

	"github.com/geange/gods-generic/maps/treemap"
	"github.com/geange/lucene-go/codecs/utils"
	"github.com/geange/lucene-go/core/document"
	coreIndex "github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util/fst"
)

var (
	_ index.FieldsProducer = &FieldsReader{}
	_ index.Fields         = &FieldsReader{}
)

type FieldsReader struct {
	fields     *treemap.Map[string, int64]
	in         store.IndexInput
	fieldInfos index.FieldInfos
	maxDoc     int
	termsCache map[string]*simpleTextTerms
}

func (s *FieldsReader) Names() []string {
	return s.fields.Keys()
}

func (s *FieldsReader) Terms(field string) (index.Terms, error) {
	v, ok := s.termsCache[field]
	if !ok {
		fp, ok := s.fields.Get(field)
		if !ok {
			return nil, nil
		}
		terms, err := s.newFieldsReaderTerm(context.Background(), field, fp, s.maxDoc)
		if err != nil {
			return nil, err
		}
		s.termsCache[field] = terms
		return terms, nil
	}
	return v, nil
}

func (s *FieldsReader) Size() int {
	return -1
}

func NewSimpleTextFieldsReader(state *index.SegmentReadState) (*FieldsReader, error) {
	maxDoc, err := state.SegmentInfo.MaxDoc()
	if err != nil {
		return nil, err
	}

	name := getPostingsFileName(state.SegmentInfo.Name(), state.SegmentSuffix)
	input, err := state.Directory.OpenInput(nil, name)
	if err != nil {
		return nil, err
	}

	reader := &FieldsReader{
		fields:     nil,
		in:         input,
		fieldInfos: state.FieldInfos,
		maxDoc:     maxDoc,
		termsCache: make(map[string]*simpleTextTerms),
	}

	fields, err := reader.readFields(reader.in.Clone().(store.IndexInput))
	if err != nil {
		_ = input.Close()
		return nil, err
	}
	reader.fields = fields
	return reader, nil
}

func (s *FieldsReader) readFields(in store.IndexInput) (*treemap.Map[string, int64], error) {
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

func (s *FieldsReader) Close() error {
	return s.in.Close()
}

func (s *FieldsReader) CheckIntegrity() error {
	return nil
}

func (s *FieldsReader) GetMergeInstance() index.FieldsProducer {
	return s
}

var _ index.Terms = &simpleTextTerms{}

type simpleTextTerms struct {
	*coreIndex.BaseTerms

	r                *FieldsReader
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

func (s *FieldsReader) newFieldsReaderTerm(ctx context.Context, field string, termsStart int64, maxDoc int) (*simpleTextTerms, error) {
	info := s.fieldInfos.FieldInfo(field)
	term := &simpleTextTerms{
		r:          s,
		termsStart: termsStart,
		fieldInfo:  info,
		maxDoc:     maxDoc,
		scratch:    new(bytes.Buffer),
	}

	term.BaseTerms = coreIndex.NewTerms(term)

	if err := s.loadTerms(ctx, term); err != nil {
		return nil, err
	}
	return term, nil
}

func (s *FieldsReader) loadTerms(ctx context.Context, term *simpleTextTerms) error {
	fstCompiler, err := fst.NewBuilder(fst.BYTE1, fst.NewPostingOutputManager())
	if err != nil {
		return err
	}

	in := s.in.Clone().(store.IndexInput)
	if _, err := in.Seek(term.termsStart, io.SeekStart); err != nil {
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
				//output := fst.NewPair(fst.NewPair(lastDocsStart, skipPointer), fst.NewPair(docFreq, totalTermFreq))

				output := fst.NewPostingOutput(lastDocsStart, skipPointer, docFreq, totalTermFreq)

				if err := fstCompiler.Add(ctx, bytes.Runes(lastTerm.Bytes()), output); err != nil {
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
				output := fst.NewPostingOutput(lastDocsStart, skipPointer, docFreq, totalTermFreq)
				if err := fstCompiler.Add(ctx, bytes.Runes(lastTerm.Bytes()), output); err != nil {
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

	term.fst, err = fstCompiler.Finish(ctx)
	return err
}

func (f *simpleTextTerms) Iterator() (index.TermsEnum, error) {
	if f.fst != nil {
		return f.r.newSimpleTextTermsEnum(f.fst, f.fieldInfo.GetIndexOptions())
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
	return f.fieldInfo.GetIndexOptions() >= document.INDEX_OPTIONS_DOCS_AND_FREQS
}

func (f *simpleTextTerms) HasOffsets() bool {
	return f.fieldInfo.GetIndexOptions() >= document.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS
}

func (f *simpleTextTerms) HasPositions() bool {
	return f.fieldInfo.GetIndexOptions() >= document.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS
}

func (f *simpleTextTerms) HasPayloads() bool {
	return f.fieldInfo.HasPayloads()
}

var _ index.TermsEnum = &simpleTextTermsEnum{}

type simpleTextTermsEnum struct {
	*coreIndex.BaseTermsEnum

	r             *FieldsReader
	indexOptions  document.IndexOptions
	docFreq       int
	totalTermFreq int64
	docsStart     int64
	skipPointer   int64
	ended         bool
	fstEnum       *fst.Enum[byte]
}

func (s *FieldsReader) newSimpleTextTermsEnum(fstInstance *fst.FST, indexOptions document.IndexOptions) (*simpleTextTermsEnum, error) {
	fstEnum, err := fst.NewEnum[byte](fstInstance)
	if err != nil {
		return nil, err
	}

	enum := &simpleTextTermsEnum{
		indexOptions: indexOptions,
		fstEnum:      fstEnum,
		r:            s,
	}
	enum.BaseTermsEnum = coreIndex.NewBaseTermsEnum(&coreIndex.BaseTermsEnumConfig{SeekCeil: enum.SeekCeil})
	return enum, nil
}

func (t *simpleTextTermsEnum) SeekExact(ctx context.Context, text []byte) (bool, error) {
	result, ok, err := t.fstEnum.SeekExact(ctx, text)
	if err != nil {
		return false, err
	}

	if !ok {
		return false, nil
	}

	if result != nil {
		output := result.GetOutput()

		if posting, ok := output.(*fst.PostingOutput); ok {
			t.docsStart = posting.LastDocsStart
			t.skipPointer = posting.SkipPointer
			t.docFreq = int(posting.DocFreq)
			t.totalTermFreq = posting.TotalTermFreq
		}

		return true, nil
	}
	return false, nil
}

func (t *simpleTextTermsEnum) Next(ctx context.Context) ([]byte, error) {
	result, err := t.fstEnum.Next(ctx)
	if err != nil {
		return nil, err
	}
	if result != nil {
		output := result.GetOutput()

		if posting, ok := output.(*fst.PostingOutput); ok {
			t.docsStart = posting.LastDocsStart
			t.skipPointer = posting.SkipPointer
			t.docFreq = int(posting.DocFreq)
			t.totalTermFreq = posting.TotalTermFreq
		}

		return result.GetInput(), nil
	} else {
		return nil, nil
	}
}

func (t *simpleTextTermsEnum) SeekCeil(ctx context.Context, text []byte) (index.SeekStatus, error) {
	result, found, err := t.fstEnum.SeekCeil(ctx, text)
	if err != nil {
		return 0, err
	}

	if !found {
		return index.SEEK_STATUS_END, nil
	}

	output := result.GetOutput()

	if posting, ok := output.(*fst.PostingOutput); ok {
		t.docsStart = posting.LastDocsStart
		t.skipPointer = posting.SkipPointer
		t.docFreq = int(posting.DocFreq)
		t.totalTermFreq = posting.TotalTermFreq
	}

	if bytes.Equal(result.GetInput(), text) {
		return index.SEEK_STATUS_FOUND, nil
	} else {
		return index.SEEK_STATUS_NOT_FOUND, nil
	}
}

func (t *simpleTextTermsEnum) SeekExactByOrd(ctx context.Context, ord int64) error {
	return errors.New("ErrUnsupportedOperation")
}

func (t *simpleTextTermsEnum) Term() ([]byte, error) {
	return t.fstEnum.Current().GetInput(), nil
}

func (t *simpleTextTermsEnum) Ord() (int64, error) {
	return 0, errors.New("ErrUnsupportedOperation")
}

func (t *simpleTextTermsEnum) DocFreq() (int, error) {
	return t.docFreq, nil
}

func (t *simpleTextTermsEnum) TotalTermFreq() (int64, error) {
	if t.indexOptions == document.INDEX_OPTIONS_DOCS {
		return int64(t.docFreq), nil
	}
	return t.totalTermFreq, nil
}

func (t *simpleTextTermsEnum) Postings(reuse index.PostingsEnum, flags int) (index.PostingsEnum, error) {
	hasPositions := t.indexOptions >= document.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS
	if hasPositions && coreIndex.FeatureRequested(flags, coreIndex.POSTINGS_ENUM_POSITIONS) {
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
	return docsEnum.Reset(t.docsStart, t.indexOptions == document.INDEX_OPTIONS_DOCS, t.docFreq, t.skipPointer)
}

func (t *simpleTextTermsEnum) Impacts(flags int) (index.ImpactsEnum, error) {
	postings, err := t.Postings(nil, flags)
	if err != nil {
		return nil, err
	}

	if t.docFreq <= BLOCK_SIZE {
		// no skip data
		return coreIndex.NewSlowImpactsEnum(postings), nil
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
	skipReader  *SkipReader
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

func (s *simpleTextPostingsEnum) Reset(fp int64, indexOptions document.IndexOptions, docFreq int, skipPointer int64) index.PostingsEnum {

	s.nextDocStart = fp
	s.docID = -1
	s.readPositions = indexOptions >= (document.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS)
	s.readOffsets = indexOptions >= (document.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS)
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
	skipReader  *SkipReader
	nextSkipDoc int
	seekTo      int64
}

func (s *FieldsReader) newSimpleTextDocsEnum() *simpleTextDocsEnum {
	return &simpleTextDocsEnum{
		inStart:     s.in,
		in:          s.in.Clone().(store.IndexInput),
		omitTF:      false,
		docID:       -1,
		tf:          0,
		cost:        0,
		skipReader:  NewSkipReader(s.in.Clone().(store.IndexInput)),
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
	return s.Advance(target)
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
		if s.skipReader.getNextSkipDoc() != types.NO_MORE_DOCS {
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
	_, err := s.in.Seek(fp, io.SeekStart)
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
	if s.docID == types.NO_MORE_DOCS {
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
				_, err := s.in.Seek(lineStart, io.SeekStart)
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
			//assert StringHelper.startsWith(scratch.get(), SimpleTextSkipWriter.SKIP_LIST)
			//|| StringHelper.startsWith(scratch.get(), TERM)
			//|| StringHelper.startsWith(scratch.get(), FIELD)
			//|| StringHelper.startsWith(scratch.get(), END)
			//: "scratch=" + scratch.get().utf8ToString();
			if bytes.HasPrefix(scratch.Bytes(), SKIP_LIST) ||
				bytes.HasPrefix(scratch.Bytes(), FIELDS_TERM) ||
				bytes.HasPrefix(scratch.Bytes(), FIELDS_FIELD) ||
				bytes.HasPrefix(scratch.Bytes(), FIELDS_END) {

				if !first {
					_, err := s.in.Seek(lineStart, io.SeekStart)
					if err != nil {
						return 0, err
					}
					if !s.omitTF {
						s.tf = termFreq
					}
					return s.docID, nil
				}

			}
			s.docID = types.NO_MORE_DOCS
			return s.docID, io.EOF
		}
	}
}

func (s *simpleTextDocsEnum) advanceTarget(target int) (int, error) {
	if s.seekTo > 0 {
		if _, err := s.in.Seek(s.seekTo, io.SeekStart); err != nil {
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
