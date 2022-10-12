package fst

import (
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util"
)

type FixedLengthArcsBuffer struct {
	bytes []byte
	bado  *store.ByteArrayDataOutput
}

func (f *FixedLengthArcsBuffer) ensureCapacity(capacity int) *FixedLengthArcsBuffer {
	if len(f.bytes) < capacity {
		f.bytes = make([]byte, util.Oversize(capacity, ByteSize))
		f.bado.Reset(f.bytes)
	}
	return f
}

func (f *FixedLengthArcsBuffer) resetPosition() *FixedLengthArcsBuffer {
	f.bado.Reset(f.bytes)
	return f
}

func (f *FixedLengthArcsBuffer) writeByte(b byte) *FixedLengthArcsBuffer {
	f.bado.WriteByte(b)
	return f
}

func (f *FixedLengthArcsBuffer) writeVInt(i int) *FixedLengthArcsBuffer {
	f.bado.WriteUvarint(uint64(i))
	return f
}

func (f *FixedLengthArcsBuffer) getPosition() int {
	return f.bado.GetPosition()
}

func (f *FixedLengthArcsBuffer) getBytes() []byte {
	return f.bytes
}
