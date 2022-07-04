package index

import (
	"github.com/geange/lucene-go/core/tokenattributes"
	"github.com/geange/lucene-go/core/types"
)

// FieldInvertState This class tracks the number and position / offset parameters of terms being added to the index.
// The information collected in this class is also used to calculate the normalization factor for a field.
type FieldInvertState struct {
	indexCreatedVersionMajor int
	name                     string
	indexOptions             types.IndexOptions
	position                 int
	length                   int
	numOverlap               int
	offset                   int
	maxTermFrequency         int
	uniqueTermCount          int
	// we must track these across field instances (multi-valued case)
	lastStartOffset int
	lastPosition    int
	attributeSource *tokenattributes.AttributeSource

	offsetAttribute   tokenattributes.OffsetAttribute
	posIncrAttribute  tokenattributes.PositionIncrementAttribute
	payloadAttribute  tokenattributes.PayloadAttribute
	termAttribute     tokenattributes.TermToBytesRefAttribute
	termFreqAttribute tokenattributes.TermFrequencyAttribute
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

// SetLength Set length value.
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
// containing "the quick brown fox jumps over the lazy dog" would have a value of 2, because "the" appears twice.
func (f *FieldInvertState) GetMaxTermFrequency() int {
	return f.maxTermFrequency
}

// GetUniqueTermCount Return the number of unique terms encountered in this field.
func (f *FieldInvertState) GetUniqueTermCount() int {
	return f.uniqueTermCount
}

// GetAttributeSource Returns the AttributeSource from the TokenStream that provided the indexed tokens for this field.
func (f *FieldInvertState) GetAttributeSource() *tokenattributes.AttributeSource {
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
func (f *FieldInvertState) GetIndexOptions() types.IndexOptions {
	return f.indexOptions
}
