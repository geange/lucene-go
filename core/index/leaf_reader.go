package index

import (
	"github.com/geange/lucene-go/core/util"
)

type LeafReader interface {
	IndexReader

	// Terms Returns the Terms index for this field, or null if it has none.
	Terms(field string) (Terms, error)

	// Postings Returns PostingsEnum for the specified term. This will return null if either the field or
	// term does not exist.
	// NOTE: The returned PostingsEnum may contain deleted docs.
	// See Also: TermsEnum.postings(PostingsEnum)
	Postings(term *Term, flags int) (PostingsEnum, error)

	// GetNumericDocValues Returns NumericDocValues for this field, or null if no numeric doc values were
	// indexed for this field. The returned instance should only be used by a single thread.
	GetNumericDocValues(field string) (NumericDocValues, error)

	// GetBinaryDocValues Returns BinaryDocValues for this field, or null if no binary doc values were indexed
	// for this field. The returned instance should only be used by a single thread.
	GetBinaryDocValues(field string) (BinaryDocValues, error)

	// GetSortedDocValues Returns SortedDocValues for this field, or null if no SortedDocValues were indexed
	// for this field. The returned instance should only be used by a single thread.
	GetSortedDocValues(field string) (SortedDocValues, error)

	// GetSortedNumericDocValues Returns SortedNumericDocValues for this field, or null if no
	// SortedNumericDocValues were indexed for this field. The returned instance should only be used by a single thread.
	GetSortedNumericDocValues(field string) (SortedNumericDocValues, error)

	// GetSortedSetDocValues Returns SortedSetDocValues for this field, or null if no SortedSetDocValues
	// were indexed for this field. The returned instance should only be used by a single thread.
	GetSortedSetDocValues(field string) (SortedSetDocValues, error)

	// GetNormValues Returns NumericDocValues representing norms for this field, or null if no NumericDocValues
	// were indexed. The returned instance should only be used by a single thread.
	GetNormValues(field string) (NumericDocValues, error)

	// GetFieldInfos Get the FieldInfos describing all fields in this reader. Note: Implementations
	// should cache the FieldInfos instance returned by this method such that subsequent calls to
	// this method return the same instance.
	GetFieldInfos() *FieldInfos

	// GetLiveDocs Returns the Bits representing live (not deleted) docs. A set bit indicates the doc ID has
	// not been deleted. If this method returns null it means there are no deleted documents
	// (all documents are live). The returned instance has been safely published for use by multiple threads
	// without additional synchronization.
	GetLiveDocs() util.Bits

	// GetPointValues Returns the PointValues used for numeric or spatial searches for the given field, or null
	// if there are no point fields.
	GetPointValues(field string) (PointValues, error)

	// CheckIntegrity Checks consistency of this reader.
	// Note that this may be costly in terms of I/O, e.g. may involve computing a checksum value against large data files.
	CheckIntegrity() error

	// GetMetaData Return metadata about this leaf.
	GetMetaData() *LeafMetaData
}

type LeafReaderDefault struct {
	LeafReaderDefaultSpi

	readerContext *LeafReaderContext
	*IndexReaderDefault
}

type LeafReaderDefaultSpi interface {
	Terms(field string) (Terms, error)
}

type LeafReaderDefaultConfig struct {
	Terms         func(field string) (Terms, error)
	ReaderContext *LeafReaderContext
}

func NewLeafReaderDefault(reader LeafReader) *LeafReaderDefault {
	return &LeafReaderDefault{
		LeafReaderDefaultSpi: reader,
		readerContext:        NewLeafReaderContext(reader),
		IndexReaderDefault:   NewIndexReaderDefault(),
	}
}

func (r *LeafReaderDefault) Postings(term *Term, flags int) (PostingsEnum, error) {
	terms, err := r.Terms(term.Field())
	if err != nil {
		return nil, err
	}
	if terms == nil {
		return nil, nil
	}
	if terms != nil {
		termsEnum, err := terms.Iterator()
		if err != nil {
			return nil, err
		}

		if ok, err := termsEnum.SeekExact(term.Bytes()); err == nil && ok {
			return termsEnum.Postings(nil, flags)
		}
	}
	return nil, nil
}

func (r *LeafReaderDefault) GetContext() IndexReaderContext {
	return r.readerContext
}

func (r *LeafReaderDefault) DocFreq(term Term) (int, error) {
	terms, err := r.Terms(term.Field())
	if err != nil {
		return 0, err
	}
	if terms == nil {
		return 0, nil
	}

	termsEnum, err := terms.Iterator()
	if err != nil {
		return 0, err
	}
	if ok, err := termsEnum.SeekExact(term.Bytes()); err == nil && ok {
		return termsEnum.DocFreq()
	} else {
		return 0, err
	}
}

func (r *LeafReaderDefault) TotalTermFreq(term *Term) (int64, error) {
	terms, err := r.Terms(term.Field())
	if err != nil {
		return 0, err
	}
	if terms == nil {
		return 0, nil
	}

	termsEnum, err := terms.Iterator()
	if err != nil {
		return 0, err
	}
	if ok, err := termsEnum.SeekExact(term.Bytes()); err == nil && ok {
		return termsEnum.TotalTermFreq()
	} else {
		return 0, err
	}
}

func (r *LeafReaderDefault) GetSumDocFreq(field string) (int64, error) {
	terms, err := r.Terms(field)
	if err != nil {
		return 0, err
	}
	if terms == nil {
		return 0, nil
	}

	return terms.GetSumDocFreq()
}

func (r *LeafReaderDefault) GetDocCount(field string) (int, error) {
	terms, err := r.Terms(field)
	if err != nil {
		return 0, err
	}
	if terms == nil {
		return 0, nil
	}

	return terms.GetDocCount()
}

func (r *LeafReaderDefault) GetSumTotalTermFreq(field string) (int64, error) {
	terms, err := r.Terms(field)
	if err != nil {
		return 0, err
	}
	if terms == nil {
		return 0, nil
	}

	return terms.GetSumTotalTermFreq()
}
