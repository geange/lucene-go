package document

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

func NewDocument() *Document {
	return &Document{fields: make([]IndexableField, 0)}
}

func (d *Document) Iterator() func() IndexableField {
	idx := 0
	return func() IndexableField {
		if idx >= len(d.fields) {
			return nil
		}
		field := d.fields[idx]
		idx++
		return field
	}
}

func (d *Document) Fields() []IndexableField {
	return d.fields
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
	tmp := make([]IndexableField, 0, len(d.fields))
	for i, field := range d.fields {
		if field.Name() != name {
			tmp = append(tmp, d.fields[i])
		}
	}
	d.fields = tmp
}

// GetBinaryValues
// Returns [][]byte for of the fields that have the name specified as the method parameter.
// This method returns an empty array when there are no matching fields. It never returns null.
// name: the name of the field
func (d *Document) GetBinaryValues(name string) [][]byte {
	ret := make([][]byte, 0, len(d.fields))
	for _, field := range d.fields {
		if field.Name() == name {
			switch v := field.Get().(type) {
			case string:
				ret = append(ret, []byte(v))
			case []byte:
				ret = append(ret, v)
			}
		}
	}
	return ret
}

// GetBinaryValue
// Returns an array of bytes for the first (or only) field that has the name specified as the method
// parameter. This method will return null if no binary fields with the specified name are available.
// There may be non-binary fields with the same name.
// Params: name – the name of the field.
// Returns: a BytesRef containing the binary field value or null
func (d *Document) GetBinaryValue(name string) ([]byte, error) {
	for _, field := range d.fields {
		if field.Name() == name {
			switch v := field.Get().(type) {
			case string:
				return []byte(v), nil
			case []byte:
				return v, nil
			}
		}
	}
	return nil, ErrFieldValueTypeNotFit
}

// GetField Returns a field with the given name if any exist in this document, or null. If multiple fields exists
// with this name, this method returns the first value added.
func (d *Document) GetField(name string) (IndexableField, error) {
	for _, field := range d.fields {
		if field.Name() == name {
			return field, nil
		}
	}
	return nil, ErrFieldNotFound
}

// GetFields
// Returns an array of IndexAbleFields with the given name. This method returns an empty array when
// there are no matching fields. It never returns null.
// name: the name of the field
func (d *Document) GetFields(name string) []IndexableField {
	ret := make([]IndexableField, 0)
	for i, field := range d.fields {
		if field.Name() == name {
			ret = append(ret, d.fields[i])
		}
	}
	return ret
}

// GetValues
// Returns an array of values of the field specified as the method parameter. This method returns
// an empty array when there are no matching fields. It never returns null. For a numeric StoredField
// it returns the string value of the number. If you want the actual numeric field instances back, use getFields.
// name: the name of the field
func (d *Document) GetValues(name string) []string {
	ret := make([]string, 0, len(d.fields))
	for _, field := range d.fields {
		if field.Name() == name {
			switch v := field.Get().(type) {
			case string:
				ret = append(ret, v)
			case []byte:
				ret = append(ret, string(v))
			}
		}
	}
	return ret
}

// Get
// Returns the string value of the field with the given name if any exist in this document, or null.
// If multiple fields exist with this name, this method returns the first value added. If only binary
// fields with this name exist, returns null. For a numeric StoredField it returns the string value of
// the number. If you want the actual numeric field instance back, use getField.
func (d *Document) Get(name string) (string, error) {
	for _, field := range d.fields {
		if field.Name() == name {
			switch v := field.Get().(type) {
			case string:
				return v, nil
			case []byte:
				return string(v), nil
			}
		}
	}
	return "", ErrFieldValueTypeNotFit
}

// Removes all the fields from document.
func (d *Document) clear() {
	d.fields = d.fields[:0]
}
