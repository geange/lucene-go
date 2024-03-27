package document

import (
	"errors"
	"fmt"
)

// FieldInfo
// Access to the Field Info file that describes document fields and whether or not they are indexed.
// Each segment has a separate Field Info file. Objects of this class are thread-safe for multiple readers,
// but only one thread can be adding documents at a time, with no other reader or writer threads accessing this object.
type FieldInfo struct {
	name             string            // Field's name
	number           int               // Internal field number
	docValuesType    DocValuesType     //
	storeTermVector  bool              // True if any document indexed term vectors
	omitNorms        bool              // omit norms associated with indexed fields
	indexOptions     IndexOptions      //
	storePayloads    bool              // whether this field stores payloads together with term positions
	attributes       map[string]string //
	dvGen            int64             //
	softDeletesField bool              // whether this field is used as the soft-deletes field

	// If both of these are positive it means this field indexed points (see org.apache.lucene.codecs.PointsFormat).
	pointDimensionCount      int
	pointIndexDimensionCount int
	pointNumBytes            int
}

func NewFieldInfo(name string, number int, storeTermVector, omitNorms, storePayloads bool,
	indexOptions IndexOptions, docValues DocValuesType, dvGen int64, attributes map[string]string,
	pointDimensionCount, pointIndexDimensionCount, pointNumBytes int, softDeletesField bool) *FieldInfo {

	info := &FieldInfo{
		name:                     name,
		number:                   number,
		docValuesType:            docValues,
		storeTermVector:          storePayloads,
		omitNorms:                false,
		indexOptions:             indexOptions,
		storePayloads:            false,
		attributes:               attributes,
		dvGen:                    dvGen,
		pointDimensionCount:      pointDimensionCount,
		pointIndexDimensionCount: pointIndexDimensionCount,
		pointNumBytes:            pointNumBytes,
		softDeletesField:         softDeletesField,
	}

	if info.indexOptions != INDEX_OPTIONS_NONE {
		info.storeTermVector = storeTermVector
		info.storePayloads = storePayloads
		info.omitNorms = omitNorms
	} else {
		info.storeTermVector = false
		info.storePayloads = false
		info.omitNorms = false
	}
	return info
}

// Performs internal consistency checks. Always returns nil (or throws IllegalStateException)
func (f *FieldInfo) checkConsistency() error {
	return nil
}

func (f *FieldInfo) Name() string {
	return f.name
}

func (f *FieldInfo) Number() int {
	return f.number
}

// SetPointDimensions Record that this field is indexed with points, with the specified number of
// dimensions and bytes per dimension.
func (f *FieldInfo) SetPointDimensions(dimensionCount, indexDimensionCount, numBytes int) error {

	f.pointDimensionCount = dimensionCount
	f.pointIndexDimensionCount = indexDimensionCount
	f.pointNumBytes = numBytes

	return f.checkConsistency()
}

// GetPointDimensionCount Return point data dimension count
func (f *FieldInfo) GetPointDimensionCount() int {
	return f.pointDimensionCount
}

// GetPointIndexDimensionCount Return point data dimension count
func (f *FieldInfo) GetPointIndexDimensionCount() int {
	return f.pointIndexDimensionCount
}

// GetPointNumBytes Return number of bytes per dimension
func (f *FieldInfo) GetPointNumBytes() int {
	return f.pointNumBytes
}

// SetDocValuesType Record that this field is indexed with docvalues, with the specified types
func (f *FieldInfo) SetDocValuesType(_type DocValuesType) error {
	f.docValuesType = _type
	return f.checkConsistency()
}

// GetIndexOptions Returns IndexOptions for the field, or IndexOptions.NONE if the field is not indexed
func (f *FieldInfo) GetIndexOptions() IndexOptions {
	return f.indexOptions
}

// SetIndexOptions Record the IndexOptions to use with this field.
func (f *FieldInfo) SetIndexOptions(newIndexOptions IndexOptions) error {
	f.indexOptions = newIndexOptions

	if f.indexOptions == INDEX_OPTIONS_NONE || f.indexOptions < INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS {
		// cannot store payloads if we don't store positions:
		f.storePayloads = false
	}

	return f.checkConsistency()
}

// GetDocValuesType Returns DocValuesType of the docValues; this is DocValuesType.NONE if the field has no docvalues.
func (f *FieldInfo) GetDocValuesType() DocValuesType {
	return f.docValuesType
}

// SetDocValuesGen Sets the docValues generation of this field.
func (f *FieldInfo) SetDocValuesGen(dvGen int64) error {
	f.dvGen = dvGen
	return f.checkConsistency()
}

