package types

import (
	"errors"
	"fmt"
)

// TermStatistics Contains statistics for a specific term
// This class holds statistics for this term across all documents for scoring purposes:
// docFreq: number of documents this term occurs in.
// totalTermFreq: number of tokens for this term.
// The following conditions are always true:
// All statistics are positive integers: never zero or negative.
// docFreq <= totalTermFreq
// docFreq <= sumDocFreq of the collection
// totalTermFreq <= sumTotalTermFreq of the collection
// Values may include statistics on deleted documents that have not yet been merged away.
// Be careful when performing calculations on these values because they are represented as 64-bit integer
// values, you may need to cast to double for your use.
type TermStatistics struct {
	term          []byte
	docFreq       int64
	totalTermFreq int64
}

// NewTermStatistics Creates statistics instance for a term.
// Params: 	term – Term bytes
//
//	docFreq – number of documents containing the term in the collection.
//	totalTermFreq – number of occurrences of the term in the collection.
//
// Throws: 	NullPointerException – if term is null.
//
//	IllegalArgumentException – if docFreq is negative or zero.
//	IllegalArgumentException – if totalTermFreq is less than docFreq.
func NewTermStatistics(term []byte, docFreq, totalTermFreq int64) (*TermStatistics, error) {
	if len(term) == 0 {
		return nil, errors.New("term require not nil")
	}

	if docFreq <= 0 {
		return nil, fmt.Errorf("docFreq must be positive, docFreq: %d", docFreq)
	}

	if totalTermFreq <= 0 {
		return nil, fmt.Errorf("totalTermFreq must be positive, totalTermFreq: %d", totalTermFreq)
	}

	if totalTermFreq < docFreq {
		return nil, fmt.Errorf("totalTermFreq must be at least docFreq, totalTermFreq: %d, docFreq: %d",
			totalTermFreq, docFreq)
	}

	return &TermStatistics{
		term:          term,
		docFreq:       docFreq,
		totalTermFreq: totalTermFreq,
	}, nil
}

func (t *TermStatistics) Term() []byte {
	return t.term
}

func (t *TermStatistics) DocFreq() int64 {
	return t.docFreq
}

func (t *TermStatistics) TotalTermFreq() int64 {
	return t.totalTermFreq
}
