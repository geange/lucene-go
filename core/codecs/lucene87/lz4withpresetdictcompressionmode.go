package lucene87

import (
	"bytes"
	"context"

	"github.com/pierrec/lz4/v4"

	"github.com/geange/lucene-go/core/codecs/compressing"
	"github.com/geange/lucene-go/core/store"
)

var _ compressing.CompressionMode = &LZ4WithPresetDictCompressionMode{}

type LZ4WithPresetDictCompressionMode struct {
}

func NewLZ4WithPresetDictCompressionMode() *LZ4WithPresetDictCompressionMode {
	return &LZ4WithPresetDictCompressionMode{}
}

func (m *LZ4WithPresetDictCompressionMode) NewCompressor() compressing.Compressor {
	return &LZ4WithPresetDictCompressor{buf: new(bytes.Buffer)}
}

func (m *LZ4WithPresetDictCompressionMode) NewDecompressor() compressing.Decompressor {
	return &LZ4WithPresetDictDecompressor{}
}

var _ compressing.Compressor = &LZ4WithPresetDictCompressor{}

type LZ4WithPresetDictCompressor struct {
	buf *bytes.Buffer
}

func (c *LZ4WithPresetDictCompressor) doCompress(ctx context.Context, bs []byte, out store.DataOutput) error {
	if len(bs) == 0 {
		return out.WriteUvarint(ctx, 0)
	}

	c.buf.Reset()
	writer := lz4.NewWriter(c.buf)
	if _, err := writer.Write(bs); err != nil {
		_ = writer.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	if err := out.WriteUvarint(ctx, uint64(c.buf.Len())); err != nil {
		return err
	}
	if _, err := out.Write(c.buf.Bytes()); err != nil {
		return err
	}
	return nil
}

func (c *LZ4WithPresetDictCompressor) Compress(ctx context.Context, bs []byte, out store.DataOutput) error {
	size := len(bs)
	dictLength := size / (NUM_SUB_BLOCKS * DICT_SIZE_FACTOR)
	blockLength := (size - dictLength + NUM_SUB_BLOCKS - 1) / NUM_SUB_BLOCKS
	if err := out.WriteUvarint(ctx, uint64(dictLength)); err != nil {
		return err
	}
	if err := out.WriteUvarint(ctx, uint64(blockLength)); err != nil {
		return err
	}
	end := size

	// Compress the dictionary first
	if err := c.doCompress(ctx, bs[:dictLength], out); err != nil {
		return err
	}

	// And then sub blocks
	for start := dictLength; start < end; start += blockLength {
		endOff := start + min(blockLength, size-start)

		if err := c.doCompress(ctx, bs[start:endOff], out); err != nil {
			return err
		}
	}

	return nil
}

var _ compressing.Decompressor = &DeflateWithPresetDictDecompressor{}

type LZ4WithPresetDictDecompressor struct {
}

func (d *LZ4WithPresetDictDecompressor) Close() error {
	return nil
}

func (d *LZ4WithPresetDictDecompressor) doDecompress(ctx context.Context, in store.DataInput, buf *bytes.Buffer) error {
	compressedLength, err := in.ReadUvarint(ctx)
	if err != nil {
		return err
	}
	bs := make([]byte, compressedLength)
	if _, err := in.Read(bs); err != nil {
		return err
	}

	reader := lz4.NewReader(buf)
	if _, err := reader.Read(bs); err != nil {
		return err
	}
	return nil
}

func (d *LZ4WithPresetDictDecompressor) Decompress(ctx context.Context, in store.DataInput, offset int64, length int64, buf *bytes.Buffer) error {
	dictLength, err := in.ReadUvarint(ctx)
	if err != nil {
		return err
	}
	blockLength, err := in.ReadUvarint(ctx)
	if err != nil {
		return err
	}
	// Read the dictionary
	if err := d.doDecompress(ctx, in, buf); err != nil {
		return err
	}

	offsetInBlock := int64(dictLength)
	offsetInBytesRef := offset

	// Skip unneeded blocks
	for offsetInBlock+int64(blockLength) < offset {
		compressedLength, err := in.ReadUvarint(ctx)
		if err != nil {
			return err
		}

		if err := in.SkipBytes(ctx, int(compressedLength)); err != nil {
			return err
		}
		offsetInBlock += int64(blockLength)
		offsetInBytesRef -= int64(blockLength)
	}

	// Read blocks that intersect with the interval we need
	for offsetInBlock < offset+length {
		if err := d.doDecompress(ctx, in, buf); err != nil {
			return err
		}
		offsetInBlock += int64(blockLength)
	}

	return nil
}

func (d *LZ4WithPresetDictDecompressor) Clone() compressing.Decompressor {
	return &LZ4WithPresetDictDecompressor{}
}
