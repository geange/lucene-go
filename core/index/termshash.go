package index

import (
	"context"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/util/bytesref"
	"github.com/geange/lucene-go/core/util/ints"
)

// TermsHash
// This class is passed each token produced by the analyzer on each field during indexing,
// and it stores these tokens in a hash table, and allocates separate byte streams per token.
// Consumers of this class, eg FreqProxTermsWriter and TermVectorsConsumer, write their own byte
// streams under each term.
type TermsHash interface {
	Flush(ctx context.Context, fieldsToFlush map[string]TermsHashPerField, state *index.SegmentWriteState, sortMap index.DocMap, norms index.NormsProducer) error

	AddField(fieldInvertState *index.FieldInvertState, fieldInfo *document.FieldInfo) (TermsHashPerField, error)

	SetTermBytePool(termBytePool *bytesref.BlockPool)

	FinishDocument(ctx context.Context, docID int) error

	Abort() error

	Reset() error

	StartDocument() error

	GetIntPool() *ints.BlockPool
	GetBytePool() *bytesref.BlockPool
	GetTermBytePool() *bytesref.BlockPool
}

type BaseTermsHash struct {
	nextTermsHash TermsHash
	intPool       *ints.BlockPool
	bytePool      *bytesref.BlockPool
	termBytePool  *bytesref.BlockPool
}

func NewBaseTermsHash(intBlockAllocator ints.IntsAllocator, byteBlockAllocator bytesref.Allocator,
	nextTermsHash TermsHash) *BaseTermsHash {
	termHash := &BaseTermsHash{
		nextTermsHash: nextTermsHash,
		intPool:       ints.NewBlockPool(intBlockAllocator),
		bytePool:      bytesref.NewBlockPool(byteBlockAllocator),
	}

	if nextTermsHash != nil {
		termHash.termBytePool = termHash.bytePool
		nextTermsHash.SetTermBytePool(termHash.bytePool)
	}
	return termHash
}

func (h *BaseTermsHash) GetIntPool() *ints.BlockPool {
	return h.intPool
}

func (h *BaseTermsHash) GetBytePool() *bytesref.BlockPool {
	return h.bytePool
}

func (h *BaseTermsHash) GetTermBytePool() *bytesref.BlockPool {
	return h.termBytePool
}

func (h *BaseTermsHash) Flush(ctx context.Context, fieldsToFlush map[string]TermsHashPerField,
	state *index.SegmentWriteState, sortMap index.DocMap, norms index.NormsProducer) error {

	if h.nextTermsHash != nil {
		nextChildFields := make(map[string]TermsHashPerField)

		for k, v := range fieldsToFlush {
			nextChildFields[k] = v.GetNextPerField()
		}

		return h.nextTermsHash.Flush(ctx, nextChildFields, state, sortMap, norms)
	}
	return nil
}

func (h *BaseTermsHash) Abort() error {
	if err := h.Reset(); err != nil {
		return err
	}
	if h.nextTermsHash != nil {
		return h.nextTermsHash.Abort()
	}
	return nil
}

func (h *BaseTermsHash) Reset() error {
	h.intPool.Reset(false, false)
	h.bytePool.Reset(false, false)
	return nil
}

func (h *BaseTermsHash) FinishDocument(ctx context.Context, docID int) error {
	if h.nextTermsHash != nil {
		return h.nextTermsHash.FinishDocument(ctx, docID)
	}
	return nil
}

func (h *BaseTermsHash) StartDocument() error {
	if h.nextTermsHash != nil {
		return h.nextTermsHash.StartDocument()
	}
	return nil
}
