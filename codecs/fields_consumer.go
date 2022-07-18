package codecs

// FieldsConsumer Abstract API that consumes terms, doc, freq, prox, offset and payloads postings.
// Concrete implementations of this actually do "something" with the postings (write it into the
// index in a specific format).
type FieldsConsumer interface {
}
