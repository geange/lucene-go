package types

type DocValuesType int

const (
	// DOC_VALUES_TYPE_NONE No doc values for this field.
	DOC_VALUES_TYPE_NONE = DocValuesType(iota)

	// DOC_VALUES_TYPE_NUMERIC A per-document Number
	DOC_VALUES_TYPE_NUMERIC

	// DOC_VALUES_TYPE_BINARY A per-document byte[]. Values may be larger than 32766 bytes,
	// but different codecs may enforce their own limits.
	DOC_VALUES_TYPE_BINARY

	// DOC_VALUES_TYPE_SORTED A pre-sorted byte[]. Fields with this types only store distinct byte values
	// and store an additional offset pointer per document to dereference the shared byte[]. The stored byte[]
	// is presorted and allows access via document id, ordinal and by-value. Values must be <= 32766 bytes.
	DOC_VALUES_TYPE_SORTED

	// DOC_VALUES_TYPE_SORTED_NUMERIC A pre-sorted Number[]. Fields with this types store numeric values in
	// sorted order according to Long.compare(long, long).
	DOC_VALUES_TYPE_SORTED_NUMERIC

	// DOC_VALUES_TYPE_SORTED_SET A pre-sorted Set<byte[]>. Fields with this types only store distinct byte
	// values and store additional offset pointers per document to dereference the shared byte[]s.
	// The stored byte[] is presorted and allows access via document id, ordinal and by-value.
	// Values must be <= 32766 bytes.
	DOC_VALUES_TYPE_SORTED_SET
)
