package index

import (
	"errors"

	"github.com/bits-and-blooms/bitset"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util"
)

var _ DocIdSet = &DocsWithFieldSet{}

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

func (d *DocsWithFieldSet) Bits() (util.Bits, error) {
	return nil, nil
}

func (d *DocsWithFieldSet) Add(docID int) error {
	if d.lastDocId >= docID {
		return errors.New("out of order doc ids")
	}
	d.lastDocId = docID
	d.set.Set(uint(docID))
	return nil
}
