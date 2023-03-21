package index

import (
	"github.com/bits-and-blooms/bitset"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util"
)

var _ PendingDeletes = &PendingSoftDeletes{}

type PendingSoftDeletes struct {
}

func (p *PendingSoftDeletes) GetDelCount() int {
	//TODO implement me
	panic("implement me")
}

func (p *PendingSoftDeletes) GetMutableBits() *bitset.BitSet {
	//TODO implement me
	panic("implement me")
}

func (p *PendingSoftDeletes) Delete(docID int) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PendingSoftDeletes) GetLiveDocs() util.Bits {
	//TODO implement me
	panic("implement me")
}

func (p *PendingSoftDeletes) GetHardLiveDocs() util.Bits {
	//TODO implement me
	panic("implement me")
}

func (p *PendingSoftDeletes) NumPendingDeletes() int {
	//TODO implement me
	panic("implement me")
}

func (p *PendingSoftDeletes) OnNewReader(reader CodecReader, info *SegmentCommitInfo) error {
	//TODO implement me
	panic("implement me")
}

func (p *PendingSoftDeletes) DropChanges() {
	//TODO implement me
	panic("implement me")
}

func (p *PendingSoftDeletes) WriteLiveDocs(dir store.Directory) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PendingSoftDeletes) IsFullyDeleted(readerIOSupplier func() CodecReader) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PendingSoftDeletes) OnDocValuesUpdate(info *types.FieldInfo, iterator DocValuesFieldUpdatesIterator) {
	//TODO implement me
	panic("implement me")
}

func NewPendingSoftDeletes(field string,
	reader *SegmentReader, info *SegmentCommitInfo) (*PendingSoftDeletes, error) {

	panic("")
}
