package index

import (
	"bytes"
	"context"
	"errors"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/util/bytesref"
	"github.com/geange/lucene-go/core/util/ints"
	"golang.org/x/exp/maps"
	"io"
	"slices"

	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/util/automaton"
)

var _ index.Fields = &FreqProxFields{}

// FreqProxFields
// Implements limited (iterators only, no stats) Fields interface over the in-RAM
// buffered fields/terms/postings, to flush postings through the PostingsFormat.
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
	return maps.Keys(f.fields)
}

func (f *FreqProxFields) Terms(field string) (index.Terms, error) {
	perField, ok := f.fields[field]
	if !ok {
		return nil, nil
	}
	return NewFreqProxTerms(perField), nil
}

func (f *FreqProxFields) Size() int {
	return len(f.fields)
}

var _ index.Terms = &FreqProxTerms{}

type FreqProxTerms struct {
	terms *FreqProxTermsWriterPerField
}

func NewFreqProxTerms(terms *FreqProxTermsWriterPerField) *FreqProxTerms {
	return &FreqProxTerms{terms: terms}
}

func (f *FreqProxTerms) Iterator() (index.TermsEnum, error) {
	termsEnum := NewFreqProxTermsEnum(f.terms)
	termsEnum.reset()
	return termsEnum, nil
}

func (f *FreqProxTerms) Intersect(compiled *automaton.CompiledAutomaton, startTerm []byte) (index.TermsEnum, error) {
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
		document.INDEX_OPTIONS_DOCS_AND_FREQS
}

func (f *FreqProxTerms) HasOffsets() bool {
	return f.terms.indexOptions >=
		document.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS
}

