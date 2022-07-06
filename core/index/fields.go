package index

// Fields Provides a Terms index for fields that have it, and lists which fields do. This is primarily an
// internal/experimental API (see FieldsProducer), although it is also used to expose the set of term
// vectors per document.
type Fields interface {
	// Iterator Returns an iterator that will step through all fields names. This will not return null.
	Iterator() func() string

	// Terms Get the Terms for this field. This will return null if the field does not exist.
	Terms(field string) (Terms, error)

	// Size Returns the number of fields or -1 if the number of distinct field names is unknown. If >= 0,
	// iterator will return as many field names.
	Size() int

	// Zero-length Fields array.

}
