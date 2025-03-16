package compressing

import (
	"bytes"
	"compress/flate"
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
	level int
}

func NewDeflateCompressor(level int) *DeflateCompressor {
	return &DeflateCompressor{level: level}
}

func (d *DeflateCompressor) Compress(ctx context.Context, bs []byte, out store.DataOutput) error {
	buf := new(bytes.Buffer)
	writer, err := flate.NewWriter(buf, d.level)
	if err != nil {
		return err
	}

	if _, err := writer.Write(bs); err != nil {
		_ = writer.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

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

func (d *DeflateDecompressor) Decompress(ctx context.Context, in store.DataInput, offset int64, length int64, buf *bytes.Buffer) error {
	size, err := in.ReadUvarint(ctx)
	if err != nil {
		return err
	}
	bs := make([]byte, size)
	if _, err := in.Read(bs); err != nil {
		return err
	}

	reader := flate.NewReader(bytes.NewBuffer(bs))
	if _, err := io.Copy(buf, reader); err != nil {
		return err
	}
	return nil
}

func (d *DeflateDecompressor) Clone() Decompressor {
	return NewDeflateDecompressor()
}
