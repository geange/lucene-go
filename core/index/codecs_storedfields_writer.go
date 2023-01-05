package index

import (
	"github.com/geange/lucene-go/core/types"
	"io"
)

// StoredFieldsWriter Codec API for writing stored fields:
// 1. For every document, startDocument() is called, informing the Codec that a new document has started.
// 2. writeField(FieldInfo, IndexableField) is called for each field in the document.
// 3. After all documents have been written, finish(FieldInfos, int) is called for verification/sanity-checks.
// 4. Finally the writer is closed (close())
// lucene.experimental
type StoredFieldsWriter interface {
	io.Closer

	// StartDocument Called before writing the stored fields of the document. writeField(FieldInfo, IndexableField) will be called for each stored field. Note that this is called even if the document has no stored fields.
	StartDocument() error

	// FinishDocument Called when a document and all its fields have been added.
	FinishDocument() error

	// WriteField Writes a single stored field.
	WriteField(info *types.FieldInfo, field types.IndexableField) error

	// Finish Called before close(), passing in the number of documents that were written. Note that this is intentionally redundant (equivalent to the number of calls to startDocument(), but a Codec should check that this is the case to detect the JRE bug described in LUCENE-1282.
	Finish(fis FieldInfos, numDocs int) error
}
