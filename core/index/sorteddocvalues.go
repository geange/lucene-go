package index

import (
	"bytes"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util/automaton"
)

type SortedDocValuesDefaultConfig struct {
	OrdValue      func() (int, error)
	LookupOrd     func(ord int) ([]byte, error)
	GetValueCount func() int
}

type BaseSortedDocValues struct {
	BaseBinaryDocValues

	FnOrdValue      func() (int, error)
	FnLookupOrd     func(ord int) ([]byte, error)
	FnGetValueCount func() int
}

func NewBaseSortedDocValues(cfg *SortedDocValuesDefaultConfig) *BaseSortedDocValues {
	return &BaseSortedDocValues{
		FnOrdValue:      cfg.OrdValue,
		FnLookupOrd:     cfg.LookupOrd,
		FnGetValueCount: cfg.GetValueCount,
	}
}

func (r *BaseSortedDocValues) BinaryValue() ([]byte, error) {
	ord, err := r.FnOrdValue()
	if err != nil {
		return nil, err
	}

	if ord == -1 {
		return []byte{}, nil
	}
	return r.FnLookupOrd(ord)
}

func (r *BaseSortedDocValues) LookupTerm(key []byte) (int, error) {
	low := 0
	high := r.FnGetValueCount() - 1

	for low <= high {
		mid := (low + high) >> 1
		term, err := r.FnLookupOrd(mid)
		if err != nil {
			return 0, err
		}

		cmp := bytes.Compare(term, key)

		if cmp < 0 {
			low = mid + 1
		} else if cmp > 0 {
			high = mid - 1
		} else {
			return mid, nil
		}
	}

	return -(low + 1), nil // key not found.
}

func (r *BaseSortedDocValues) Intersect(automaton *automaton.CompiledAutomaton) (index.TermsEnum, error) {
	//TODO implement me
	panic("implement me")
}

var _ DocValuesWriter = &SortedDocValuesWriter{}

type SortedDocValuesWriter struct {
}

func (s *SortedDocValuesWriter) Flush(state *index.SegmentWriteState, sortMap index.DocMap, consumer index.DocValuesConsumer) error {
	//TODO implement me
	panic("implement me")
}

func (s *SortedDocValuesWriter) GetDocValues() types.DocIdSetIterator {
	//TODO implement me
	panic("implement me")
}
