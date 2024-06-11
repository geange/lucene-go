package codecs

import (
	"github.com/geange/lucene-go/core/interface/index"
)

// FieldsProducer Abstract API that produces terms, doc, freq, prox, offset and payloads postings.
type FieldsProducer interface {
	index.Fields
}
