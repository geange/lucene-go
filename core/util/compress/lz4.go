package compress

import (
	"github.com/pierrec/lz4/v4"

	"github.com/geange/lucene-go/core/store"
)

var LZ4Compression = &LZ4{}

type LZ4 struct {
}

func (*LZ4) Compress(in []byte, out store.DataOutput) error {
	w := lz4.NewWriter(out)
	_, err := w.Write(in)
	return err
}

func (*LZ4) Decompress(in store.DataInput, out []byte) error {
	r := lz4.NewReader(in)
	_, err := r.Read(out)
	return err
}
