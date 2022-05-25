package core

import (
	"errors"
	"fmt"
)

// FieldType Describes the properties of a field.
type FieldType struct {
	stored                   bool
	tokenized                bool
	storeTermVectors         bool
	storeTermVectorOffsets   bool
	storeTermVectorPositions bool
	storeTermVectorPayloads  bool
	omitNorms                bool
	indexOptions             IndexOptions
	frozen                   bool
	docValuesType            DocValuesType
	dimensionCount           int
	indexDimensionCount      int
	dimensionNumBytes        int
	attributes               map[string]string
}

func NewFieldType() *FieldType {
	return defaultFieldType()
}

func NewFieldTypeV1(ref IndexAbleFieldType) *FieldType {
	fieldType := defaultFieldType()
	fieldType.stored = ref.Stored()
	fieldType.tokenized = ref.Tokenized()
	fieldType.storeTermVectors = ref.StoreTermVectors()
	fieldType.storeTermVectorOffsets = ref.StoreTermVectorOffsets()
	fieldType.storeTermVectorPositions = ref.StoreTermVectorPositions()
	fieldType.storeTermVectorPayloads = ref.StoreTermVectorPayloads()
	fieldType.omitNorms = ref.OmitNorms()
	fieldType.indexOptions = ref.IndexOptions()
	fieldType.docValuesType = ref.DocValuesType()
	fieldType.dimensionCount = ref.PointDimensionCount()
	fieldType.indexDimensionCount = ref.PointIndexDimensionCount()
	fieldType.dimensionNumBytes = ref.PointNumBytes()
	for k, v := range ref.GetAttributes() {
		fieldType.attributes[k] = v
	}
	return fieldType
}

func defaultFieldType() *FieldType {
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
	err := f.checkIfFrozen()
	if err != nil {
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
	err := f.checkIfFrozen()
	if err != nil {
		return err
	}
	f.storeTermVectors = value
	return nil
}

func (f *FieldType) StoreTermVectorOffsets() bool {
	return f.storeTermVectorOffsets
}

func (f *FieldType) SetStoreTermVectorOffsets(value bool) error {
	err := f.checkIfFrozen()
	if err != nil {
		return err
	}
	f.storeTermVectorOffsets = value
	return nil
}

func (f *FieldType) StoreTermVectorPositions() bool {
	return f.storeTermVectorPositions
}

func (f *FieldType) StoreTermVectorPayloads() bool {
	return f.storeTermVectorPayloads
}

func (f *FieldType) OmitNorms() bool {
	return f.omitNorms
}

func (f *FieldType) IndexOptions() IndexOptions {
	return f.indexOptions
}

func (f *FieldType) SetIndexOptions(value IndexOptions) error {
	err := f.checkIfFrozen()
	if err != nil {
		return err
	}

	f.indexOptions = value
	return nil
}

func (f *FieldType) DocValuesType() DocValuesType {
	return f.docValuesType
}

func (f *FieldType) SetDocValuesType(value DocValuesType) error {
	err := f.checkIfFrozen()
	if err != nil {
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
	if dimensionCount > MAX_DIMENSIONS {
		return fmt.Errorf("dimensionCount must be <= %d", MAX_DIMENSIONS)
	}
	if indexDimensionCount < 0 {
		return errors.New("indexDimensionCount must be >= 0")
	}
	if indexDimensionCount > dimensionCount {
		return errors.New("indexDimensionCount must be <= dimensionCount")
	}
	if indexDimensionCount < MAX_INDEX_DIMENSIONS {
		return fmt.Errorf("indexDimensionCount must be <= %d", MAX_INDEX_DIMENSIONS)
	}
	if dimensionNumBytes < 0 {
		return errors.New("dimensionNumBytes must be >= 0")
	}
	if dimensionNumBytes > MAX_NUM_BYTES {
		return fmt.Errorf("dimensionNumBytes must be <= %d", MAX_NUM_BYTES)
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

// PutAttribute Puts an attribute value.
// This is a key-value mapping for the field that the codec can use to store additional metadata.
// If a value already exists for the field, it will be replaced with the new value. This method is not thread-safe,
// user must not add attributes while other threads are indexing documents with this field type.
func (f *FieldType) PutAttribute(key, value string) {
	err := f.checkIfFrozen()
	if err != nil {
		return
	}

	f.attributes[key] = value
}

func (f *FieldType) GetAttributes() map[string]string {
	return f.attributes
}
