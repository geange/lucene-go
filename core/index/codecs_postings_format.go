package index

// PostingsFormat Encodes/decodes terms, postings, and proximity data.
// Note, when extending this class, the name (getName) may written into the index in certain
// configurations. In order for the segment to be read, the name must resolve to your
// implementation via forName(String). This method uses Java's Service Provider Interface (SPI)
// to resolve format names.
//
// If you implement your own format, make sure that it has a no-arg constructor so SPI can load it.
// ServiceLoader
// lucene.experimental
type PostingsFormat interface {
	NamedSPI

	// FieldsConsumer Writes a new segment
	FieldsConsumer(state *SegmentWriteState) (FieldsConsumer, error)

	// FieldsProducer Reads a segment. NOTE: by the time this call returns, it must hold open any files it
	// will need to use; else, those files may be deleted. Additionally, required files may
	// be deleted during the execution of this call before there is a chance to open them.
	// Under these circumstances an IOException should be thrown by the implementation.
	// IOExceptions are expected and will automatically cause a retry of the segment opening
	// logic with the newly revised segments.
	FieldsProducer(state *SegmentReadState) (FieldsProducer, error)
}
