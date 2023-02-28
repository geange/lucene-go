package index

import (
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util"
)

var _ TermsHash = &FreqProxTermsWriter{}

type FreqProxTermsWriter struct {
	*TermsHashDefault
}

func (f *FreqProxTermsWriter) AddField(fieldInvertState *FieldInvertState, fieldInfo *types.FieldInfo) (TermsHashPerField, error) {
	//TODO implement me
	panic("implement me")
}

func (f *FreqProxTermsWriter) SetTermBytePool(termBytePool *util.ByteBlockPool) {
	f.termBytePool = termBytePool
}
