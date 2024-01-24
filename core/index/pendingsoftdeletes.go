package index

import (
	"errors"
	"github.com/bits-and-blooms/bitset"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util"
	"io"
)

var _ PendingDeletes = &PendingSoftDeletes{}

type PendingSoftDeletes struct {
	*PendingDeletesDefault

	field        string
	dvGeneration int64
	hardDeletes  PendingDeletes
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

func (p *PendingSoftDeletes) OnDocValuesUpdate(info *document.FieldInfo, iterator DocValuesFieldUpdatesIterator) {
	//TODO implement me
	panic("implement me")
}

func NewPendingSoftDeletes(field string, info *SegmentCommitInfo) *PendingSoftDeletes {

	return &PendingSoftDeletes{
		PendingDeletesDefault: NewPendingDeletesV2(info, nil, info.GetDelCountV1(true) == 0),
		field:                 field,
		dvGeneration:          -2,
		hardDeletes:           NewPendingDeletesV1(info),
	}
}

func NewPendingSoftDeletesV1(field string,
	reader *SegmentReader, info *SegmentCommitInfo) *PendingSoftDeletes {

	return &PendingSoftDeletes{
		PendingDeletesDefault: NewPendingDeletes(reader, info),
		field:                 field,
		dvGeneration:          -2,
		hardDeletes:           NewPendingDeletes(reader, info),
	}
}

/*
  static int countSoftDeletes(DocIdSetIterator softDeletedDocs, Bits hardDeletes) throws IOException {
    int count = 0;
    if (softDeletedDocs != null) {
      int doc;
      while ((doc = softDeletedDocs.nextDoc()) != DocIdSetIterator.NO_MORE_DOCS) {
        if (hardDeletes == null || hardDeletes.get(doc)) {
          count++;
        }
      }
    }
    return count;
  }
*/

func countSoftDeletes(softDeletedDocs types.DocIdSetIterator, hardDeletes util.Bits) (int, error) {
	count := 0
	if softDeletedDocs != nil {
		for {
			docId, err := softDeletedDocs.NextDoc()
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
