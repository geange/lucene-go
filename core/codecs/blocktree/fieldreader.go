package blocktree

import (
	"github.com/geange/lucene-go/core/document"
	coreIndex "github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/util/fst"
)

var _ index.Terms = &FieldReader{}

type FieldReader struct {
	*coreIndex.BaseTerms

	numTerms         int
	fieldInfo        *document.FieldInfo
	sumTotalTermFreq int64
	sumDocFreq       int64
	docCount         int
	rootBlockFP      int64
	rootCode         []byte
	minTerm          []byte
	maxTerm          []byte
	parent           *TermsReader
	index            *fst.FST
}

func (f *FieldReader) Iterator() (index.TermsEnum, error) {
	return NewSegmentTermsEnum(f)
}

func (f *FieldReader) Size() (int, error) {
	return f.numTerms, nil
}

func (f *FieldReader) GetSumTotalTermFreq() (int64, error) {
	return f.sumTotalTermFreq, nil
}

func (f *FieldReader) GetSumDocFreq() (int64, error) {
	return f.sumDocFreq, nil
}

func (f *FieldReader) GetDocCount() (int, error) {
	return f.docCount, nil
}

func (f *FieldReader) HasFreqs() bool {
	return f.fieldInfo.GetIndexOptions() >= document.INDEX_OPTIONS_DOCS_AND_FREQS
}

func (f *FieldReader) HasOffsets() bool {
	return f.fieldInfo.GetIndexOptions() >= document.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS
}

func (f *FieldReader) HasPositions() bool {
	return f.fieldInfo.GetIndexOptions() >= document.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS
}

func (f *FieldReader) HasPayloads() bool {
	return f.fieldInfo.HasPayloads()
}

func (f *FieldReader) GetMin() ([]byte, error) {
	if f.minTerm != nil {
		return f.minTerm, nil
	}
	return f.BaseTerms.GetMin()
}

func (f *FieldReader) GetMax() ([]byte, error) {
	if f.maxTerm != nil {
		return f.maxTerm, nil
	}
	return f.BaseTerms.GetMax()
}
