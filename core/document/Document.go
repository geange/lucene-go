package document

// Document Documents are the unit of indexing and search. A Document is a set of fields. Each field has a name
// and a textual value. A field may be stored with the document, in which case it is returned with search
// hits on the document. Thus each document should typically contain one or more stored fields which
// uniquely identify it.
//
// Note that fields which are not stored are not available in documents retrieved from the index,
// e.g. with ScoreDoc.doc or IndexReader.document(int).
type Document struct {
}
