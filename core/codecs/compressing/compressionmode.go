package compressing

import (
	"archive/zip"

	"github.com/geange/lucene-go/core/store"
)

type FASTCompressionMode struct {
}

type LZ4FastCompressor struct {
}

var _ Compressor = &DeflateCompressor{}

type DeflateCompressor struct {
	compressor zip.Compressor
}

func (d *DeflateCompressor) Compress(bytes []byte, out store.DataOutput) error {
	panic("")
}
