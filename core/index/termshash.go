package index

import (
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/util/bytesutils"
	"github.com/geange/lucene-go/core/util/ints"
)

// TermsHash This class is passed each token produced by the analyzer on each field during indexing,
// and it stores these tokens in a hash table, and allocates separate byte streams per token.
// Consumers of this class, eg FreqProxTermsWriter and TermVectorsConsumer, write their own byte
// streams under each term.
type TermsHash interface {
	Flush(fieldsToFlush map[string]TermsHashPerField, state *SegmentWriteState,
		sortMap *DocMap, norms NormsProducer) error

	AddField(fieldInvertState *FieldInvertState, fieldInfo *document.FieldInfo) (TermsHashPerField, error)

	SetTermBytePool(termBytePool *bytesutils.BlockPool)

	FinishDocument(docID int) error

	Abort() error

	Reset() error

	StartDocument() error

	GetIntPool() *ints.BlockPool
	GetBytePool() *bytesutils.BlockPool
	GetTermBytePool() *bytesutils.BlockPool
}

type TermsHashDefault struct {
	nextTermsHash TermsHash
	intPool       *ints.BlockPool
	bytePool      *bytesutils.BlockPool
	termBytePool  *bytesutils.BlockPool
}

func NewTermsHashDefault(intBlockAllocator ints.IntsAllocator, byteBlockAllocator bytesutils.Allocator,
	nextTermsHash TermsHash) *TermsHashDefault {
	termHash := &TermsHashDefault{
		nextTermsHash: nextTermsHash,
		intPool:       ints.NewBlockPool(intBlockAllocator),
		bytePool:      bytesutils.NewBlockPool(byteBlockAllocator),
	}

	if nextTermsHash != nil {
		termHash.termBytePool = termHash.bytePool
		nextTermsHash.SetTermBytePool(termHash.bytePool)
	}
	return termHash
}

func (h *TermsHashDefault) GetIntPool() *ints.BlockPool {
	return h.intPool
}

func (h *TermsHashDefault) GetBytePool() *bytesutils.BlockPool {
	return h.bytePool
}

func (h *TermsHashDefault) GetTermBytePool() *bytesutils.BlockPool {
	return h.termBytePool
}

func (h *TermsHashDefault) Flush(fieldsToFlush map[string]TermsHashPerField,
	state *SegmentWriteState, sortMap *DocMap, norms NormsProducer) error {

	if h.nextTermsHash != nil {
		nextChildFields := make(map[string]TermsHashPerField)

		for k, v := range fieldsToFlush {
			nextChildFields[k] = v.GetNextPerField()
		}

		return h.nextTermsHash.Flush(nextChildFields, state, sortMap, norms)
	}
	return nil
}

func (h *TermsHashDefault) Abort() error {
	h.Reset()
	if h.nextTermsHash != nil {
		return h.nextTermsHash.Abort()
	}
	return nil
}

func (h *TermsHashDefault) Reset() error {
	h.intPool.Reset(false, false)
	h.bytePool.Reset(false, false)
	return nil
}

func (h *TermsHashDefault) FinishDocument(docID int) error {
	if h.nextTermsHash != nil {
		return h.nextTermsHash.FinishDocument(docID)
	}
	return nil
}

func (h *TermsHashDefault) StartDocument() error {
	if h.nextTermsHash != nil {
		return h.nextTermsHash.StartDocument()
	}
	return nil
}
