package index

import (
	"context"

	"github.com/bits-and-blooms/bitset"

	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util"
)

type PendingDeletes interface {
	GetMutableBits() *bitset.BitSet

	// Delete
	// Marks a document as deleted in this segment and return true if a document got actually deleted or
	// if the document was already deleted.
	Delete(docID int) (bool, error)

	// GetLiveDocs
	// Returns a snapshot of the current live docs.
	GetLiveDocs() util.Bits

	// GetHardLiveDocs
	// Returns a snapshot of the hard live docs.
	GetHardLiveDocs() util.Bits

	// NumPendingDeletes
	// Returns the number of pending deletes that are not written to disk.
	NumPendingDeletes() int

	// OnNewReader
	// Called once a new reader is opened for this segment ie. when deletes or updates are applied.
	OnNewReader(reader index.CodecReader, info index.SegmentCommitInfo) error

	// DropChanges
	// Resets the pending docs
	DropChanges()

	// WriteLiveDocs
	// Writes the live docs to disk and returns true if any new docs were written.
	WriteLiveDocs(ctx context.Context, dir store.Directory) (bool, error)

	// IsFullyDeleted
	// Returns true iff the segment represented by this PendingDeletes is fully deleted
	IsFullyDeleted(ctx context.Context, readerIOSupplier func() index.CodecReader) (bool, error)

	// OnDocValuesUpdate
	// Called for every field update for the given field at flush time
	// info: the field info of the field that's updated
	// iterator: the values to apply
	OnDocValuesUpdate(info *document.FieldInfo, iterator DocValuesFieldUpdatesIterator)

	// NeedsRefresh
	// Returns true if the given reader needs to be refreshed in order to see the latest deletes
	NeedsRefresh(reader index.CodecReader) bool

	// GetDelCount
	// Returns the number of deleted docs in the segment.
	GetDelCount() int

	// NumDocs
	// Returns the number of live documents in this segment
	NumDocs() (int, error)

	// MustInitOnDelete
	// Returns true if we have to initialize this PendingDeletes before delete(int);
	// otherwise this PendingDeletes is ready to accept deletes. A PendingDeletes can
	// be initialized by providing it a reader via onNewReader(CodecReader, SegmentCommitInfo).
	MustInitOnDelete() bool
}

// pendingDeletes
// This class handles accounting and applying pending deletes for live segment readers
type pendingDeletes struct {
	info index.SegmentCommitInfo

	// Read-only live docs, null until live docs are initialized or if all docs are alive
	liveDocs util.Bits

	// Writeable live docs, null if this instance is not ready to accept writes, in which
	// case getMutableBits needs to be called
	writeableLiveDocs *bitset.BitSet

	// Writeable live docs, null if this instance is not ready to accept writes, in which
	// case getMutableBits needs to be called

	pendingDeleteCount  int
	liveDocsInitialized bool
}

func (p *pendingDeletes) NeedsRefresh(reader index.CodecReader) bool {
	return reader.GetLiveDocs() != p.GetLiveDocs() || reader.NumDeletedDocs() != p.GetDelCount()
}

func (p *pendingDeletes) NumDocs() (int, error) {
	maxDoc, err := p.info.Info().MaxDoc()
	if err != nil {
		return 0, err
	}
	delCount := p.GetDelCount()
	return maxDoc - delCount, nil
}

func (p *pendingDeletes) MustInitOnDelete() bool {
	return false
}

func (p *pendingDeletes) GetMutableBits() *bitset.BitSet {
	// if we pull mutable bits but we haven't been initialized something is completely off.
	// this means we receive deletes without having the bitset that is on-disk ready to be cloned

	if p.writeableLiveDocs == nil {
		// Copy on write: this means we've cloned a
		// SegmentReader sharing the current liveDocs
		// instance; must now make a private clone so we can
		// change it:
		if p.liveDocs != nil {
			p.writeableLiveDocs = p.liveDocs.(*bitset.BitSet).Clone()
		} else {
			doc, _ := p.info.Info().MaxDoc()
			p.writeableLiveDocs = bitset.New(uint(doc))
			p.writeableLiveDocs.FlipRange(0, uint(doc))
		}
		p.liveDocs = p.writeableLiveDocs
	}
	return p.writeableLiveDocs
}

