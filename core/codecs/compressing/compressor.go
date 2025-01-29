package compressing

import (
	"io"

	"github.com/geange/lucene-go/core/store"
)

type Compressor interface {
	Compress(bytes []byte, out store.DataOutput) error
}

type Decompressor interface {
	io.Closer

	Decompress(in store.DataInput, originalLength int, bs []byte) error
}
