package core

var (
	TYPE_STORED     *FieldType
	TYPE_NOT_STORED *FieldType
)

func init() {
	TYPE_STORED = NewFieldType()
	TYPE_STORED.SetIndexOptions(INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS)
	TYPE_STORED.SetTokenized(true)
	TYPE_STORED.SetStored(true)
	TYPE_STORED.Freeze()

	TYPE_NOT_STORED = NewFieldType()
	TYPE_NOT_STORED.SetIndexOptions(INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS)
	TYPE_NOT_STORED.SetTokenized(true)
	TYPE_NOT_STORED.Freeze()
}

type TextField struct {
	*Field
}
