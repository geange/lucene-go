package index

import (
	"github.com/bits-and-blooms/bitset"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util"
)

type PendingDeletes interface {
	GetMutableBits() *bitset.BitSet

	// Delete
	// Marks a document as deleted in this segment and return true if a document got actually deleted or
	// if the document was already deleted.
	Delete(docID int) (bool, error)

	// GetLiveDocs Returns a snapshot of the current live docs.
	GetLiveDocs() util.Bits

	// GetHardLiveDocs Returns a snapshot of the hard live docs.
	GetHardLiveDocs() util.Bits

	// NumPendingDeletes
	// Returns the number of pending deletes that are not written to disk.
	NumPendingDeletes() int

	// OnNewReader
	// Called once a new reader is opened for this segment ie. when deletes or updates are applied.
	OnNewReader(reader CodecReader, info *SegmentCommitInfo) error

	// DropChanges Resets the pending docs
	DropChanges()

	// WriteLiveDocs Writes the live docs to disk and returns true if any new docs were written.
	WriteLiveDocs(dir store.Directory) (bool, error)

	// IsFullyDeleted
	// Returns true iff the segment represented by this PendingDeletes is fully deleted
	IsFullyDeleted(readerIOSupplier func() CodecReader) (bool, error)

	// OnDocValuesUpdate Called for every field update for the given field at flush time
	// Params: 	info – the field info of the field that's updated
	//			iterator – the values to apply
	OnDocValuesUpdate(info *types.FieldInfo, iterator DocValuesFieldUpdatesIterator)

	GetDelCount() int
}

// PendingDeletesDefault
// This class handles accounting and applying pending deletes for live segment readers
type PendingDeletesDefault struct {
	info *SegmentCommitInfo

	// Read-only live docs, null until live docs are initialized or if all docs are alive
	liveDocs util.Bits

	// Writeable live docs, null if this instance is not ready to accept writes, in which
	// case getMutableBits needs to be called

	pendingDeleteCount  int
	liveDocsInitialized bool
}

func (p *PendingDeletesDefault) GetMutableBits() *bitset.BitSet {
	//TODO implement me
	panic("implement me")
}

func (p *PendingDeletesDefault) Delete(docID int) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PendingDeletesDefault) GetLiveDocs() util.Bits {
	//TODO implement me
	panic("implement me")
}

func (p *PendingDeletesDefault) GetHardLiveDocs() util.Bits {
	//TODO implement me
	panic("implement me")
}

func (p *PendingDeletesDefault) NumPendingDeletes() int {
	//TODO implement me
	panic("implement me")
}

func (p *PendingDeletesDefault) OnNewReader(reader CodecReader, info *SegmentCommitInfo) error {
	//TODO implement me
	panic("implement me")
}

func (p *PendingDeletesDefault) DropChanges() {
	//TODO implement me
	panic("implement me")
}

func (p *PendingDeletesDefault) WriteLiveDocs(dir store.Directory) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PendingDeletesDefault) IsFullyDeleted(readerIOSupplier func() CodecReader) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PendingDeletesDefault) OnDocValuesUpdate(info *types.FieldInfo, iterator DocValuesFieldUpdatesIterator) {
	//TODO implement me
	panic("implement me")
}

func (p *PendingDeletesDefault) GetDelCount() int {
	delCount := p.info.GetDelCount() + p.info.GetSoftDelCount() + p.NumPendingDeletes()
	return delCount
}

func NewPendingDeletes(reader *SegmentReader, info *SegmentCommitInfo) (*PendingDeletesDefault, error) {
	panic("")
}
