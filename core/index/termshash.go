package index

import (
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util"
)

// TermsHash This class is passed each token produced by the analyzer on each field during indexing,
// and it stores these tokens in a hash table, and allocates separate byte streams per token.
// Consumers of this class, eg FreqProxTermsWriter and TermVectorsConsumer, write their own byte
// streams under each term.
type TermsHash struct {
	nextTermsHash *TermsHash
	intPool       *util.IntBlockPool
	bytePool      *util.ByteBlockPool
	termBytePool  *util.ByteBlockPool

	addField func(fieldInvertState *FieldInvertState, fieldInfo *types.FieldInfo)
}
