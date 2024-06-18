package index

import (
	"context"
	"github.com/geange/lucene-go/core/interface/index"
	"io"
)

type FieldsConsumer interface {
	io.Closer

	// Write
	// Write all fields, terms and postings. This the "pull" API, allowing you to iterate more than once
	// over the postings, somewhat analogous to using a DOM API to traverse an XML tree.
	// Notes:
	// * You must compute index statistics, including each Term's docFreq and totalTermFreq, as well as the
	//	 summary sumTotalTermFreq, sumTotalDocFreq and docCount.
	// * You must skip terms that have no docs and fields that have no terms, even though the provided Fields
	//	 API will expose them; this typically requires lazily writing the field or term until you've actually
	//	 seen the first term or document.
	// * The provided Fields instance is limited: you cannot call any methods that return statistics/counts;
	//	 you cannot pass a non-null live docs when pulling docs/positions enums.
	Write(ctx context.Context, fields index.Fields, norms NormsProducer) error

	// Merge
	// Merges in the fields from the readers in mergeState. The default implementation skips and
	// maps around deleted documents, and calls write(Fields, NormsProducer). Implementations can override
	// this method for more sophisticated merging (bulk-byte copying, etc).
	Merge(ctx context.Context, mergeState *MergeState, norms NormsProducer) error
}

type BaseFieldsConsumer struct {

	// Merges in the fields from the readers in mergeState.
	// The default implementation skips and maps around deleted documents,
	// and calls write(Fields, NormsProducer). Implementations can override
	// this method for more sophisticated merging (bulk-byte copying, etc).
	// Write func(ctx context.Context, fields Fields, norms NormsProducer) error

	// NOTE: strange but necessary so javadocs linting is happy:
	// Closer func() error
}

// Merge
// Merges in the fields from the readers in mergeState. The default implementation skips and
// maps around deleted documents, and calls write(Fields, NormsProducer). Implementations can override
// this method for more sophisticated merging (bulk-byte copying, etc).
func (f *BaseFieldsConsumer) Merge(ctx context.Context, mergeState *MergeState, norms NormsProducer) error {
	return nil
}

// FieldsProducer Sole constructor. (For invocation by subclass constructors, typically implicit.)
type FieldsProducer interface {
	io.Closer

	index.Fields

	// CheckIntegrity
	// Checks consistency of this reader.
	// Note that this may be costly in terms of I/O, e.g. may involve computing a checksum item against large
	// data files.
	CheckIntegrity() error

	// GetMergeInstance Returns an instance optimized for merging. This instance may only be consumed in the
	// thread that called getMergeInstance().
	// The default implementation returns this
	GetMergeInstance() FieldsProducer
}
