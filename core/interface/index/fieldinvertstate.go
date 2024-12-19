package index

import (
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/util/attribute"
)

// FieldInvertState
// This class tracks the number and position / offset parameters of terms being added to the index.
// The information collected in this class is also used to calculate the normalization factor for a field.
type FieldInvertState struct {
	IndexCreatedVersionMajor int
	Name                     string
	IndexOptions             document.IndexOptions
	Position                 int
	Length                   int
	NumOverlap               int
	Offset                   int
	MaxTermFrequency         int
	UniqueTermCount          int
	// we must track these across field instances (multi-valued case)
	LastStartOffset int
	LastPosition    int

	AttributeSource   *attribute.Source
	OffsetAttribute   attribute.OffsetAttr
	PosIncrAttribute  attribute.PositionIncrAttr
	PayloadAttribute  attribute.PayloadAttr
	TermAttribute     attribute.Term2BytesAttr
	TermFreqAttribute attribute.TermFreqAttr
}

// NewFieldInvertState Creates {code FieldInvertState} for the specified field name and values for all fields.
func NewFieldInvertState(indexCreatedVersionMajor int, name string, indexOptions document.IndexOptions, position int, length int, numOverlap int, offset int, maxTermFrequency int, uniqueTermCount int) *FieldInvertState {
	return &FieldInvertState{
		IndexCreatedVersionMajor: indexCreatedVersionMajor,
		Name:                     name,
		IndexOptions:             indexOptions,
		Position:                 position,
		Length:                   length,
		NumOverlap:               numOverlap,
		Offset:                   offset,
		MaxTermFrequency:         maxTermFrequency,
		UniqueTermCount:          uniqueTermCount,
	}
}

func NewFieldInvertStateV1(indexCreatedVersionMajor int, name string, indexOptions document.IndexOptions) *FieldInvertState {
	return &FieldInvertState{
		IndexCreatedVersionMajor: indexCreatedVersionMajor,
		Name:                     name,
		IndexOptions:             indexOptions,
	}
}

func (f *FieldInvertState) Reset() {
	f.Position = -1
	f.Length = 0
	f.NumOverlap = 0
	f.Offset = 0
	f.MaxTermFrequency = 0
	f.UniqueTermCount = 0
	f.LastStartOffset = 0
	f.LastPosition = 0
}

func (f *FieldInvertState) SetAttributeSource(attributeSource *attribute.Source) {
	if f.AttributeSource != attributeSource {
		f.AttributeSource = attributeSource
		f.TermAttribute = attributeSource.CharTerm()
		f.TermFreqAttribute = attributeSource.TermFrequency()
		f.PosIncrAttribute = attributeSource.PositionIncrement()
		f.OffsetAttribute = attributeSource.Offset()
		f.PayloadAttribute = attributeSource.Payload()
	}
}

// GetPosition Get the last processed term position.
// Returns: the position
func (f *FieldInvertState) GetPosition() int {
	return f.Position
}

// GetLength Get total number of terms in this field.
// Returns: the length
func (f *FieldInvertState) GetLength() int {
	return f.Length
}

// SetLength Set length item.
func (f *FieldInvertState) SetLength(length int) {
	f.Length = length
}

// GetNumOverlap Get the number of terms with positionIncrement == 0.
// Returns: the numOverlap
func (f *FieldInvertState) GetNumOverlap() int {
	return f.NumOverlap
}

// SetNumOverlap Set number of terms with positionIncrement == 0.
func (f *FieldInvertState) SetNumOverlap(numOverlap int) {
	f.NumOverlap = numOverlap
}

// GetOffset Get end offset of the last processed term.
// Returns: the offset
func (f *FieldInvertState) GetOffset() int {
	return f.Offset
}

// GetMaxTermFrequency Get the maximum term-frequency encountered for any term in the field. A field
// containing "the quick brown fox jumps over the lazy dog" would have a item of 2, because "the" appears twice.
func (f *FieldInvertState) GetMaxTermFrequency() int {
	return f.MaxTermFrequency
}

// GetUniqueTermCount Return the number of unique terms encountered in this field.
func (f *FieldInvertState) GetUniqueTermCount() int {
	return f.UniqueTermCount
}

// GetAttributeSource Returns the AttributeSourceV2 from the TokenStream that provided the indexed tokens for this field.
func (f *FieldInvertState) GetAttributeSource() *attribute.Source {
	return f.AttributeSource
}

// GetName Return the field's name
func (f *FieldInvertState) GetName() string {
	return f.Name
}

// GetIndexCreatedVersionMajor Return the version that was used to create the index, or 6 if it was created before 7.0.
func (f *FieldInvertState) GetIndexCreatedVersionMajor() int {
	return f.IndexCreatedVersionMajor
}

// GetIndexOptions Get the index options for this field
func (f *FieldInvertState) GetIndexOptions() document.IndexOptions {
	return f.IndexOptions
}

func (f *FieldInvertState) GetPosIncrAttribute() attribute.PositionIncrAttr {
	return f.PosIncrAttribute
}

// 	attributeSource   *attribute.Source
//	offsetAttribute   attribute.OffsetAttr
//	posIncrAttribute  attribute.PositionIncrAttr
//	payloadAttribute  attribute.PayloadAttr
//	termAttribute     attribute.Term2BytesAttr
//	termFreqAttribute attribute.TermFreqAttr
