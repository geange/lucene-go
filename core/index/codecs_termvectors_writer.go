package index

import (
	"io"

	"github.com/geange/lucene-go/core/document"
)

// TermVectorsWriter Codec API for writing term vectors:
// 1. For every document, startDocument(int) is called, informing the Codec how many fields will be written.
// 2. startField(FieldInfo, int, boolean, boolean, boolean) is called for each field in the document, informing the codec how many terms will be written for that field, and whether or not positions, offsets, or payloads are enabled.
// 3. Within each field, startTerm(BytesRef, int) is called for each term.
// 4. If offsets and/or positions are enabled, then addPosition(int, int, int, BytesRef) will be called for each term occurrence.
// 5. After all documents have been written, finish(FieldInfos, int) is called for verification/sanity-checks.
// 6. Finally the writer is closed (close())
// lucene.experimental
type TermVectorsWriter interface {
	io.Closer

	// StartDocument Called before writing the term vectors of the document. startField(FieldInfo, int, boolean, boolean, boolean) will be called numVectorFields times. Note that if term vectors are enabled, this is called even if the document has no vector fields, in this case numVectorFields will be zero.
	StartDocument(numVectorFields int) error

	// FinishDocument Called after a doc and all its fields have been added.
	FinishDocument() error

	// StartField Called before writing the terms of the field. startTerm(BytesRef, int) will be called numTerms times.
	StartField(info *document.FieldInfo, numTerms int, positions, offsets, payloads bool) error

	// FinishField Called after a field and all its terms have been added.
	FinishField() error

	// StartTerm Adds a term and its term frequency freq. If this field has positions and/or offsets enabled, then addPosition(int, int, int, BytesRef) will be called freq times respectively.
	StartTerm(term []byte, freq int) error

	// FinishTerm Called after a term and all its positions have been added.
	FinishTerm() error

	// AddPosition Adds a term position and offsets
	AddPosition(position, startOffset, endOffset int, payload []byte) error

	// Finish Called before close(), passing in the number of documents that were written. Note that this is intentionally redundant (equivalent to the number of calls to startDocument(int), but a Codec should check that this is the case to detect the JRE bug described in LUCENE-1282.
	Finish(fis *FieldInfos, numDocs int) error
}
