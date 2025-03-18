package compressing

import (
	"bytes"
	"compress/flate"
	"context"
	"io"

	"github.com/pierrec/lz4/v4"

	"github.com/geange/lucene-go/core/store"
)

var (
	FAST             = NewFASTCompressionMode()
	HIGH_COMPRESSION = NewHighCompressionMode()
)

var _ CompressionMode = &FASTCompressionMode{}

type FASTCompressionMode struct {
}

func NewFASTCompressionMode() *FASTCompressionMode {
	return &FASTCompressionMode{}
}

func (m *FASTCompressionMode) NewCompressor() Compressor {
	return &LZ4FastCompressor{}
}

func (m *FASTCompressionMode) NewDecompressor() Decompressor {
	return &LZ4Decompressor{}
}

var _ CompressionMode = &HighCompressionMode{}

type HighCompressionMode struct {
}

func NewHighCompressionMode() *HighCompressionMode {
	return &HighCompressionMode{}
}

func (m *HighCompressionMode) NewCompressor() Compressor {
	return NewDeflateCompressor(6)
}

func (m *HighCompressionMode) NewDecompressor() Decompressor {
	return NewDeflateDecompressor()
}

var _ Compressor = &LZ4FastCompressor{}

type LZ4FastCompressor struct {
}

func (c *LZ4FastCompressor) Compress(ctx context.Context, bs []byte, out store.DataOutput) error {
	w := lz4.NewWriter(out)
	defer w.Close()

	if _, err := w.Write(bs); err != nil {
		return err
	}
	return nil
}

var _ Compressor = &LZ4HighCompressor{}

type LZ4HighCompressor struct {
}

func (l *LZ4HighCompressor) Compress(ctx context.Context, bs []byte, out store.DataOutput) error {
	w := lz4.NewWriter(out)
	defer w.Close()

	if _, err := w.Write(bs); err != nil {
		return err
	}
	return nil
}

var _ Decompressor = &LZ4Decompressor{}

type LZ4Decompressor struct {
}

func (d *LZ4Decompressor) Close() error {
	return nil
}

func (d *LZ4Decompressor) Decompress(ctx context.Context, in store.DataInput, offset int64, length int64, buf *bytes.Buffer) error {
	bs := make([]byte, length)
	if _, err := in.Read(bs); err != nil {
		return err
	}

	reader := lz4.NewReader(buf)
	if _, err := reader.Read(bs); err != nil {
		return err
	}
	return nil
}

func (d *LZ4Decompressor) Clone() Decompressor {
	return &LZ4Decompressor{}
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
