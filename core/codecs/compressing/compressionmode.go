package compressing

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"io"

	"github.com/geange/lucene-go/core/store"
)

type FASTCompressionMode struct {
}

type LZ4FastCompressor struct {
}

var _ Compressor = &DeflateCompressor{}

type DeflateCompressor struct {
}

func NewDeflateCompressor() *DeflateCompressor {
	return &DeflateCompressor{}
}

func (d *DeflateCompressor) Compress(ctx context.Context, bs []byte, out store.DataOutput) error {
	buf := bytes.NewBuffer(bs)
	zw := zip.NewWriter(buf)
	defer zw.Close()

	size := buf.Len()
	if err := out.WriteUvarint(ctx, uint64(size)); err != nil {
		return err
	}

	if _, err := io.Copy(out, buf); err != nil {
		return nil
	}
	return nil
}

var _ Decompressor = &DeflateDecompressor{}

type DeflateDecompressor struct {
}

func NewDeflateDecompressor() *DeflateDecompressor {
	return &DeflateDecompressor{}
}

func (d *DeflateDecompressor) Close() error {
	return nil
}

func (d *DeflateDecompressor) Decompress(ctx context.Context, in store.DataInput, buf *bytes.Buffer) error {
	size, err := in.ReadUvarint(ctx)
	if err != nil {
		return err
	}
	bs := make([]byte, size)
	if _, err := in.Read(bs); err != nil {
		return err
	}

	zr, err := gzip.NewReader(bytes.NewBuffer(bs))
	if err != nil {
		return err
	}
	defer zr.Close()

	if _, err := io.Copy(buf, zr); err != nil {
		return err
	}
	return nil
}

func (d *DeflateDecompressor) Clone() Decompressor {
	return NewDeflateDecompressor()
}
