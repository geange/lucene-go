package simpletext

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/geange/gods-generic/maps/treemap"
	coreIndex "github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/types"
)

var _ index.Terms = &SimpleTVTerms{}

type SimpleTVTerms struct {
	*coreIndex.BaseTerms

	terms        *treemap.Map[[]byte, *SimpleTVPostings]
	hasOffsets   bool
	hasPositions bool
	hasPayloads  bool
}

func NewSimpleTVTerms(hasOffsets, hasPositions, hasPayloads bool) *SimpleTVTerms {
	terms := &SimpleTVTerms{
		terms:        treemap.NewWith[[]byte, *SimpleTVPostings](bytes.Compare),
		hasOffsets:   hasOffsets,
		hasPositions: hasPositions,
		hasPayloads:  hasPayloads,
	}
	terms.BaseTerms = coreIndex.NewTerms(terms)
	return terms
}

func (s *SimpleTVTerms) Iterator() (index.TermsEnum, error) {
	return NewSimpleTVTermsEnum(s.terms), nil
}

func (s *SimpleTVTerms) Size() (int, error) {
	return s.terms.Size(), nil
}

func (s *SimpleTVTerms) GetSumTotalTermFreq() (int64, error) {
	ttf := int64(0)
	iterator, err := s.Iterator()
	if err != nil {
		return 0, err
	}

	for {
		next, err := iterator.Next(nil)
		if err != nil {
			if err == io.EOF {
				break
			}
			return 0, err
		}

		if next == nil {
			break
		}

		freq, err := iterator.TotalTermFreq()
		if err != nil {
			return 0, err
		}
		ttf += freq
	}
	return ttf, nil
}

func (s *SimpleTVTerms) GetSumDocFreq() (int64, error) {
	return int64(s.terms.Size()), nil
}

func (s *SimpleTVTerms) GetDocCount() (int, error) {
	return 1, nil
}

func (s *SimpleTVTerms) HasFreqs() bool {
	return true
}

func (s *SimpleTVTerms) HasOffsets() bool {
	return s.hasOffsets
}

func (s *SimpleTVTerms) HasPositions() bool {
	return s.hasPositions
}

func (s *SimpleTVTerms) HasPayloads() bool {
	return s.hasPayloads
}

type SimpleTVPostings struct {
	freq         int
	positions    []int
	startOffsets []int
	endOffsets   []int
	payloads     [][]byte
}

func NewSimpleTVPostings() *SimpleTVPostings {
	return &SimpleTVPostings{
		freq:         0,
		positions:    make([]int, 0),
		startOffsets: make([]int, 0),
		endOffsets:   make([]int, 0),
		payloads:     make([][]byte, 0),
	}
}

var _ index.TermsEnum = &SimpleTVTermsEnum{}

func NewSimpleTVTermsEnum(terms *treemap.Map[[]byte, *SimpleTVPostings]) *SimpleTVTermsEnum {
	iterator := terms.Iterator()
	enum := &SimpleTVTermsEnum{
		terms:    terms,
		iterator: &iterator,
	}
	enum.BaseTermsEnum = coreIndex.NewBaseTermsEnum(&coreIndex.BaseTermsEnumConfig{
		SeekCeil: enum.SeekCeil,
	})
	return enum
}

type SimpleTVTermsEnum struct {
	*coreIndex.BaseTermsEnum

	terms *treemap.Map[[]byte, *SimpleTVPostings]

	iterator *treemap.Iterator[[]byte, *SimpleTVPostings]
}

func (s *SimpleTVTermsEnum) Next(context.Context) ([]byte, error) {
	if !s.iterator.Next() {
		return nil, io.EOF
	}
	return s.iterator.Key(), nil
}

func (s *SimpleTVTermsEnum) SeekCeil(ctx context.Context, text []byte) (index.SeekStatus, error) {
	_, _, ok := s.terms.Ceiling(text)
	if ok {
		return index.SEEK_STATUS_FOUND, nil
	}
	return index.SEEK_STATUS_NOT_FOUND, nil
}

func (s *SimpleTVTermsEnum) SeekExactByOrd(ctx context.Context, ord int64) error {
	return errors.New("unsupported operation exception")
}

func (s *SimpleTVTermsEnum) Term() ([]byte, error) {
	return s.iterator.Key(), nil
}

func (s *SimpleTVTermsEnum) Ord() (int64, error) {
	return 0, errors.New("unsupported operation exception")
}

func (s *SimpleTVTermsEnum) DocFreq() (int, error) {
	return 1, nil
}

func (s *SimpleTVTermsEnum) TotalTermFreq() (int64, error) {
	postings := s.iterator.Value()
	return int64(postings.freq), nil
}

func (s *SimpleTVTermsEnum) Postings(reuse index.PostingsEnum, flags int) (index.PostingsEnum, error) {
	if coreIndex.FeatureRequested(flags, coreIndex.POSTINGS_ENUM_POSITIONS) {
		postings := s.iterator.Value()

		if len(postings.positions) > 0 || len(postings.startOffsets) > 0 {
			o := NewSimpleTVPostingsEnum()
			o.Reset(postings.positions, postings.startOffsets, postings.endOffsets, postings.payloads)
			return o, nil
		}
	}

	e := NewSimpleTVDocsEnum()
	freq := 1
	if !coreIndex.FeatureRequested(flags, coreIndex.POSTINGS_ENUM_FREQS) {
		freq = s.iterator.Value().freq
	}
	e.Reset(freq)
	return e, nil

}

