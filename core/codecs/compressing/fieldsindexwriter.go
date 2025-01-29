package compressing

// FieldsIndexWriter
// Efficient index format for block-based Codecs.
// For each block of compressed stored fields, this stores the first document of the block and the start
// pointer of the block in a DirectMonotonicWriter. At read time, the docID is binary-searched in the
// DirectMonotonicReader that records doc IDS, and the returned index is used to look up the start pointer
// in the DirectMonotonicReader that records start pointers.
type FieldsIndexWriter struct {
}
