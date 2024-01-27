package packed

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/geange/lucene-go/core/store"
)

type AbstractBlockPackedWriter struct {
	out          store.DataOutput
	values       []uint64
	valuesOffset int
	blocks       []byte
	blockSize    int
	count        int
	finished     *atomic.Bool
	flusher      BlockPackedFlusher

	MIN_BLOCK_SIZE     int
	MAX_BLOCK_SIZE     int
	MIN_VALUE_EQUALS_0 int
	BPV_SHIFT          int
}

func newAbstractBlockPackedWriter(out store.DataOutput, blockSize int, flusher BlockPackedFlusher) *AbstractBlockPackedWriter {
	writer := &AbstractBlockPackedWriter{
		//out:       out,
		values:    make([]uint64, 0, blockSize),
		blocks:    make([]byte, 0),
		blockSize: blockSize,
		count:     0,
		finished:  &atomic.Bool{},
		flusher:   flusher,
	}
	writer.reset(out)
	return writer
}

func (a *AbstractBlockPackedWriter) initConst() {
	a.MIN_BLOCK_SIZE = 64
	a.MAX_BLOCK_SIZE = 1 << (30 - 3)
	a.MIN_VALUE_EQUALS_0 = 1 << 0
	a.BPV_SHIFT = 1
}

// Reset this writer to wrap out. The block size remains unchanged.
func (a *AbstractBlockPackedWriter) reset(out store.DataOutput) {
	a.out = out
	a.count = 0
	a.finished.Store(false)
	a.valuesOffset = 0
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

	if a.valuesOffset == a.blockSize {
		if err := a.flush(ctx); err != nil {
			return err
		}
	}
	a.values[a.valuesOffset] = v
	a.valuesOffset++
	a.count++
	return nil
}

func (a *AbstractBlockPackedWriter) Finish(ctx context.Context) error {
	if len(a.values) > 0 {
		if err := a.flush(ctx); err != nil {
			return err
		}
	}
	a.finished.Store(true)
	return nil
}

type BlockPackedFlusher interface {
	Flush(ctx context.Context, w *AbstractBlockPackedWriter) error
}

func (a *AbstractBlockPackedWriter) flush(ctx context.Context) error {
	return a.flusher.Flush(ctx, a)
}

func (a *AbstractBlockPackedWriter) writeValues(bitsRequired int) error {
	encoder, err := GetEncoder(FormatPacked, VERSION_CURRENT, bitsRequired)
	if err != nil {
		return err
	}
	iterations := len(a.values) / encoder.ByteValueCount()
	blockSize := encoder.ByteBlockCount() * iterations
	if len(a.blocks) == 0 || len(a.blocks) < blockSize {
		a.blocks = make([]byte, blockSize)
	}

	if a.valuesOffset < a.blockSize {
		// make it to zero
		clear(a.values[a.valuesOffset:])
	}
	encoder.EncodeBytes(a.values, a.blocks, iterations)
	blockCount := FormatPacked.ByteCount(VERSION_CURRENT, a.valuesOffset, bitsRequired)
	if _, err := a.out.Write(a.blocks[:blockCount]); err != nil {
		return err
	}
	return nil
}
