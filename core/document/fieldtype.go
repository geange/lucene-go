package document

import (
	"errors"
	"fmt"
)

// FieldType
// Describes the properties of a field.
type FieldType struct {
	stored                   bool
	tokenized                bool
	storeTermVectors         bool
	storeTermVectorOffsets   bool
	storeTermVectorPositions bool
	storeTermVectorPayloads  bool
	omitNorms                bool
	frozen                   bool
	indexOptions             IndexOptions
	docValuesType            DocValuesType
	dimensionCount           int
	indexDimensionCount      int
	dimensionNumBytes        int
	attributes               map[string]string
}

func NewFieldType() *FieldType {
	return newFieldType()
}

func NewFieldTypeFrom(fieldType IndexableFieldType) *FieldType {
	t := newFieldType()
	t.stored = fieldType.Stored()
	t.tokenized = fieldType.Tokenized()
	t.storeTermVectors = fieldType.StoreTermVectors()
	t.storeTermVectorOffsets = fieldType.StoreTermVectorOffsets()
	t.storeTermVectorPositions = fieldType.StoreTermVectorPositions()
	t.storeTermVectorPayloads = fieldType.StoreTermVectorPayloads()
	t.omitNorms = fieldType.OmitNorms()
	t.indexOptions = fieldType.IndexOptions()
	t.docValuesType = fieldType.DocValuesType()
	t.dimensionCount = fieldType.PointDimensionCount()
	t.indexDimensionCount = fieldType.PointIndexDimensionCount()
	t.dimensionNumBytes = fieldType.PointNumBytes()
	for k, v := range fieldType.GetAttributes() {
		t.attributes[k] = v
	}
	return t
}

func newFieldType() *FieldType {
	return &FieldType{
		stored:                   false,
		tokenized:                true,
		storeTermVectors:         false,
		storeTermVectorOffsets:   false,
		storeTermVectorPositions: false,
		storeTermVectorPayloads:  false,
		omitNorms:                false,
		indexOptions:             INDEX_OPTIONS_NONE,
		frozen:                   false,
		docValuesType:            DOC_VALUES_TYPE_NONE,
		dimensionCount:           0,
		indexDimensionCount:      0,
		dimensionNumBytes:        0,
		attributes:               make(map[string]string),
	}
}

func (f *FieldType) checkIfFrozen() error {
	if f.frozen {
		return errors.New("this FieldType is already frozen and cannot be changed")
	}
	return nil
}

func (f *FieldType) Freeze() {
	f.frozen = true
}

func (f *FieldType) Stored() bool {
	return f.stored
}

func (f *FieldType) SetStored(value bool) error {
	if err := f.checkIfFrozen(); err != nil {
		return err
	}

	f.stored = value
	return nil
}

func (f *FieldType) Tokenized() bool {
	return f.tokenized
}

func (f *FieldType) SetTokenized(value bool) error {
	err := f.checkIfFrozen()
	if err != nil {
		return err
	}

	f.tokenized = value
	return nil
}

func (f *FieldType) StoreTermVectors() bool {
	return f.storeTermVectors
}

func (f *FieldType) SetStoreTermVectors(value bool) error {
	if err := f.checkIfFrozen(); err != nil {
		return err
	}
	f.storeTermVectors = value
	return nil
}

func (f *FieldType) StoreTermVectorOffsets() bool {
	return f.storeTermVectorOffsets
}

func (f *FieldType) SetStoreTermVectorOffsets(value bool) error {
	if err := f.checkIfFrozen(); err != nil {
		return err
	}
	f.storeTermVectorOffsets = value
	return nil
}

func (f *FieldType) StoreTermVectorPositions() bool {
	return f.storeTermVectorPositions
}

// SetStoreTermVectorPositions
// Set to true to also store token positions into the term vector for this field.
// value: true if this field should store term vector positions.
func (f *FieldType) SetStoreTermVectorPositions(value bool) error {
	if err := f.checkIfFrozen(); err != nil {
		return err
	}
	f.storeTermVectorPositions = value
	return nil
}

