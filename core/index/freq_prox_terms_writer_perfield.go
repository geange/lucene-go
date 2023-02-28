package index

import (
	"github.com/geange/lucene-go/core/tokenattributes"
	"github.com/geange/lucene-go/core/types"
)

var _ TermsHashPerField = &FreqProxTermsWriterPerField{}

type FreqProxTermsWriterPerField struct {
	*TermsHashPerFieldDefault

	freqProxPostingsArray *FreqProxPostingsArray

	fieldState       *FieldInvertState
	fieldInfo        *types.FieldInfo
	hasFreq          bool
	hasProx          bool
	hasOffsets       bool
	payloadAttribute tokenattributes.PayloadAttribute
	offsetAttribute  tokenattributes.OffsetAttribute
	termFreqAtt      tokenattributes.TermFrequencyAttribute

	// Set to true if any token had a payload in the current segment.
	sawPayloads bool
}

func (f *FreqProxTermsWriterPerField) Finish() error {
	//TODO implement me
	panic("implement me")
}

func (f *FreqProxTermsWriterPerField) NewTerm(termID, docID int) error {
	//TODO implement me
	panic("implement me")
}

func (f *FreqProxTermsWriterPerField) AddTerm(termID, docID int) error {
	//TODO implement me
	panic("implement me")
}

func (f *FreqProxTermsWriterPerField) NewPostingsArray() {
	//TODO implement me
	panic("implement me")
}

func (f *FreqProxTermsWriterPerField) CreatePostingsArray(size int) ParallelPostingsArray {
	//TODO implement me
	panic("implement me")
}
