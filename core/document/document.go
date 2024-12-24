package document

import (
	"iter"
	"slices"
	"strings"
)

// Document
// Documents are the unit of indexing and search. A Document is a set of fields. Each field has a name
// and a textual value. A field may be stored with the document, in which case it is returned with search
// hits on the document. Thus each document should typically contain one or more stored fields which
// uniquely identify it.
//
// Note that fields which are not stored are not available in documents retrieved from the index,
// e.g. with ScoreDoc.doc or IndexReader.document(int).
type Document struct {
	fields []IndexableField
}

func NewDocument(fields ...IndexableField) *Document {
	if len(fields) == 0 {
		fields = make([]IndexableField, 0)
	}
	return &Document{fields: fields}
}

// Add
// Add a field to a document. Several fields may be added with the same name. In this case,
// if the fields are indexed, their text is treated as though appended for the purposes of search.
// Note that add like the removeField(s) methods only makes sense prior to adding a document to an index.
// These methods cannot be used to change the content of an existing index! In order to achieve this,
// a document has to be deleted from an index and a new changed version of that document has to be added.
func (d *Document) Add(field IndexableField) {
	d.fields = append(d.fields, field)
}

// RemoveField
// Removes field with the specified name from the document. If multiple fields exist with this name,
// this method removes the first field that has been added. If there is no field with the specified name,
// the document remains unchanged.
// Note that the removeField(s) methods like the add method only make sense prior to adding a document to an index.
// These methods cannot be used to change the content of an existing index! In order to achieve this, a document
// has to be deleted from an index and a new changed version of that document has to be added.
func (d *Document) RemoveField(name string) {
	for i, field := range d.fields {
		if field.Name() == name {
			d.fields = append(d.fields[:i], d.fields[i+1:]...)
			return
		}
	}
}

// RemoveFields
// Removes all fields with the given name from the document. If there is no field with the
// specified name, the document remains unchanged.
// Note that the removeField(s) methods like the add method only make sense prior to adding a document to an
// index. These methods cannot be used to change the content of an existing index! In order to achieve this,
// a document has to be deleted from an index and a new changed version of that document has to be added.
func (d *Document) RemoveFields(name string) {
	slices.DeleteFunc(d.fields, func(field IndexableField) bool {
		return field.Name() == name
	})
}

// GetField
// Returns a field with the given name if any exist in this document, or null. If multiple fields exists
// with this name, this method returns the first value added.
func (d *Document) GetField(name string) (IndexableField, bool) {
	for _, field := range d.fields {
		if field.Name() == name {
			return field, true
		}
	}
	return nil, false
}

// GetFields
// Returns an array of IndexAbleFields with the given name. This method returns an empty array when
// there are no matching fields. It never returns null.
// name: the name of the field
func (d *Document) GetFields(names ...string) iter.Seq[IndexableField] {
	return func(yield func(IndexableField) bool) {
		for _, field := range d.fields {
			if len(names) > 0 && !slices.Contains(names, field.Name()) {
				continue
			}

			if !yield(field) {
				return
			}
		}
	}
}

// Removes all the fields from document.
func (d *Document) clear() {
	d.fields = d.fields[:0]
}

func (d *Document) SortFieldsByName() {
	slices.SortFunc(d.fields, func(a, b IndexableField) int {
		return strings.Compare(a.Name(), b.Name())
	})
}