func (p *pendingDeletes) Delete(docID int) (bool, error) {
	mutableBits := p.GetMutableBits()

	i := uint(docID)
	didDelete := mutableBits.Test(i)
	if didDelete {
		mutableBits.Clear(i)
		p.pendingDeleteCount++
	}
	return didDelete, nil
}

func (p *pendingDeletes) GetLiveDocs() util.Bits {
	return p.liveDocs
}

func (p *pendingDeletes) GetHardLiveDocs() util.Bits {
	return p.GetLiveDocs()
}

func (p *pendingDeletes) NumPendingDeletes() int {
	return p.pendingDeleteCount
}

func (p *pendingDeletes) OnNewReader(reader index.CodecReader, info index.SegmentCommitInfo) error {
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

func (p *pendingDeletes) DropChanges() {
	p.pendingDeleteCount = 0
}

func (p *pendingDeletes) WriteLiveDocs(ctx context.Context, dir store.Directory) (bool, error) {
	if p.pendingDeleteCount == 0 {
		return false, nil
	}

	liveDocs := p.liveDocs

	// Do this so we can delete any created files on
	// exception; this saves all codecs from having to do
	// it:
	trackingDir := store.NewTrackingDirectoryWrapper(dir)

	codec := p.info.Info().GetCodec()
	err := codec.LiveDocsFormat().WriteLiveDocs(ctx, liveDocs, trackingDir, p.info, p.pendingDeleteCount, store.DEFAULT)
	if err != nil {
		return false, err
	}

	// If we hit an exc in the line above (eg disk full)
	// then info's delGen remains pointing to the previous
	// (successfully written) del docs:
	p.info.AdvanceDelGen()
	p.info.SetDelCount(p.info.GetDelCount() + p.pendingDeleteCount)
	p.DropChanges()
	return true, nil
}

func (p *pendingDeletes) IsFullyDeleted(ctx context.Context, readerIOSupplier func() index.CodecReader) (bool, error) {
	delCount := p.GetDelCount()
	maxDoc, err := p.info.Info().MaxDoc()
	if err != nil {
		return false, err
	}
	return delCount == maxDoc, nil
}

func (p *pendingDeletes) OnDocValuesUpdate(info *document.FieldInfo, iterator DocValuesFieldUpdatesIterator) {
	return
}

func (p *pendingDeletes) GetDelCount() int {
	delCount := p.info.GetDelCount() + p.info.GetSoftDelCount() + p.NumPendingDeletes()
	return delCount
}

func NewPendingDeletes(reader *SegmentReader, info index.SegmentCommitInfo) PendingDeletes {
	pd := NewPendingDeletesV2(info, reader.GetLiveDocs(), true).(*pendingDeletes)
	pd.pendingDeleteCount = reader.NumDeletedDocs() - info.GetDelCount()
	return pd
}

func NewPendingDeletesV1(info index.SegmentCommitInfo) PendingDeletes {
	return NewPendingDeletesV2(info, nil, info.HasDeletions() == false)
	// if we don't have deletions we can mark it as initialized since we might receive deletes on a segment
	// without having a reader opened on it ie. after a merge when we apply the deletes that IW received while merging.
	// For segments that were published we enforce a reader in the BufferedUpdatesStream.SegmentState ctor
}

func NewPendingDeletesV2(info index.SegmentCommitInfo, liveDocs util.Bits, liveDocsInitialized bool) PendingDeletes {
	return &pendingDeletes{
		info:                info,
		liveDocs:            liveDocs,
		pendingDeleteCount:  0,
		liveDocsInitialized: liveDocsInitialized,
	}
}