func (s *SimpleTVTermsEnum) Impacts(flags int) (index.ImpactsEnum, error) {
	enum, err := s.Postings(nil, coreIndex.POSTINGS_ENUM_FREQS)
	if err != nil {
		return nil, err
	}
	return coreIndex.NewSlowImpactsEnum(enum), nil
}

var _ index.Fields = &SimpleTVFields{}

func NewSimpleTVFields(fields *treemap.Map[string, index.Terms]) *SimpleTVFields {
	return &SimpleTVFields{fields: fields}
}

type SimpleTVFields struct {
	fields *treemap.Map[string, index.Terms]
}

func (s *SimpleTVFields) Names() []string {
	return s.fields.Keys()
}

func (s *SimpleTVFields) Terms(field string) (index.Terms, error) {
	obj, ok := s.fields.Get(field)
	if !ok {
		return nil, errors.New("not found")
	}
	return obj.(index.Terms), nil
}

func (s *SimpleTVFields) Size() int {
	return s.fields.Size()
}

var _ index.PostingsEnum = &SimpleTVPostingsEnum{}

func NewSimpleTVPostingsEnum() *SimpleTVPostingsEnum {
	enum := &SimpleTVPostingsEnum{
		didNext:      false,
		doc:          -1,
		nextPos:      0,
		positions:    nil,
		payloads:     nil,
		startOffsets: nil,
		endOffsets:   nil,
	}
	return enum
}

type SimpleTVPostingsEnum struct {
	didNext      bool
	doc          int
	nextPos      int
	positions    []int
	payloads     [][]byte
	startOffsets []int
	endOffsets   []int
}

func (s *SimpleTVPostingsEnum) Reset(positions, startOffsets, endOffsets []int, payloads [][]byte) {
	s.positions = positions
	s.startOffsets = startOffsets
	s.endOffsets = endOffsets
	s.payloads = payloads
	s.doc = -1
	s.didNext = false
	s.nextPos = 0
}

func (s *SimpleTVPostingsEnum) DocID() int {
	return s.doc
}

func (s *SimpleTVPostingsEnum) NextDoc() (int, error) {
	if !s.didNext {
		s.didNext = true
		s.doc = 0
		return s.doc, nil
	}
	return -1, io.EOF
}

func (s *SimpleTVPostingsEnum) Advance(target int) (int, error) {
	return s.SlowAdvance(target)
}

func (s *SimpleTVPostingsEnum) SlowAdvance(target int) (int, error) {
	return types.SlowAdvance(s, target)
}

func (s *SimpleTVPostingsEnum) Cost() int64 {
	return 1
}

func (s *SimpleTVPostingsEnum) Freq() (int, error) {
	if len(s.positions) > 0 {
		return len(s.positions), nil
	}
	return len(s.startOffsets), nil
}

func (s *SimpleTVPostingsEnum) NextPosition() (int, error) {
	if len(s.positions) > 0 {
		pos := s.nextPos
		s.nextPos++
		return s.positions[pos], nil
	}
	s.nextPos++
	return -1, nil
}

func (s *SimpleTVPostingsEnum) StartOffset() (int, error) {
	if len(s.startOffsets) > 0 {
		return s.startOffsets[s.nextPos-1], nil
	}
	return -1, nil
}

func (s *SimpleTVPostingsEnum) EndOffset() (int, error) {
	if len(s.endOffsets) > 0 {
		return s.endOffsets[s.nextPos-1], nil
	}
	return -1, nil
}

func (s *SimpleTVPostingsEnum) GetPayload() ([]byte, error) {
	if s.payloads != nil {
		return s.payloads[len(s.payloads)-1], nil
	}
	return nil, nil
}

var _ index.PostingsEnum = &SimpleTVDocsEnum{}

// SimpleTVDocsEnum note: these two enum classes are exactly like the Default impl...
type SimpleTVDocsEnum struct {
	didNext bool
	doc     int
	freq    int
}

func NewSimpleTVDocsEnum() *SimpleTVDocsEnum {
	enum := &SimpleTVDocsEnum{
		doc: -1,
	}
	return enum
}

func (s *SimpleTVDocsEnum) Reset(freq int) {
	s.freq = freq
	s.doc = -1
	s.didNext = false
}

func (s *SimpleTVDocsEnum) DocID() int {
	return s.doc
}

func (s *SimpleTVDocsEnum) NextDoc() (int, error) {
	if !s.didNext {
		s.didNext = true
		s.doc = 0
		return s.doc, nil
	}
	return 0, io.EOF
}

func (s *SimpleTVDocsEnum) Advance(target int) (int, error) {
	return s.SlowAdvance(target)
}

func (s *SimpleTVDocsEnum) SlowAdvance(target int) (int, error) {
	return types.SlowAdvance(s, target)
}

func (s *SimpleTVDocsEnum) Cost() int64 {
	return 1
}

func (s *SimpleTVDocsEnum) Freq() (int, error) {
	return s.freq, nil
}

func (s *SimpleTVDocsEnum) NextPosition() (int, error) {
	return -1, nil
}

func (s *SimpleTVDocsEnum) StartOffset() (int, error) {
	return -1, nil
}

func (s *SimpleTVDocsEnum) EndOffset() (int, error) {
	return -1, nil
}

func (s *SimpleTVDocsEnum) GetPayload() ([]byte, error) {
	return nil, nil
}
