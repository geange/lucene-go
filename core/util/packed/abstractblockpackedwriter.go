package packed

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/geange/lucene-go/core/store"
)

const (
	ABP_MIN_BLOCK_SIZE     = 64
	ABP_MAX_BLOCK_SIZE     = 1 << (30 - 3)
	ABP_MIN_VALUE_EQUALS_0 = 1 << 0
	ABP_BPV_SHIFT          = 1
)

type AbstractBlockPackedWriter struct {
	out      store.DataOutput
	values   []uint64
	blocks   []byte
	off      int
	ord      int
	finished *atomic.Bool
	flusher  BlockPackedFlusher
}

func newAbstractBlockPackedWriter(out store.DataOutput, blockSize int) *AbstractBlockPackedWriter {
	writer := &AbstractBlockPackedWriter{
		finished: new(atomic.Bool),
	}
	writer.reset(out)
	writer.values = make([]uint64, blockSize)

	return writer
}

// Reset this writer to wrap out. The block size remains unchanged.
func (a *AbstractBlockPackedWriter) reset(out store.DataOutput) {
	a.out = out
	a.off = 0
	a.ord = 0
	a.finished.Store(false)
}

func (a *AbstractBlockPackedWriter) checkNotFinished() error {
	if a.finished.Load() {
		return errors.New("already finished")
	}
	return nil
}

// Add
// Append a new long.
func (a *AbstractBlockPackedWriter) Add(ctx context.Context, v uint64) error {
	if err := a.checkNotFinished(); err != nil {
		return err
	}

	if a.off == len(a.values) {
		if err := a.flush(ctx); err != nil {
			return err
		}
	}
	a.values[a.off] = v
	a.off++
	a.ord++

	return nil
}

func (a *AbstractBlockPackedWriter) Finish(ctx context.Context) error {
	if a.off > 0 {
		if err := a.flush(ctx); err != nil {
			return err
		}
	}
	a.finished.Store(true)
	return nil
}

type BlockPackedFlusher interface {
	Flush(ctx context.Context) error
}

func (a *AbstractBlockPackedWriter) flush(ctx context.Context) error {
	return a.flusher.Flush(ctx)
}

func (a *AbstractBlockPackedWriter) writeValues(bitsRequired int) error {
	encoder, err := GetEncoder(FormatPacked, VERSION_CURRENT, bitsRequired)
	if err != nil {
		return err
	}
	iterations := len(a.values) / encoder.ByteValueCount()
	blockSize := encoder.ByteBlockCount() * iterations

	if len(a.blocks) < blockSize {
		a.blocks = make([]byte, blockSize)
	}
	if a.off < len(a.values) {
		clear(a.values[a.off:])
	}

	encoder.EncodeBytes(a.values, a.blocks, iterations)
	blockCount := FormatPacked.ByteCount(VERSION_CURRENT, a.off, bitsRequired)

	if _, err := a.out.Write(a.blocks[:blockCount]); err != nil {
		return err
	}
	return nil
}
