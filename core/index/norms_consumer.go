package index

import "github.com/geange/lucene-go/core/types"

// NormsConsumer Abstract API that consumes normalization values. Concrete implementations of this actually do "something" with the norms (write it into the index in a specific format).
// The lifecycle is:
// NormsConsumer is created by NormsFormat.normsConsumer(SegmentWriteState).
// addNormsField is called for each field with normalization values. The API is a "pull" rather than "push", and the implementation is free to iterate over the values multiple times (Iterable.iterator()).
// After all fields are added, the consumer is closed.
type NormsConsumer interface {
	NormsConsumerBase
	NormsConsumerExt
}

type NormsConsumerBase interface {
	// AddNormsField Writes normalization values for a field.
	//Params:
	//field – field information
	//normsProducer – NormsProducer of the numeric norm values
	//Throws:
	//IOException – if an I/O error occurred.
	AddNormsField(field *types.FieldInfo, normsProducer NormsProducer) error
}

type NormsConsumerExt interface {
	// Merge Merges in the fields from the readers in mergeState. The default implementation calls mergeNormsField for each field, filling segments with missing norms for the field with zeros. Implementations can override this method for more sophisticated merging (bulk-byte copying, etc).
	Merge(mergeState *MergeState) error

	// MergeNormsField Merges the norms from toMerge.
	//The default implementation calls addNormsField, passing an Iterable that merges and filters deleted documents on the fly.
	MergeNormsField(mergeFieldInfo *types.FieldInfo, mergeState *MergeState) error
}

var _ NormsConsumerExt = &NormsConsumerImp{}

type NormsConsumerImp struct {
	NormsConsumerBase
}

func (n *NormsConsumerImp) Merge(mergeState *MergeState) error {
	//TODO implement me
	panic("implement me")
}

func (n *NormsConsumerImp) MergeNormsField(mergeFieldInfo *types.FieldInfo, mergeState *MergeState) error {
	//TODO implement me
	panic("implement me")
}
