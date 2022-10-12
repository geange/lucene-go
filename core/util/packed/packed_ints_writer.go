package packed

import (
	"github.com/geange/lucene-go/codecs"
	"github.com/geange/lucene-go/core/store"
)

// Writer A write-once Writer.
// lucene.internal
type Writer interface {
	WriteHeader() error

	// GetFormat The format used to serialize values.
	GetFormat() Format

	// Add a value to the stream.
	Add(v uint64) error

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

type writer struct {
	spi          WriterSpi
	out          store.DataOutput
	valueCount   int
	bitsPerValue int
}

func newWriter(out store.DataOutput, valueCount int, bitsPerValue int) *writer {
	return &writer{
		out:          out,
		valueCount:   valueCount,
		bitsPerValue: bitsPerValue,
	}
}

func (w *writer) WriteHeader() error {
	err := codecs.WriteHeader(w.out, CODEC_NAME, VERSION_CURRENT)
	if err != nil {
		return err
	}

	err = w.out.WriteUvarint(uint64(w.bitsPerValue))
	if err != nil {
		return err
	}

	err = w.out.WriteUvarint(uint64(w.valueCount))
	if err != nil {
		return err
	}

	return w.out.WriteUvarint(uint64(w.spi.GetFormat().GetId()))
}

func (w *writer) BitsPerValue() int {
	return w.bitsPerValue
}
