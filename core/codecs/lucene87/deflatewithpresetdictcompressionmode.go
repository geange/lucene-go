package lucene87

import (
	"bytes"
	"compress/flate"
	"context"

	"github.com/geange/lucene-go/core/codecs/compressing"
	"github.com/geange/lucene-go/core/store"
)

var _ compressing.CompressionMode = &DeflateWithPresetDictCompressionMode{}

const (
	NUM_SUB_BLOCKS   = 10 // Shoot for 10 sub blocks
	DICT_SIZE_FACTOR = 6  // And a dictionary whose size is about 6x smaller than sub blocks
)

// DeflateWithPresetDictCompressionMode A compression mode that trades speed for compression ratio.
// Although compression and decompression might be slow, this compression mode should provide a good compression ratio.
// This mode might be interesting if/ when your index size is much bigger than your OS cache.
type DeflateWithPresetDictCompressionMode struct {
}

func NewDeflateWithPresetDictCompressionMode() *DeflateWithPresetDictCompressionMode {
	return &DeflateWithPresetDictCompressionMode{}
}

func (d *DeflateWithPresetDictCompressionMode) NewCompressor() compressing.Compressor {
	return &DeflateWithPresetDictCompressor{buf: new(bytes.Buffer)}
}

func (d *DeflateWithPresetDictCompressionMode) NewDecompressor() compressing.Decompressor {
	return &DeflateWithPresetDictDecompressor{}
}

var _ compressing.Compressor = &DeflateWithPresetDictCompressor{}

type DeflateWithPresetDictCompressor struct {
	buf *bytes.Buffer
}

func (d *DeflateWithPresetDictCompressor) Compress(ctx context.Context, bs []byte, out store.DataOutput) error {

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
	if err := d.doCompress(ctx, bs[:dictLength], out); err != nil {
		return err
	}

	// And then sub blocks
	for start := dictLength; start < end; start += blockLength {
		endOff := start + min(blockLength, size-start)

		if err := d.doCompress(ctx, bs[start:endOff], out); err != nil {
			return err
		}
	}

	return nil
}

func (d *DeflateWithPresetDictCompressor) doCompress(ctx context.Context, bs []byte, out store.DataOutput) error {
	if len(bs) == 0 {
		return out.WriteUvarint(ctx, 0)
	}

	d.buf.Reset()
	writer, err := flate.NewWriter(d.buf, 6)
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

	if err := out.WriteUvarint(ctx, uint64(d.buf.Len())); err != nil {
		return err
	}
	if _, err := out.Write(d.buf.Bytes()); err != nil {
		return err
	}
	return nil
}

var _ compressing.Decompressor = &DeflateWithPresetDictDecompressor{}

type DeflateWithPresetDictDecompressor struct {
}

func (d *DeflateWithPresetDictDecompressor) Close() error {
	return nil
}

func (d *DeflateWithPresetDictDecompressor) doDecompress(ctx context.Context, in store.DataInput, buf *bytes.Buffer) error {
	compressedLength, err := in.ReadUvarint(ctx)
	if err != nil {
		return err
	}
	bs := make([]byte, compressedLength)
	if _, err := in.Read(bs); err != nil {
		return err
	}

	reader := flate.NewReader(buf)
	if _, err := reader.Read(bs); err != nil {
		return err
	}
	return nil
}

func (d *DeflateWithPresetDictDecompressor) Decompress(ctx context.Context, in store.DataInput, offset int64, length int64, buf *bytes.Buffer) error {
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

func (d *DeflateWithPresetDictDecompressor) Clone() compressing.Decompressor {
	return &DeflateWithPresetDictDecompressor{}
}
