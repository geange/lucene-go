package core

// FieldInfo Access to the Field Info file that describes document fields and whether or not they are indexed.
// Each segment has a separate Field Info file. Objects of this class are thread-safe for multiple readers,
// but only one thread can be adding documents at a time, with no other reader or writer threads accessing this object.
type FieldInfo struct {
	name          string
	number        int
	docValuesType DocValuesType

	// True if any document indexed term vectors
	storeTermVector bool

	// omit norms associated with indexed fields
	omitNorms bool

	indexOptions IndexOptions

	// whether this field stores payloads together with term positions
	storePayloads bool

	attributes map[string]string

	dvGen int64

	// If both of these are positive it means this field indexed points (see org.apache.lucene.codecs.PointsFormat).
	pointDimensionCount      int
	pointIndexDimensionCount int
	pointNumBytes            int

	// whether this field is used as the soft-deletes field
	softDeletesField bool
}

// Performs internal consistency checks. Always returns nil (or throws IllegalStateException)
func (f *FieldInfo) checkConsistency() error {
	panic("")
}

// SetPointDimensions Record that this field is indexed with points, with the specified number of
// dimensions and bytes per dimension.
func (f *FieldInfo) SetPointDimensions(dimensionCount, indexDimensionCount, numBytes int) error {
	panic("")
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

// SetDocValuesType Record that this field is indexed with docvalues, with the specified type
func (f *FieldInfo) SetDocValuesType(_type DocValuesType) error {
	panic("")
}

// GetIndexOptions Returns IndexOptions for the field, or IndexOptions.NONE if the field is not indexed
func (f *FieldInfo) GetIndexOptions() IndexOptions {
	return f.indexOptions
}

// SetIndexOptions Record the IndexOptions to use with this field.
func (f *FieldInfo) SetIndexOptions(newIndexOptions IndexOptions) error {
	panic("")
}

// GetDocValuesType Returns DocValuesType of the docValues; this is DocValuesType.NONE if the field has no docvalues.
func (f *FieldInfo) GetDocValuesType() DocValuesType {
	return f.docValuesType
}

// SetDocValuesGen Sets the docValues generation of this field.
func (f *FieldInfo) SetDocValuesGen(dvGen int64) error {
	panic("")
}

// GetDocValuesGen Returns the docValues generation of this field, or -1 if no docValues updates exist for it.
func (f *FieldInfo) GetDocValuesGen() int64 {
	return f.dvGen
}

func (f *FieldInfo) SetStoreTermVectors() error {
	panic("")
}

func (f *FieldInfo) SetStorePayloads() error {
	panic("")
}

// OmitsNorms Returns true if norms are explicitly omitted for this field
func (f *FieldInfo) OmitsNorms() bool {
	return f.omitNorms
}

// SetOmitsNorms Omit norms for this field.
func (f *FieldInfo) SetOmitsNorms() error {
	panic("")
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
