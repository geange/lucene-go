package document

// StoredField A field whose value is stored so that IndexSearcher.doc and IndexReader.document() will
// return the field and its value.
type StoredField struct {
	*Field
}