func (f *FieldType) StoreTermVectorPayloads() bool {
	return f.storeTermVectorPayloads
}

// SetStoreTermVectorPayloads
// Set to true to also store token payloads into the term vector for this field.
// value: true if this field should store term vector payloads.
// 抛出: IllegalStateException – if this FieldType is frozen against future modifications.
// 请参阅: storeTermVectorPayloads()
func (f *FieldType) SetStoreTermVectorPayloads(value bool) error {
	if err := f.checkIfFrozen(); err != nil {
		return err
	}
	f.storeTermVectorPayloads = value
	return nil
}

func (f *FieldType) OmitNorms() bool {
	return f.omitNorms
}

func (f *FieldType) SetOmitNorms(value bool) error {
	f.omitNorms = value
	return nil
}

func (f *FieldType) IndexOptions() IndexOptions {
	return f.indexOptions
}

func (f *FieldType) SetIndexOptions(value IndexOptions) error {
	if err := f.checkIfFrozen(); err != nil {
		return err
	}

	f.indexOptions = value
	return nil
}

func (f *FieldType) DocValuesType() DocValuesType {
	return f.docValuesType
}

func (f *FieldType) SetDocValuesType(value DocValuesType) error {
	if err := f.checkIfFrozen(); err != nil {
		return err
	}

	f.docValuesType = value
	return nil
}

func (f *FieldType) SetDimensions(dimensionCount, dimensionNumBytes int) error {
	return f.SetDimensionsV1(dimensionCount, dimensionCount, dimensionNumBytes)
}

// SetDimensionsV1 Enables points indexing with selectable dimension indexing.
func (f *FieldType) SetDimensionsV1(dimensionCount, indexDimensionCount, dimensionNumBytes int) error {
	if dimensionCount < 0 {
		return errors.New("dimensionCount must be >= 0")
	}
	if dimensionCount > MaxDimensions {
		return fmt.Errorf("dimensionCount must be <= %d", MaxDimensions)
	}
	if indexDimensionCount < 0 {
		return errors.New("indexDimensionCount must be >= 0")
	}
	if indexDimensionCount > dimensionCount {
		return errors.New("indexDimensionCount must be <= dimensionCount")
	}
	if indexDimensionCount < MaxIndexDimensions {
		return fmt.Errorf("indexDimensionCount must be <= %d", MaxIndexDimensions)
	}
	if dimensionNumBytes < 0 {
		return errors.New("dimensionNumBytes must be >= 0")
	}
	if dimensionNumBytes > MaxNumBytes {
		return fmt.Errorf("dimensionNumBytes must be <= %d", MaxNumBytes)
	}
	if dimensionCount == 0 {
		if indexDimensionCount != 0 {
			return errors.New("when dimensionCount is 0, indexDimensionCount must be 0")
		}

		if dimensionNumBytes != 0 {
			return errors.New("when dimensionCount is 0, dimensionNumBytes must be 0")
		}
	} else if indexDimensionCount == 0 {
		return errors.New("when dimensionCount is > 0, indexDimensionCount must be > 0")
	} else if dimensionCount == 0 {
		return errors.New("when dimensionNumBytes is 0, dimensionCount must be 0")
	}

	f.dimensionCount = dimensionCount
	f.indexDimensionCount = indexDimensionCount
	f.dimensionNumBytes = dimensionNumBytes
	return nil
}

func (f *FieldType) PointDimensionCount() int {
	return f.dimensionCount
}

func (f *FieldType) PointIndexDimensionCount() int {
	return f.indexDimensionCount
}

func (f *FieldType) PointNumBytes() int {
	return f.dimensionNumBytes
}

// PutAttribute
// Puts an attribute value.
// This is a key-value mapping for the field that the codec can use to store additional metadata.
// If a value already exists for the field, it will be replaced with the new value. This method is not thread-safe,
// user must not add attributes while other threads are indexing documents with this field types.
func (f *FieldType) PutAttribute(key, value string) {
	if err := f.checkIfFrozen(); err != nil {
		return
	}

	f.attributes[key] = value
}

func (f *FieldType) GetAttributes() map[string]string {
	return f.attributes
}