func (f *FreqProxTerms) HasPositions() bool {
	return f.terms.indexOptions >=
		document.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS
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

var _ index.TermsEnum = &FreqProxTermsEnum{}

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

func (f *FreqProxTermsEnum) Next(context.Context) ([]byte, error) {
	f.ord++
	if f.ord >= f.numTerms {
		return nil, io.EOF
	}
	textStart := f.postingsArray.textStarts[f.sortedTermIDs[f.ord]]
	scratch, err := f.terms.bytePool.GetAddress(textStart)
	if err != nil {
		return nil, err
	}
	f.scratch = scratch
	return f.scratch, nil
}

func (f *FreqProxTermsEnum) SeekCeil(ctx context.Context, text []byte) (index.SeekStatus, error) {
	// binary search:
	lo := 0
	hi := f.numTerms - 1
	for hi >= lo {
		mid := (lo + hi) >> 1
		textStart := f.postingsArray.textStarts[f.sortedTermIDs[mid]]
		scratch, err := f.terms.bytePool.GetAddress(textStart)
		if err != nil {
			return 0, err
		}
		f.scratch = scratch
		cmp := bytes.Compare(f.scratch, text)
		if cmp < 0 {
			lo = mid + 1
		} else if cmp > 0 {
			hi = mid - 1
		} else {
			// found:
			f.ord = mid
			//assert term().compareTo(text) == 0;
			return index.SEEK_STATUS_FOUND, nil
		}
	}

	// not found:
	f.ord = lo
	if f.ord >= f.numTerms {
		return index.SEEK_STATUS_END, nil
	} else {
		textStart := f.postingsArray.textStarts[f.sortedTermIDs[f.ord]]
		f.scratch, _ = f.terms.bytePool.GetAddress(textStart)
		//assert term().compareTo(text) > 0;
		return index.SEEK_STATUS_NOT_FOUND, nil
	}
}

func (f *FreqProxTermsEnum) SeekExactByOrd(ctx context.Context, ord int64) error {
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

func (f *FreqProxTermsEnum) Postings(reuse index.PostingsEnum, flags int) (index.PostingsEnum, error) {
	if flags&POSTINGS_ENUM_POSITIONS == POSTINGS_ENUM_POSITIONS {
		if !f.terms.hasProx {
			return nil, errors.New("did not index positions")
		}

		if !f.terms.hasOffsets && flags&POSTINGS_ENUM_OFFSETS == POSTINGS_ENUM_OFFSETS {
			return nil, errors.New("did not index offsets")
		}

		if posEnum, ok := reuse.(*FreqProxPostingsEnum); ok {
			if posEnum.postingsArray != f.postingsArray {
				posEnum = f.terms.newFreqProxPostingsEnum(f.terms, f.postingsArray)
			}
			if err := posEnum.reset(f.sortedTermIDs[f.ord]); err != nil {
				return nil, err
			}
			return posEnum, nil
		}
		posEnum := f.terms.newFreqProxPostingsEnum(f.terms, f.postingsArray)
		if err := posEnum.reset(f.sortedTermIDs[f.ord]); err != nil {
			return nil, err
		}
		return posEnum, nil
	}

	if !f.terms.hasFreq && flags&POSTINGS_ENUM_FREQS == POSTINGS_ENUM_FREQS {
		return nil, errors.New("did not index freq")
	}

	//  if (reuse instanceof FreqProxDocsEnum) {
	//        docsEnum = (FreqProxDocsEnum) reuse;
	//        if (docsEnum.postingsArray != postingsArray) {
	//          docsEnum = new FreqProxDocsEnum(terms, postingsArray);
	//        }
	//      } else {
	//        docsEnum = new FreqProxDocsEnum(terms, postingsArray);
	//      }
	//      docsEnum.reset(sortedTermIDs[ord]);
	//      return docsEnum;

	panic("implement me")
}

func (f *FreqProxTermsEnum) Impacts(flags int) (index.ImpactsEnum, error) {
	//TODO implement me
	panic("implement me")
}

var _ index.PostingsEnum = &FreqProxPostingsEnum{}

type FreqProxPostingsEnum struct {
	pf *FreqProxTermsWriterPerField

	terms         *FreqProxTermsWriterPerField
	postingsArray *FreqProxPostingsArray
	reader        *ByteSliceReader
	posReader     *ByteSliceReader
	readOffsets   bool
	docID         int
	freq          int
	pos           int
	startOffset   int
	endOffset     int
	posLeft       int
	termID        int
	ended         bool
	hasPayload    bool
	payload       []byte
}

func (f *FreqProxTermsWriterPerField) newFreqProxPostingsEnum(terms *FreqProxTermsWriterPerField, postingsArray *FreqProxPostingsArray) *FreqProxPostingsEnum {
	return &FreqProxPostingsEnum{
		pf:            f,
		docID:         -1,
		terms:         terms,
		postingsArray: postingsArray,
		readOffsets:   terms.hasOffsets,
		reader:        NewByteSliceReader(),
		posReader:     NewByteSliceReader(),
	}
}

func (f *FreqProxPostingsEnum) reset(termID int) error {
	f.termID = termID
	if err := f.terms.initReader(f.reader, termID, 0); err != nil {
		return err
	}
	if err := f.terms.initReader(f.posReader, termID, 1); err != nil {
		return err
	}
	f.ended = false
	f.docID = -1
	f.posLeft = 0
	return nil
}

func (f *FreqProxPostingsEnum) initReader(reader *ByteSliceReader, termID, stream int) error {
	streamStartOffset := f.postingsArray.addressOffset[termID]
	streamAddressBuffer := f.pf.intPool.Get(streamStartOffset >> ints.INT_BLOCK_SHIFT)
	offsetInAddressBuffer := streamStartOffset & ints.INT_BLOCK_MASK
	return reader.init(f.pf.bytePool,
		f.postingsArray.byteStarts[termID]+stream*bytesref.FIRST_LEVEL_SIZE,
		streamAddressBuffer[offsetInAddressBuffer+stream])
}

func (f *FreqProxPostingsEnum) DocID() int {
	return f.docID
}

func (f *FreqProxPostingsEnum) NextDoc() (int, error) {
	if f.docID == -1 {
		f.docID = 0
	}

	for f.posLeft != 0 {
		_, err := f.NextPosition()
		if err != nil {
			return 0, err
		}
	}

	if f.reader.EOF() {
		if f.ended {
			return 0, io.EOF
		} else {
			f.ended = true
			f.docID = f.postingsArray.lastDocIDs[f.termID]
			f.freq = f.postingsArray.termFreqs[f.termID]
		}
	} else {
		code, err := f.reader.ReadUvarint(context.Background())
		if err != nil {
			return 0, err
		}
		f.docID += int(code >> 1)
		if code&1 != 0 {
			f.freq = 1
		} else {
			freq, err := f.reader.ReadUvarint(context.Background())
			if err != nil {
				return 0, err
			}
			f.freq = int(freq)
		}
	}

	f.posLeft = f.freq
	f.pos = 0
	f.startOffset = 0
	return f.docID, nil
}

func (f *FreqProxPostingsEnum) Advance(ctx context.Context, target int) (int, error) {
	return 0, errors.New("implement me")
}

func (f *FreqProxPostingsEnum) SlowAdvance(ctx context.Context, target int) (int, error) {
	return 0, errors.New("implement me")
}

func (f *FreqProxPostingsEnum) Cost() int64 {
	return -1
}

func (f *FreqProxPostingsEnum) Freq() (int, error) {
	return f.freq, nil
}

func (f *FreqProxPostingsEnum) NextPosition() (int, error) {
	f.posLeft--
	code, err := f.posReader.ReadUvarint(context.Background())
	if err != nil {
		return 0, err
	}

	f.hasPayload = false
	f.pos += int(code >> 1)

	if code&1 != 0 {
		f.hasPayload = true
		size, err := f.posReader.ReadUvarint(context.Background())
		if err != nil {
			return 0, err
		}
		slices.Grow(f.payload, int(size))
		if _, err := f.posReader.Read(f.payload); err != nil {
			return 0, err
		}
	}

	if f.readOffsets {
		numStart, err := f.posReader.ReadUvarint(context.Background())
		if err != nil {
			return 0, err
		}
		f.startOffset += int(numStart)

		numEnd, err := f.posReader.ReadUvarint(context.Background())
		if err != nil {
			return 0, err
		}
		f.endOffset = f.startOffset + int(numEnd)
	}

	return f.pos, nil
}

func (f *FreqProxPostingsEnum) StartOffset() (int, error) {
	if !f.readOffsets {
		return -1, errors.New("offsets were not indexed")
	}
	return f.startOffset, nil
}

func (f *FreqProxPostingsEnum) EndOffset() (int, error) {
	if !f.readOffsets {
		return -1, errors.New("offsets were not indexed")
	}
	return f.endOffset, nil
}

func (f *FreqProxPostingsEnum) GetPayload() ([]byte, error) {
	if f.hasPayload {
		return f.payload, nil
	}
	return nil, io.EOF
}
