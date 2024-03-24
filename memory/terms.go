package memory

import (
	"bytes"
	"context"
	"io"

	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/tokenattr"
	"github.com/geange/lucene-go/core/util/bytesref"
)

func (r *Index) newTerms(info *info) *Terms {
	terms := &Terms{
		info:  info,
		index: r,
	}
	terms.TermsBase = index.NewTerms(terms)
	return terms
}

type Terms struct {
	*index.TermsBase

	index         *Index
	info          *info
	storeOffsets  bool
	storePayloads bool
}

func (t *Terms) Iterator() (index.TermsEnum, error) {
	return t.index.newTermsEnum(t.info), nil
}

func (t *Terms) Size() (int, error) {
	return t.info.terms.Size(), nil
}

func (t *Terms) GetSumTotalTermFreq() (int64, error) {
	return t.info.sumTotalTermFreq, nil
}

func (t *Terms) GetSumDocFreq() (int64, error) {
	return int64(t.info.terms.Size()), nil
}

func (t *Terms) GetDocCount() (int, error) {
	size, err := t.Size()
	if err != nil {
		return 0, err
	}
	if size > 0 {
		return 1, nil
	}
	return 0, nil
}

func (t *Terms) HasFreqs() bool {
	return true
}

func (t *Terms) HasOffsets() bool {
	return t.storeOffsets
}

func (t *Terms) HasPositions() bool {
	return true
}

func (t *Terms) HasPayloads() bool {
	return t.storePayloads
}

var _ index.TermsEnum = &memTermsEnum{}

type memTermsEnum struct {
	info     *info
	termUpto int
	content  []byte
	atts     *tokenattr.AttributeSource
	index    *Index
}

func (r *Index) newTermsEnum(info *info) *memTermsEnum {
	info.sortTerms()
	return &memTermsEnum{
		info:     info,
		termUpto: -1,
		content:  nil,
		atts:     tokenattr.NewAttributeSource(),
		index:    r,
	}
}

func (m *memTermsEnum) binarySearch(text []byte, low, high int, hash *bytesref.BytesHash, ords []int) int {
	mid := 0
	for low <= high {
		mid = (low + high) >> 1 // mid

		bytesId := ords[mid]
		if bytesId == -1 {
			return -(low + 1)
		}
		bs := hash.Get(bytesId)

		cmp := bytes.Compare(bs, text)
		if cmp == 0 {
			return cmp
		}

		if cmp < 0 {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	return -(low + 1)
}

func (m *memTermsEnum) Next(ctx context.Context) ([]byte, error) {
	m.termUpto++
	if m.termUpto >= m.info.terms.Size() {
		return nil, io.EOF
	}
	m.content = m.info.terms.Get(m.info.sortedTerms[m.termUpto])
	return m.content, nil
}

func (m *memTermsEnum) Attributes() *tokenattr.AttributeSource {
	return m.atts
}

func (m *memTermsEnum) SeekExact(ctx context.Context, text []byte) (bool, error) {
	m.termUpto = m.binarySearch(text, 0, m.info.terms.Size(), m.info.terms, m.info.sortedTerms)
	return m.termUpto >= 0, nil
}

func (m *memTermsEnum) SeekCeil(ctx context.Context, text []byte) (index.SeekStatus, error) {
	m.termUpto = m.binarySearch(text, 0, m.info.terms.Size()-1, m.info.terms, m.info.sortedTerms)
	if m.termUpto < 0 { // not found; choose successor
		m.termUpto = -m.termUpto - 1
		if m.termUpto >= m.info.terms.Size() {
			return index.SEEK_STATUS_END, nil
		}
		m.content = m.info.terms.Get(m.info.sortedTerms[m.termUpto])
		return index.SEEK_STATUS_NOT_FOUND, nil
	}
	return index.SEEK_STATUS_FOUND, nil
}

func (m *memTermsEnum) SeekExactByOrd(ctx context.Context, ord int64) error {
	m.termUpto = int(ord)
	m.content = m.info.terms.Get(m.info.sortedTerms[m.termUpto])
	return nil
}

func (m *memTermsEnum) SeekExactExpert(ctx context.Context, term []byte, state index.TermState) error {
	return m.SeekExactByOrd(ctx, state.(*index.OrdTermState).Ord)
}

func (m *memTermsEnum) Term() ([]byte, error) {
	return m.content, nil
}

func (m *memTermsEnum) Ord() (int64, error) {
	return int64(m.termUpto), nil
}

func (m *memTermsEnum) DocFreq() (int, error) {
	return 1, nil
}

func (m *memTermsEnum) TotalTermFreq() (int64, error) {
	return int64(m.info.sliceArray.freq[m.info.sortedTerms[m.termUpto]]), nil
}

func (m *memTermsEnum) Postings(reuse index.PostingsEnum, flags int) (index.PostingsEnum, error) {
	idx := m.index

	if reuse == nil {
		reuse = newPostingsEnum(idx.intBlockPool, idx.storePayloads)
	}

	if _, ok := reuse.(*memPostingsEnum); !ok {
		reuse = newPostingsEnum(idx.intBlockPool, idx.storePayloads)
	}

	ord := m.info.sortedTerms[m.termUpto]

	array := m.info.sliceArray
	return reuse.(*memPostingsEnum).reset(array.start[ord], array.end[ord], array.freq[ord]), nil
}

func (m *memTermsEnum) Impacts(flags int) (index.ImpactsEnum, error) {
	postings, err := m.Postings(nil, flags)
	if err != nil {
		return nil, err
	}
	return index.NewSlowImpactsEnum(postings), nil
}

func (m *memTermsEnum) TermState() (index.TermState, error) {
	ts := index.NewOrdTermState()
	ts.Ord = int64(m.termUpto)
	return ts, nil
}
