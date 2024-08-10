package index

import (
	"errors"

	"github.com/bits-and-blooms/bitset"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/types"
)

var _ index.DocIdSet = &DocsWithFieldSet{}

// DocsWithFieldSet
// Accumulator for documents that have a value for a field.
// This is optimized for the case that all documents have a value.
type DocsWithFieldSet struct {
	set       *bitset.BitSet
	cost      int
	lastDocId int
}

func NewDocsWithFieldSet() *DocsWithFieldSet {
	return &DocsWithFieldSet{
		set:       bitset.New(1),
		cost:      0,
		lastDocId: -1,
	}
}

func (d *DocsWithFieldSet) Iterator() (types.DocIdSetIterator, error) {
	return NewBitSetIterator(d.set, int64(d.cost)), nil
}

func (d *DocsWithFieldSet) Bits() (index.Bits, error) {
	return d.set, nil
}

func (d *DocsWithFieldSet) Add(docID int) error {
	if d.lastDocId >= docID {
		return errors.New("out of order doc ids")
	}
	d.lastDocId = docID
	d.set.Set(uint(docID))
	return nil
}
