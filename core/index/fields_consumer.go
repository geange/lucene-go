package index

import (
	"io"
)

type FieldsConsumer interface {
	FieldsConsumerBase
	FieldsConsumerExt
}

type FieldsConsumerBase interface {
	io.Closer

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
	Write(fields Fields, norms NormsProducer) error
}

type FieldsConsumerExt interface {
	// Merge Merges in the fields from the readers in mergeState. The default implementation skips and
	// maps around deleted documents, and calls write(Fields, NormsProducer). Implementations can override
	// this method for more sophisticated merging (bulk-byte copying, etc).
	Merge(mergeState *MergeState, norms NormsProducer) error
}

var _ FieldsConsumerExt = &FieldsConsumerImp{}

type FieldsConsumerImp struct {
	FieldsConsumerBase
}

func (f *FieldsConsumerImp) Merge(mergeState *MergeState, norms NormsProducer) error {
	//TODO implement me
	panic("implement me")
}
