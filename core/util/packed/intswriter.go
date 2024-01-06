package packed

import (
	"context"
	"github.com/geange/lucene-go/codecs/utils"
	"github.com/geange/lucene-go/core/store"
)

// PackIntsWriter A write-once PackIntsWriter.
// lucene.internal
type PackIntsWriter interface {
	WriteHeader(ctx context.Context) error

	// GetFormat The format used to serialize values.
	GetFormat() Format

	// Add a value to the stream.
	Add(v int64) error

	// BitsPerValue The number of bits per value.
	BitsPerValue() int

	// Finish Perform end-of-stream operations.
	Finish() error

	// Ord Returns the current ord in the stream (number of values that have been written so far minus one).
	Ord() int
}

type WriterSpi interface {
	GetFormat() Format
}

type BasePackIntsWriter struct {
	FnGetFormat func() Format

	out          store.DataOutput
	valueCount   int
	bitsPerValue int
}

type PackIntsWriterDefaultConfig struct {
	GetFormat    func() Format
	out          store.DataOutput
	valueCount   int
	bitsPerValue int
}

func NewPackIntsWriterDefault(cfg *PackIntsWriterDefaultConfig) *BasePackIntsWriter {
	return &BasePackIntsWriter{
		FnGetFormat:  cfg.GetFormat,
		out:          cfg.out,
		valueCount:   cfg.valueCount,
		bitsPerValue: cfg.bitsPerValue,
	}
}

func newWriter(out store.DataOutput, valueCount int, bitsPerValue int) *BasePackIntsWriter {
	return &BasePackIntsWriter{
		out:          out,
		valueCount:   valueCount,
		bitsPerValue: bitsPerValue,
	}
}

func (w *BasePackIntsWriter) WriteHeader(ctx context.Context) error {
	err := utils.WriteHeader(w.out, CODEC_NAME, VERSION_CURRENT)
	if err != nil {
		return err
	}

	err = w.out.WriteUvarint(ctx, uint64(w.bitsPerValue))
	if err != nil {
		return err
	}

	err = w.out.WriteUvarint(ctx, uint64(w.valueCount))
	if err != nil {
		return err
	}

	return w.out.WriteUvarint(ctx, uint64(w.FnGetFormat().GetId()))
}

func (w *BasePackIntsWriter) BitsPerValue() int {
	return w.bitsPerValue
}
