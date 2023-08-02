package index

import (
	"context"
	"github.com/geange/lucene-go/core/interface/index"
)

type BaseFieldsConsumer struct {

	// Merges in the fields from the readers in mergeState.
	// The default implementation skips and maps around deleted documents,
	// and calls write(Fields, NormsProducer). Implementations can override
	// this method for more sophisticated merging (bulk-byte copying, etc).
	// Write func(ctx context.Context, fields Fields, norms NormsProducer) error

	// NOTE: strange but necessary so javadocs linting is happy:
	// Closer func() error
}

func (f *BaseFieldsConsumer) Merge(ctx context.Context, mergeState *MergeState, norms index.NormsProducer) error {
	return nil
}

// MergeFromReaders
// Merges in the fields from the readers in mergeState. The default implementation skips and
// maps around deleted documents, and calls write(Fields, NormsProducer). Implementations can override
// this method for more sophisticated merging (bulk-byte copying, etc).
func MergeFromReaders(ctx context.Context, consumer index.FieldsConsumer, mergeState *MergeState, norms index.NormsProducer) error {
	return nil
}
