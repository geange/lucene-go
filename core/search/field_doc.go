package search

var _ ScoreDoc = &FieldDoc{}

// FieldDoc
// Expert: A ScoreDoc which also contains information about how to sort the referenced document.
// In addition to the document number and score, this object contains an array of values for the
// document from the field(s) used to sort. For example, if the sort criteria was to sort by
// fields "a", "b" then "c", the fields object array will have three elements, corresponding
// respectively to the term values for the document in fields "a", "b" and "c". The class of
// each element in the array will be either Integer, Float or String depending on the type of
// values in the terms of each field.
//
// Created: Feb 11, 2004 1:23:38 PM
// Since: lucene 1.4
// See Also: ScoreDoc, TopFieldDocs
type FieldDoc struct {
	// Expert: The values which are used to sort the referenced document. The order of these will match the original sort criteria given by a Sort object. Each Object will have been returned from the value method corresponding FieldComparator used to sort this field.
	// See Also: Sort, IndexSearcher.search(Query, int, Sort)
	fields []any

	*ScoreDocDefault
}

// NewFieldDoc Expert: Creates one of these objects with empty sort information.
func NewFieldDoc(doc int, score float64) *FieldDoc {
	return &FieldDoc{
		fields:          make([]any, 0),
		ScoreDocDefault: NewScoreDoc(score, doc),
	}
}

// NewFieldDocV1
// Expert: Creates one of these objects with the given sort information.
func NewFieldDocV1(doc int, score float64, fields []any) *FieldDoc {
	return &FieldDoc{
		fields:          fields,
		ScoreDocDefault: NewScoreDoc(score, doc),
	}
}

// NewFieldDocV2
// Expert: Creates one of these objects with the given sort information.
func NewFieldDocV2(doc int, score float64, fields []any, shardIndex int) *FieldDoc {
	return &FieldDoc{
		fields:          fields,
		ScoreDocDefault: NewScoreDocV1(score, doc, shardIndex),
	}
}

func (f *FieldDoc) GetFields() []any {
	return f.fields
}

func (f *FieldDoc) SetFields(fields []any) {
	f.fields = fields
}
