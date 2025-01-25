package index

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/bits-and-blooms/bitset"

	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util"
)

var _ PendingDeletes = &PendingSoftDeletes{}

type PendingSoftDeletes struct {
	*pendingDeletes

	field        string
	dvGeneration int64
	hardDeletes  PendingDeletes
}

func (p *PendingSoftDeletes) Delete(docID int) (bool, error) {
	// we need to fetch this first it might be a shared instance with hardDeletes
	mutableBits := p.GetMutableBits()

	deleteOk, err := p.hardDeletes.Delete(docID)
	if err != nil {
		return false, err
	}

	idx := uint(docID)
	if deleteOk {
		if mutableBits.Test(idx) { // delete it here too!
			mutableBits.Clear(idx)
		} else {
			// if it was deleted subtract the delCount
			p.pendingDeleteCount--
		}
		return true, nil
	}
	return false, nil
}

func (p *PendingSoftDeletes) GetHardLiveDocs() util.Bits {
	return p.hardDeletes.GetMutableBits()
}

func (p *PendingSoftDeletes) NumPendingDeletes() int {
	return p.pendingDeletes.NumPendingDeletes() + p.hardDeletes.NumPendingDeletes()
}

func (p *PendingSoftDeletes) OnNewReader(reader index.CodecReader, info index.SegmentCommitInfo) error {
	err := p.pendingDeletes.OnNewReader(reader, info)
	if err != nil {
		return err
	}

	err = p.hardDeletes.OnNewReader(reader, info)
	if err != nil {
		return err
	}

	if p.dvGeneration < info.GetDocValuesGen() { // only re-calculate this if we haven't seen this generation
		iterator, err := getDocValuesDocIdSetIterator(p.field, reader)
		if err != nil {
			return err
		}
		//var newDelCount int
		if iterator != nil { // nothing is deleted we don't have a soft deletes field in this segment
			_, err := applySoftDeletes(iterator, p.GetMutableBits())
			if err != nil {
				return err
			}
		} else {
			//newDelCount = 0
		}
		p.dvGeneration = info.GetDocValuesGen()
	}
	return nil
}

// Clears all bits in the given bitset that are set and are also in the given DocIdSetIterator.
func applySoftDeletes(iterator types.DocIdSetIterator, bits *bitset.BitSet) (int, error) {
	newDeletes := 0

	hasValue, ok := iterator.(DocValuesFieldUpdatesIterator)
	if !ok {
		return 0, nil
	}

	for hasValue.HasValue() {
		doc := hasValue.DocID()
		idx := uint(doc)

		if bits.Test(idx) {
			// doc is live - clear it
			bits.Clear(idx)
			newDeletes++
		} else {
			bits.Set(idx)
			newDeletes--
		}
	}

	return newDeletes, nil
}

func (p *PendingSoftDeletes) DropChanges() {
	// don't reset anything here - this is called after a merge (successful or not) to prevent
	// rewriting the deleted docs to disk. we only pass it on and reset the number of pending deletes
	p.hardDeletes.DropChanges()
}

func (p *PendingSoftDeletes) WriteLiveDocs(ctx context.Context, dir store.Directory) (bool, error) {
	// we need to set this here to make sure our stats in SCI are up-to-date otherwise we might hit an assertion
	// when the hard deletes are set since we need to account for docs that used to be only soft-delete but now hard-deleted
	p.info.SetSoftDelCount(p.info.GetSoftDelCount() + p.pendingDeleteCount)
	p.pendingDeletes.DropChanges()
	// delegate the write to the hard deletes - it will only write if somebody used it.
	ok, err := p.hardDeletes.WriteLiveDocs(ctx, dir)
	if err != nil {
		return false, err
	}
	if ok {
		return true, nil
	}
	return false, nil
}

func (p *PendingSoftDeletes) IsFullyDeleted(ctx context.Context, readerIOSupplier func() index.CodecReader) (bool, error) {
	err := p.ensureInitialized(ctx, readerIOSupplier)
	if err != nil {
		return false, err
	} // initialize to ensure we have accurate counts - only needed in the soft-delete case
	return p.pendingDeletes.IsFullyDeleted(ctx, readerIOSupplier)
}

