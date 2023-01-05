package index

// PointsFormat Encodes/decodes indexed points.
// lucene.experimental
type PointsFormat interface {

	// FieldsWriter Writes a new segment
	FieldsWriter(state *SegmentWriteState) (PointsWriter, error)

	// FieldsReader Reads a segment. NOTE: by the time this call returns, it must hold open any files
	// it will need to use; else, those files may be deleted. Additionally, required files may be
	// deleted during the execution of this call before there is a chance to open them. Under these
	// circumstances an IOException should be thrown by the implementation. IOExceptions are expected
	// and will automatically cause a retry of the segment opening logic with the newly revised segments.
	FieldsReader(state *SegmentReadState) (PointsReader, error)
}
