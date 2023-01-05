package index

// NormsFormat Encodes/decodes per-document score normalization values.
type NormsFormat interface {

	// NormsConsumer Returns a NormsConsumer to write norms to the index.
	NormsConsumer(state *SegmentWriteState) (NormsConsumer, error)

	// NormsProducer Returns a NormsProducer to read norms from the index.
	// NOTE: by the time this call returns, it must hold open any files it will need to use; else,
	// those files may be deleted. Additionally, required files may be deleted during the execution
	// of this call before there is a chance to open them. Under these circumstances an IOException
	// should be thrown by the implementation. IOExceptions are expected and will automatically
	// cause a retry of the segment opening logic with the newly revised segments.
	NormsProducer(state *SegmentReadState) (NormsProducer, error)
}
