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

type WriterFormat interface {
	GetFormat() Format
}

type BasePackIntsWriter struct {
	format       WriterFormat
	out          store.DataOutput
	valueCount   int
	bitsPerValue int
}

func newBasePackIntsWriter(format WriterFormat,
	out store.DataOutput, valueCount int, bitsPerValue int) *BasePackIntsWriter {
	return &BasePackIntsWriter{
		format:       format,
		out:          out,
		valueCount:   valueCount,
		bitsPerValue: bitsPerValue,
	}
}

func (w *BasePackIntsWriter) WriteHeader(ctx context.Context) error {
	if err := utils.WriteHeader(w.out, CODEC_NAME, VERSION_CURRENT); err != nil {
		return err
	}

	if err := w.out.WriteUvarint(ctx, uint64(w.bitsPerValue)); err != nil {
		return err
	}

	if err := w.out.WriteUvarint(ctx, uint64(w.valueCount)); err != nil {
		return err
	}

	return w.out.WriteUvarint(ctx, uint64(w.format.GetFormat().GetId()))
}

func (w *BasePackIntsWriter) BitsPerValue() int {
	return w.bitsPerValue
}
