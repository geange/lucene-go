package index

import (
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/util/attribute"
)

// FieldInvertState
// This class tracks the number and position / offset parameters of terms being added to the index.
// The information collected in this class is also used to calculate the normalization factor for a field.
type FieldInvertState struct {
	indexCreatedVersionMajor int
	name                     string
	indexOptions             document.IndexOptions
	position                 int
	length                   int
	numOverlap               int
	offset                   int
	maxTermFrequency         int
	uniqueTermCount          int
	// we must track these across field instances (multi-valued case)
	lastStartOffset int
	lastPosition    int

	attributeSource   *attribute.Source
	offsetAttribute   attribute.OffsetAttr
	posIncrAttribute  attribute.PositionIncrAttr
	payloadAttribute  attribute.PayloadAttr
	termAttribute     attribute.Term2BytesAttr
	termFreqAttribute attribute.TermFreqAttr
}

// NewFieldInvertState Creates {code FieldInvertState} for the specified field name and values for all fields.
func NewFieldInvertState(indexCreatedVersionMajor int, name string, indexOptions document.IndexOptions, position int, length int, numOverlap int, offset int, maxTermFrequency int, uniqueTermCount int) *FieldInvertState {
	return &FieldInvertState{
		indexCreatedVersionMajor: indexCreatedVersionMajor,
		name:                     name,
		indexOptions:             indexOptions,
		position:                 position,
		length:                   length,
		numOverlap:               numOverlap,
		offset:                   offset,
		maxTermFrequency:         maxTermFrequency,
		uniqueTermCount:          uniqueTermCount,
	}
}

func NewFieldInvertStateV1(indexCreatedVersionMajor int, name string, indexOptions document.IndexOptions) *FieldInvertState {
	return &FieldInvertState{
		indexCreatedVersionMajor: indexCreatedVersionMajor,
		name:                     name,
		indexOptions:             indexOptions,
	}
}

func (f *FieldInvertState) Reset() {
	f.position = -1
	f.length = 0
	f.numOverlap = 0
	f.offset = 0
	f.maxTermFrequency = 0
	f.uniqueTermCount = 0
	f.lastStartOffset = 0
	f.lastPosition = 0
}

func (f *FieldInvertState) SetAttributeSource(attributeSource *attribute.Source) {
	if f.attributeSource != attributeSource {
		f.attributeSource = attributeSource
		f.termAttribute = attributeSource.Term2Bytes()
		f.termFreqAttribute = attributeSource.TermFrequency()
		f.posIncrAttribute = attributeSource.PositionIncrement()
		f.offsetAttribute = attributeSource.Offset()
		f.payloadAttribute = attributeSource.Payload()
	}
}

// GetPosition Get the last processed term position.
// Returns: the position
func (f *FieldInvertState) GetPosition() int {
	return f.position
}

// GetLength Get total number of terms in this field.
// Returns: the length
func (f *FieldInvertState) GetLength() int {
	return f.length
}

// SetLength Set length item.
func (f *FieldInvertState) SetLength(length int) {
	f.length = length
}

// GetNumOverlap Get the number of terms with positionIncrement == 0.
// Returns: the numOverlap
func (f *FieldInvertState) GetNumOverlap() int {
	return f.numOverlap
}

// SetNumOverlap Set number of terms with positionIncrement == 0.
func (f *FieldInvertState) SetNumOverlap(numOverlap int) {
	f.numOverlap = numOverlap
}

// GetOffset Get end offset of the last processed term.
// Returns: the offset
func (f *FieldInvertState) GetOffset() int {
	return f.offset
}

// GetMaxTermFrequency Get the maximum term-frequency encountered for any term in the field. A field
// containing "the quick brown fox jumps over the lazy dog" would have a item of 2, because "the" appears twice.
func (f *FieldInvertState) GetMaxTermFrequency() int {
	return f.maxTermFrequency
}

// GetUniqueTermCount Return the number of unique terms encountered in this field.
func (f *FieldInvertState) GetUniqueTermCount() int {
	return f.uniqueTermCount
}

// GetAttributeSource Returns the AttributeSourceV2 from the TokenStream that provided the indexed tokens for this field.
func (f *FieldInvertState) GetAttributeSource() *attribute.Source {
	return f.attributeSource
}

// GetName Return the field's name
func (f *FieldInvertState) GetName() string {
	return f.name
}

// GetIndexCreatedVersionMajor Return the version that was used to create the index, or 6 if it was created before 7.0.
func (f *FieldInvertState) GetIndexCreatedVersionMajor() int {
	return f.indexCreatedVersionMajor
}

// GetIndexOptions Get the index options for this field
func (f *FieldInvertState) GetIndexOptions() document.IndexOptions {
	return f.indexOptions
}
