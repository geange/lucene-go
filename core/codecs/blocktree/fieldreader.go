package blocktree

import (
	"context"
	"io"

	"github.com/geange/lucene-go/core/document"
	coreIndex "github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
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
	index            *fst.FST[[]byte]
}

func NewFieldReader(ctx context.Context, parent *TermsReader, fieldInfo *document.FieldInfo, numTerms int,
	rootCode []byte, sumTotalTermFreq int64, sumDocFreq int64, docCount int,
	indexStartFP int64, metaIn store.IndexInput, indexIn store.IndexInput,
	minTerm []byte, maxTerm []byte) (*FieldReader, error) {

	this := &FieldReader{
		fieldInfo:        fieldInfo,
		parent:           parent,
		numTerms:         numTerms,
		sumTotalTermFreq: sumTotalTermFreq,
		sumDocFreq:       sumDocFreq,
		docCount:         docCount,
		rootCode:         rootCode,
		minTerm:          minTerm,
		maxTerm:          maxTerm,
	}

	n, err := store.NewByteArrayDataInput(rootCode).ReadUvarint(ctx)
	if err != nil {
		return nil, err
	}
	this.rootBlockFP = int64(n >> OUTPUT_FLAGS_NUM_BITS)
	clone := indexIn.Clone().(store.IndexInput)
	if _, err := clone.Seek(indexStartFP, io.SeekStart); err != nil {
		return nil, err
	}
	if metaIn == indexIn { // Only true before Lucene 8.6
		fstIndex, err := fst.NewFstV2(ctx, fst.NewByteSequenceOutputs(), new(fst.OffHeapStore), clone, clone)
		if err != nil {
			return nil, err
		}
		this.index = fstIndex
	} else {
		fstIndex, err := fst.NewFstV2(ctx, fst.NewByteSequenceOutputs(), new(fst.OffHeapStore), metaIn, clone)
		if err != nil {
			return nil, err
		}
		this.index = fstIndex
	}
	return this, nil
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
