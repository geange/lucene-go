package index

import (
	"fmt"
	"github.com/geange/lucene-go/core/tokenattributes"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util"
	"sort"
	"strings"
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

func NewFreqProxTermsWriterPerField(invertState *FieldInvertState, termsHash TermsHash,
	fieldInfo *types.FieldInfo, nextPerField TermsHashPerField) *FreqProxTermsWriterPerField {

	streamCount := 1
	if fieldInfo.GetIndexOptions() >= types.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS {
		streamCount = 2
	}

	indexOptions := fieldInfo.GetIndexOptions()
	termBytePool := termsHash.GetTermBytePool()

	perfield := &FreqProxTermsWriterPerField{
		TermsHashPerFieldDefault: nil,
		freqProxPostingsArray:    nil,
		fieldState:               invertState,
		fieldInfo:                fieldInfo,
		hasFreq:                  indexOptions >= types.INDEX_OPTIONS_DOCS_AND_FREQS,
		hasProx:                  indexOptions >= types.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS,
		hasOffsets:               indexOptions >= types.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS,
		payloadAttribute:         nil,
		offsetAttribute:          nil,
		termFreqAtt:              nil,
		sawPayloads:              false,
	}

	perfield.TermsHashPerFieldDefault = NewTermsHashPerFieldDefault(streamCount,
		termsHash.GetIntPool(), termsHash.GetBytePool(), termsHash.GetTermBytePool(),
		nextPerField, fieldInfo.Name(), indexOptions, perfield)

	byteStarts := NewPostingsBytesStartArray(perfield)
	perfield.bytesHash = util.NewBytesRefHashV1(termBytePool, HASH_INIT_SIZE, byteStarts)

	return perfield
}

func (f *FreqProxTermsWriterPerField) Finish() error {
	err := f.TermsHashPerFieldDefault.Finish()
	if err != nil {
		return err
	}
	if f.sawPayloads {
		return f.fieldInfo.SetStorePayloads()
	}
	return nil
}

func (f *FreqProxTermsWriterPerField) Start(field types.IndexableField, first bool) bool {
	f.TermsHashPerFieldDefault.Start(field, first)
	f.termFreqAtt = f.fieldState.termFreqAttribute
	f.payloadAttribute = f.fieldState.payloadAttribute
	f.offsetAttribute = f.fieldState.offsetAttribute
	return true
}

func (f *FreqProxTermsWriterPerField) writeProx(termID, proxCode int) {
	if f.payloadAttribute == nil {
		f.writeVInt(1, proxCode<<1)
	} else {
		payload := f.payloadAttribute.GetPayload()
		if payload != nil && len(payload) > 0 {
			f.writeVInt(1, (proxCode<<1)|1)
			f.writeVInt(1, len(payload))
			f.writeBytes(1, payload)
			f.sawPayloads = true
		} else {
			f.writeVInt(1, proxCode<<1)
		}
	}

	//assert postingsArray == freqProxPostingsArray;
	f.freqProxPostingsArray.SetLastPositions(termID, f.fieldState.position)
}

func (f *FreqProxTermsWriterPerField) writeOffsets(termID, offsetAccum int) {
	startOffset := offsetAccum + f.offsetAttribute.StartOffset()
	endOffset := offsetAccum + f.offsetAttribute.EndOffset()
	//assert startOffset - freqProxPostingsArray.lastOffsets[termID] >= 0;
	f.writeVInt(1, startOffset-f.freqProxPostingsArray.lastOffsets[termID])
	f.writeVInt(1, endOffset-startOffset)
	f.freqProxPostingsArray.SetLastOffsets(termID, startOffset)
}

func (f *FreqProxTermsWriterPerField) NewTerm(termID, docID int) error {
	// First time we're seeing this term since the last Flush
	postings := f.freqProxPostingsArray
	postings.SetLastDocIDs(termID, docID)
	if !f.hasFreq {
		//assert postings.termFreqs == null;
		postings.SetLastDocCodes(termID, docID)
		f.fieldState.maxTermFrequency = max(1, f.fieldState.maxTermFrequency)
	} else {
		postings.SetLastDocCodes(termID, docID<<1)
		termFreq, err := f.getTermFreq()
		if err != nil {
			return err
		}
		postings.SetTermFreqs(termID, termFreq)
		if f.hasProx {
			f.writeProx(termID, f.fieldState.position)
			if f.hasOffsets {
				f.writeOffsets(termID, f.fieldState.offset)
			}
		} else {
			//assert !hasOffsets;
		}
		f.fieldState.maxTermFrequency = max(postings.termFreqs[termID], f.fieldState.maxTermFrequency)
	}
	f.fieldState.uniqueTermCount++
	return nil
}

func (f *FreqProxTermsWriterPerField) AddTerm(termID, docID int) error {
	postings := f.freqProxPostingsArray

	if !f.hasFreq {
		if f.termFreqAtt.GetTermFrequency() != 1 {
			return fmt.Errorf(
				"field %s: must index term freq while using custom TermFrequencyAttribute",
				f.getFieldName())
		}
		if docID != postings.lastDocIDs[termID] {
			// New document; now encode docCode for previous doc:
			//assert docID > postings.lastDocIDs[termID];
			f.writeVInt(0, postings.lastDocCodes[termID])
			postings.SetLastDocCodes(termID, docID-postings.lastDocIDs[termID])
			postings.SetLastDocIDs(termID, docID)
			f.fieldState.uniqueTermCount++
		}
	} else if docID != postings.lastDocIDs[termID] {
		// assert docID > postings.lastDocIDs[termID]:"id: "+docID + " postings ID: "+ postings.lastDocIDs[termID] + " termID: "+termID;
		// Term not yet seen in the current doc but previously
		// seen in other doc(s) since the last Flush

		// Now that we know doc freq for previous doc,
		// write it & lastDocCode
		if 1 == postings.termFreqs[termID] {
			f.writeVInt(0, postings.lastDocCodes[termID]|1)
		} else {
			f.writeVInt(0, postings.lastDocCodes[termID])
			f.writeVInt(0, postings.termFreqs[termID])
		}

		// Init freq for the current document
		termFreq, err := f.getTermFreq()
		if err != nil {
			return err
		}
		postings.SetTermFreqs(termID, termFreq)
		f.fieldState.maxTermFrequency = max(postings.termFreqs[termID], f.fieldState.maxTermFrequency)
		postings.lastDocCodes[termID] = (docID - postings.lastDocIDs[termID]) << 1
		postings.lastDocIDs[termID] = docID
		if f.hasProx {
			f.writeProx(termID, f.fieldState.position)
			if f.hasOffsets {
				postings.lastOffsets[termID] = 0
				f.writeOffsets(termID, f.fieldState.offset)
			}
		} else {
			//assert !hasOffsets;
		}
		f.fieldState.uniqueTermCount++
	} else {
		termFreq, err := f.getTermFreq()
		if err != nil {
			return err
		}

		postings.SetTermFreqs(termID, postings.termFreqs[termID]+termFreq)
		f.fieldState.maxTermFrequency = max(f.fieldState.maxTermFrequency, postings.termFreqs[termID])
		if f.hasProx {
			f.writeProx(termID, f.fieldState.position-postings.lastPositions[termID])
			if f.hasOffsets {
				f.writeOffsets(termID, f.fieldState.offset)
			}
		}
	}
	return nil
}

func (f *FreqProxTermsWriterPerField) NewPostingsArray() {
	f.freqProxPostingsArray = f.postingsArray.(*FreqProxPostingsArray)
}

func (f *FreqProxTermsWriterPerField) CreatePostingsArray(size int) ParallelPostingsArray {
	hasFreq := f.indexOptions >= types.INDEX_OPTIONS_DOCS_AND_FREQS
	hasProx := f.indexOptions >= types.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS
	hasOffsets := f.indexOptions >= types.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS
	return NewFreqProxPostingsArray(hasFreq, hasProx, hasOffsets)
}

func (f *FreqProxTermsWriterPerField) getTermFreq() (int, error) {
	freq := f.termFreqAtt.GetTermFrequency()
	if freq != 1 {
		if f.hasProx {
			return 0, fmt.Errorf(
				"field %s: cannot index positions while using custom TermFrequencyAttribute",
				f.getFieldName())
		}
	}
	return freq, nil
}

func (f *FreqProxTermsWriterPerField) getFieldName() string {
	return f.fieldName
}

func SortFreqProxTermsWriterPerField(fields []*FreqProxTermsWriterPerField) {
	sort.Sort(FreqProxTermsWriterPerFields(fields))
}

var _ sort.Interface = TermVectorsConsumerPerFields{}

type FreqProxTermsWriterPerFields []*FreqProxTermsWriterPerField

func (p FreqProxTermsWriterPerFields) Len() int {
	return len(p)
}

func (p FreqProxTermsWriterPerFields) Less(i, j int) bool {
	return strings.Compare(p[i].fieldName, p[j].fieldName) < 0
}

func (p FreqProxTermsWriterPerFields) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
