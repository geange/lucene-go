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
	if p.liveDocsInitialized == false {
		//assert writeableLiveDocs == null;
		if reader.HasDeletions() {
			// we only initialize this once either in the ctor or here
			// if we use the live docs from a reader it has to be in a situation where we don't
			// have any existing live docs
			//assert pendingDeleteCount == 0 : "pendingDeleteCount: " + pendingDeleteCount;
			p.liveDocs = reader.GetLiveDocs()
			//assert liveDocs == null || assertCheckLiveDocs(liveDocs, info.info.maxDoc(), info.getDelCount());
		}
		p.liveDocsInitialized = true
	}
	return nil
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

func NewPendingDeletes(reader *SegmentReader, info *SegmentCommitInfo) *PendingDeletesDefault {
	pd := NewPendingDeletesV2(info, reader.GetLiveDocs(), true)
	pd.pendingDeleteCount = reader.NumDeletedDocs() - info.GetDelCount()
	return pd
}

func NewPendingDeletesV1(info *SegmentCommitInfo) *PendingDeletesDefault {
	return NewPendingDeletesV2(info, nil, info.HasDeletions() == false)
	// if we don't have deletions we can mark it as initialized since we might receive deletes on a segment
	// without having a reader opened on it ie. after a merge when we apply the deletes that IW received while merging.
	// For segments that were published we enforce a reader in the BufferedUpdatesStream.SegmentState ctor
}

func NewPendingDeletesV2(info *SegmentCommitInfo, liveDocs util.Bits, liveDocsInitialized bool) *PendingDeletesDefault {
	return &PendingDeletesDefault{
		info:                info,
		liveDocs:            liveDocs,
		pendingDeleteCount:  0,
		liveDocsInitialized: liveDocsInitialized,
	}
}
