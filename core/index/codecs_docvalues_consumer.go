package index

import (
	"io"

	"github.com/geange/lucene-go/core/document"
)

// DocValuesConsumer Abstract API that consumes numeric, binary and sorted docvalues.
// Concrete implementations of this actually do "something" with the docvalues
// (write it into the index in a specific format).
// The lifecycle is:
//  1. DocValuesConsumer is created by NormsFormat.normsConsumer(SegmentWriteState).
//  2. addNumericField, addBinaryField, addSortedField, addSortedSetField, or addSortedNumericField
//     are called for each Numeric, Binary, Sorted, SortedSet, or SortedNumeric docvalues field.
//     The API is a "pull" rather than "push", and the implementation is free to iterate over the
//     values multiple times (Iterable.iterator()).
//  3. After all fields are added, the consumer is closed.
//
// lucene.experimental
type DocValuesConsumer interface {
	io.Closer

	// AddNumericField Writes numeric docvalues for a field.
	// @param field field information
	// @param valuesProducer Numeric values to write.
	// @throws IOException if an I/O error occurred.
	AddNumericField(field *document.FieldInfo, valuesProducer DocValuesProducer) error

	// AddBinaryField Writes binary docvalues for a field.
	// @param field field information
	// @param valuesProducer Binary values to write.
	// @throws IOException if an I/O error occurred.
	AddBinaryField(field *document.FieldInfo, valuesProducer DocValuesProducer) error

	// AddSortedField Writes pre-sorted binary docvalues for a field.
	// @param field field information
	// @param valuesProducer produces the values and ordinals to write
	// @throws IOException if an I/O error occurred.
	AddSortedField(field *document.FieldInfo, valuesProducer DocValuesProducer) error

	// AddSortedNumericField Writes pre-sorted numeric docvalues for a field
	// @param field field information
	// @param valuesProducer produces the values to write
	// @throws IOException if an I/O error occurred.
	AddSortedNumericField(field *document.FieldInfo, valuesProducer DocValuesProducer) error

	// AddSortedSetField Writes pre-sorted set docvalues for a field
	// @param field field information
	// @param valuesProducer produces the values to write
	// @throws IOException if an I/O error occurred.
	AddSortedSetField(field *document.FieldInfo, valuesProducer DocValuesProducer) error
}

// Merges in the fields from the readers in mergeState. The default implementation calls mergeNumericField,
// mergeBinaryField, mergeSortedField, mergeSortedSetField, or mergeSortedNumericField for each field,
// depending on its type. Implementations can override this method for more sophisticated merging
// (bulk-byte copying, etc).
