package search

import "fmt"

// CollectionStatistics Contains statistics for a collection (field).
// This class holds statistics across all documents for scoring purposes:
// * maxDoc(): number of documents.
// * docCount(): number of documents that contain this field.
// * sumDocFreq(): number of postings-list entries.
// * sumTotalTermFreq(): number of tokens.
// The following conditions are always true:
// All statistics are positive integers: never zero or negative.
// docCount <= maxDoc
// docCount <= sumDocFreq <= sumTotalTermFreq
// Values may include statistics on deleted documents that have not yet been merged away.
// Be careful when performing calculations on these values because they are represented as 64-bit integer
// values, you may need to cast to double for your use.
type CollectionStatistics struct {
	field            string
	maxDoc           int64
	docCount         int64
	sumTotalTermFreq int64
	sumDocFreq       int64
}

func NewCollectionStatistics(field string,
	maxDoc, docCount, sumTotalTermFreq, sumDocFreq int64) (*CollectionStatistics, error) {

	if maxDoc <= 0 {
		return nil, fmt.Errorf("maxDoc must be positive, maxDoc: %d", maxDoc)
	}

	if docCount <= 0 {
		return nil, fmt.Errorf("docCount must be positive, docCount: %d", docCount)
	}

	if docCount > maxDoc {
		return nil, fmt.Errorf("docCount must not exceed maxDoc, docCount: %d, maxDoc: %d", docCount, maxDoc)
	}

	if sumDocFreq <= 0 {
		return nil, fmt.Errorf("sumDocFreq must be positive, sumDocFreq: %d", sumDocFreq)
	}

	if sumDocFreq < docCount {
		return nil, fmt.Errorf(
			"sumDocFreq must be at least docCount, sumDocFreq: %d, docCount: %d", sumDocFreq, docCount)
	}

	if sumTotalTermFreq <= 0 {
		return nil, fmt.Errorf("sumTotalTermFreq must be positive, sumTotalTermFreq: %d", sumTotalTermFreq)
	}

	if sumTotalTermFreq < sumDocFreq {
		return nil, fmt.Errorf(
			"sumTotalTermFreq must be at least sumDocFreq, sumTotalTermFreq: %d, sumDocFreq: %d",
			sumTotalTermFreq, sumDocFreq)
	}

	return &CollectionStatistics{
		field:            field,
		maxDoc:           maxDoc,
		docCount:         docCount,
		sumTotalTermFreq: sumTotalTermFreq,
		sumDocFreq:       sumDocFreq,
	}, nil
}

func (c *CollectionStatistics) Field() string {
	return c.field
}

func (c *CollectionStatistics) MaxDoc() int64 {
	return c.maxDoc
}

func (c *CollectionStatistics) DocCount() int64 {
	return c.docCount
}

func (c *CollectionStatistics) SumTotalTermFreq() int64 {
	return c.sumTotalTermFreq
}

func (c *CollectionStatistics) SumDocFreq() int64 {
	return c.sumDocFreq
}