// GetDocValuesGen Returns the docValues generation of this field, or -1 if no docValues updates exist for it.
func (f *FieldInfo) GetDocValuesGen() int64 {
	return f.dvGen
}

func (f *FieldInfo) SetStoreTermVectors() error {
	f.storeTermVector = true
	return f.checkConsistency()
}

func (f *FieldInfo) SetStorePayloads() error {
	if f.indexOptions != INDEX_OPTIONS_NONE && f.indexOptions >= INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS {
		f.storePayloads = true
	}
	return f.checkConsistency()
}

// OmitsNorms Returns true if norms are explicitly omitted for this field
func (f *FieldInfo) OmitsNorms() bool {
	return f.omitNorms
}

// SetOmitsNorms Omit norms for this field.
func (f *FieldInfo) SetOmitsNorms() error {
	if f.indexOptions == INDEX_OPTIONS_NONE {
		return errors.New("cannot omit norms: this field is not indexed")
	}
	f.omitNorms = true
	return f.checkConsistency()
}

// HasNorms Returns true if this field actually has any norms.
func (f *FieldInfo) HasNorms() bool {
	return f.indexOptions != INDEX_OPTIONS_NONE && f.omitNorms == false
}

// HasPayloads Returns true if any payloads exist for this field.
func (f *FieldInfo) HasPayloads() bool {
	return f.storePayloads
}

// HasVectors Returns true if any term vectors exist for this field.
func (f *FieldInfo) HasVectors() bool {
	return f.storeTermVector
}

// GetAttribute Get a codec attribute value, or null if it does not exist
func (f *FieldInfo) GetAttribute(key string) string {
	return f.attributes[key]
}

// PutAttribute Puts a codec attribute value.
// This is a key-value mapping for the field that the codec can use to store additional metadata,
// and will be available to the codec when reading the segment via getAttribute(String)
// If a value already exists for the key in the field, it will be replaced with the new value.
// If the value of the attributes for a same field is changed between the documents, the behaviour
// after merge is undefined.
func (f *FieldInfo) PutAttribute(key, value string) {
	f.attributes[key] = value
}

// Attributes Returns internal codec attributes map.
func (f *FieldInfo) Attributes() map[string]string {
	return f.attributes
}

// IsSoftDeletesField Returns true if this field is configured and used as the soft-deletes field.
// See IndexWriterConfig.softDeletesField
func (f *FieldInfo) IsSoftDeletesField() bool {
	return f.softDeletesField
}

// Update should only be called by FieldInfos#addOrUpdate
func (f *FieldInfo) Update(storeTermVector, omitNorms, storePayloads bool, indexOptions IndexOptions,
	attributes map[string]string, dimensionCount, indexDimensionCount, dimensionNumBytes int) error {

	if f.indexOptions != indexOptions {
		if f.indexOptions == INDEX_OPTIONS_NONE {
			f.indexOptions = indexOptions
		} else if f.indexOptions != INDEX_OPTIONS_NONE {
			return fmt.Errorf(
				`cannot change field "%s" from index options=%s to inconsistent index options=%s`,
				f.Name(), f.indexOptions, indexOptions,
			)
		}
	}

	if f.pointDimensionCount == 0 && dimensionCount != 0 {
		f.pointDimensionCount = dimensionCount
		f.pointIndexDimensionCount = indexDimensionCount
		f.pointNumBytes = dimensionNumBytes
	} else if dimensionCount != 0 &&
		(f.pointDimensionCount != dimensionCount ||
			f.pointIndexDimensionCount != indexDimensionCount ||
			f.pointNumBytes != dimensionNumBytes) {

		return fmt.Errorf(`cannot change field "%s" from points dimensionCount=%d, indexDimensionCount=%d`,
			f.Name(), f.pointDimensionCount, f.pointIndexDimensionCount,
		)
	}

	if f.indexOptions != INDEX_OPTIONS_NONE { // if updated field data is not for indexing, leave the updates out
		f.storeTermVector = f.storeTermVector || storeTermVector // once vector, always vector
		f.storePayloads = f.storePayloads || storePayloads

		// Awkward: only drop norms if incoming update is indexed:
		if indexOptions != INDEX_OPTIONS_NONE && f.omitNorms != omitNorms {
			f.omitNorms = true // if one require omitNorms at least once, it remains off for life
		}
	}

	if f.indexOptions == INDEX_OPTIONS_NONE || f.indexOptions < INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS {
		// cannot store payloads if we don't store positions:
		f.storePayloads = false
	}

	for k, v := range attributes {
		f.attributes[k] = v
	}

	return f.checkConsistency()
}
