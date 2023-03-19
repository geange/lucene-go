package index

import (
	"errors"
	"fmt"
	"sort"

	"github.com/geange/lucene-go/core/document"
)

var _ CompositeReader = &BaseCompositeReader{}

type BaseCompositeReader struct {
	*IndexReaderDefault

	subReaders []IndexReader

	subReadersSorter func(a, b IndexReader) int
	starts           []int // 1st docno for each reader
	maxDoc           int
	numDocs          int // computed lazily

	// List view solely for getSequentialSubReaders(), for effectiveness the array is used internally.
	subReadersList []IndexReader
}

func (b *BaseCompositeReader) GetReaderCacheHelper() CacheHelper {
	//TODO implement me
	panic("implement me")
}

func (b *BaseCompositeReader) GetSequentialSubReaders() []IndexReader {
	//TODO implement me
	panic("implement me")
}

func NewBaseCompositeReader(subReaders []IndexReader,
	subReadersSorter func(a, b IndexReader) int) (*BaseCompositeReader, error) {

	sort.Sort(&IndexReaderSorter{
		Readers:   subReaders,
		FnCompare: subReadersSorter,
	})

	reader := &BaseCompositeReader{
		subReaders:       subReaders,
		subReadersSorter: subReadersSorter,
		starts:           make([]int, len(subReaders)+1),
		maxDoc:           0,
		numDocs:          0,
		subReadersList:   subReaders,
	}

	maxDoc := 0
	for i := 0; i < len(subReaders); i++ {
		reader.starts[i] = maxDoc
		r := subReaders[i]
		maxDoc += r.MaxDoc() // compute maxDocs
		//r.RegisterParentReader(reader)
	}

	if maxDoc > GetActualMaxDocs() {
		return nil, errors.New("too many documents")
	}

	reader.maxDoc = maxDoc
	reader.starts[len(subReaders)] = maxDoc

	return reader, nil
}

func (b *BaseCompositeReader) GetTermVectors(docID int) (Fields, error) {
	//ensureOpen();
	i, err := b.readerIndex(docID) // find subreader num
	if err != nil {
		return nil, err
	}
	return b.subReaders[i].GetTermVectors(docID - b.starts[i]) // dispatch to subreader
}

func (b *BaseCompositeReader) NumDocs() int {
	// Don't call ensureOpen() here (it could affect performance)
	// We want to compute numDocs() lazily so that creating a wrapper that hides
	// some documents isn't slow at wrapping time, but on the first time that
	// numDocs() is called. This can help as there are lots of use-cases of a
	// reader that don't involve calling numDocs().
	// However it's not crucial to make sure that we don't call numDocs() more
	// than once on the sub readers, since they likely cache numDocs() anyway,
	// hence the lack of synchronization.
	numDocs := b.numDocs
	if numDocs == -1 {
		numDocs = 0
		for _, r := range b.subReaders {
			numDocs += r.NumDocs()
		}
		//assert numDocs >= 0;
		b.numDocs = numDocs
	}
	return numDocs
}

func (b *BaseCompositeReader) MaxDoc() int {
	// Don't call ensureOpen() here (it could affect performance)
	return b.maxDoc
}

func (b *BaseCompositeReader) DocumentV1(docID int, visitor document.StoredFieldVisitor) error {
	//ensureOpen();
	i, err := b.readerIndex(docID) // find subreader num
	if err != nil {
		return err
	}
	return b.subReaders[i].DocumentV1(docID-b.starts[i], visitor) // dispatch to subreader
}

func (b *BaseCompositeReader) DocFreq(term Term) (int, error) {
	//ensureOpen();
	total := 0 // sum freqs in subreaders
	for i := 0; i < len(b.subReaders); i++ {
		sub, err := b.subReaders[i].DocFreq(term)
		if err != nil {
			return 0, err
		}
		//assert sub >= 0;
		//assert sub <= subReaders[i].getDocCount(term.field());
		total += sub
	}
	return total, nil
}

func (b *BaseCompositeReader) TotalTermFreq(term *Term) (int64, error) {
	//ensureOpen();
	total := int64(0) // sum freqs in subreaders
	for i := 0; i < len(b.subReaders); i++ {
		sub, err := b.subReaders[i].TotalTermFreq(term)
		if err != nil {
			return 0, err
		}
		//assert sub >= 0;
		//assert sub <= subReaders[i].getSumTotalTermFreq(term.field());
		total += sub
	}
	return total, nil
}

func (b *BaseCompositeReader) GetSumDocFreq(field string) (int64, error) {
	//ensureOpen();
	total := int64(0) // sum doc freqs in subreaders
	for _, reader := range b.subReaders {
		sub, err := reader.GetSumDocFreq(field)
		if err != nil {
			return 0, err
		}
		//assert sub >= 0;
		//assert sub <= reader.getSumTotalTermFreq(field);
		total += sub
	}
	return total, nil
}

func (b *BaseCompositeReader) GetDocCount(field string) (int, error) {
	//ensureOpen();
	total := 0 // sum doc counts in subreaders
	for _, reader := range b.subReaders {
		sub, err := reader.GetDocCount(field)
		if err != nil {
			return 0, err
		}
		//assert sub >= 0;
		//assert sub <= reader.maxDoc();
		total += sub
	}
	return total, nil
}

func (b *BaseCompositeReader) GetSumTotalTermFreq(field string) (int64, error) {
	//ensureOpen();
	total := int64(0) // sum doc total term freqs in subreaders
	for _, reader := range b.subReaders {
		sub, err := reader.GetSumTotalTermFreq(field)
		if err != nil {
			return 0, err
		}
		//assert sub >= 0;
		//assert sub >= reader.getSumDocFreq(field);
		total += sub
	}
	return total, nil
}

func (b *BaseCompositeReader) readerIndex(docID int) (int, error) {
	if docID < 0 || docID >= b.maxDoc {
		return 0, fmt.Errorf("docID must be >= 0 and < maxDoc=%d (got docID=%d)", b.maxDoc, docID)
	}
	return SubIndex(docID, b.starts), nil
}