func (p *PendingSoftDeletes) ensureInitialized(ctx context.Context, readerIOSupplier func() index.CodecReader) error {
	if p.dvGeneration == -2 {
		fieldInfos, err := p.readFieldInfos(ctx)
		if err != nil {
			return err
		}
		fieldInfo := fieldInfos.FieldInfo(p.field)
		// we try to only open a reader if it's really necessary ie. indices that are mainly append only might have
		// big segments that don't even have any docs in the soft deletes field. In such a case it's simply
		// enough to look at the FieldInfo for the field and check if the field has DocValues
		if fieldInfo != nil && fieldInfo.GetDocValuesType() != document.DOC_VALUES_TYPE_NONE {
			// in order to get accurate numbers we need to have a least one reader see here.
			err := p.OnNewReader(readerIOSupplier(), p.info)
			if err != nil {
				return err
			}
		} else {
			// we are safe here since we don't have any doc values for the soft-delete field on disk
			// no need to open a new reader
			p.dvGeneration = -1
			if fieldInfo != nil {
				p.dvGeneration = fieldInfo.GetDocValuesGen()
			}
		}
	}
	return nil
}

func (p *PendingSoftDeletes) OnDocValuesUpdate(info *document.FieldInfo, iterator DocValuesFieldUpdatesIterator) {
	if p.field == info.Name() {
		deletes, err := applySoftDeletes(iterator, p.GetMutableBits())
		if err != nil {
			return
		}
		p.pendingDeleteCount += deletes
		p.info.SetSoftDelCount(p.info.GetSoftDelCount() + p.pendingDeleteCount)
		p.pendingDeletes.DropChanges()
	}
	p.dvGeneration = info.GetDocValuesGen()
}

func NewPendingSoftDeletes(field string, info index.SegmentCommitInfo) *PendingSoftDeletes {

	return &PendingSoftDeletes{
		pendingDeletes: NewPendingDeletesV2(info, nil, info.GetDelCountWithSoftDeletes(true) == 0).(*pendingDeletes),
		field:          field,
		dvGeneration:   -2,
		hardDeletes:    NewPendingDeletesV1(info),
	}
}

func NewPendingSoftDeletesV1(field string,
	reader *SegmentReader, info index.SegmentCommitInfo) *PendingSoftDeletes {

	return &PendingSoftDeletes{
		pendingDeletes: NewPendingDeletes(reader, info).(*pendingDeletes),
		field:          field,
		dvGeneration:   -2,
		hardDeletes:    NewPendingDeletes(reader, info),
	}
}

func countSoftDeletes(softDeletedDocs types.DocIdSetIterator, hardDeletes util.Bits) (int, error) {
	count := 0
	if softDeletedDocs != nil {
		for {
			docId, err := softDeletedDocs.NextDoc(context.Background())
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return 0, err
			}
			if hardDeletes != nil && hardDeletes.Test(uint(docId)) {
				count++
			}
		}
	}
	return count, nil
}

func (p *PendingSoftDeletes) MustInitOnDelete() bool {
	return p.liveDocsInitialized == false
}

func (p *PendingSoftDeletes) readFieldInfos(ctx context.Context) (index.FieldInfos, error) {
	segInfo := p.info.Info()
	dir := segInfo.Dir()

	var err error

	if p.info.HasFieldUpdates() == false {
		// updates always outside of CFS
		if segInfo.GetUseCompoundFile() {
			dir, err = segInfo.GetCodec().CompoundFormat().GetCompoundReader(ctx, segInfo.Dir(), segInfo, store.READONCE)
			if err != nil {
				return nil, err
			}
			defer dir.Close()
		} else {
			dir = segInfo.Dir()
		}

		return segInfo.GetCodec().FieldInfosFormat().Read(ctx, dir, segInfo, "", store.READONCE)
	}

	fisFormat := segInfo.GetCodec().FieldInfosFormat()
	segmentSuffix := fmt.Sprint(p.info.GetFieldInfosGen())
	return fisFormat.Read(ctx, dir, segInfo, segmentSuffix, store.READONCE)
}
