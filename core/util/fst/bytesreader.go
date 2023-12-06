package fst

import (
	"context"
	"slices"

	"github.com/geange/lucene-go/core/store"
)

// BytesReader
// Reads bytes stored in an FST.
type BytesReader interface {
	store.DataInput

	// GetPosition Get current read position.
	GetPosition() int64

	// SetPosition Set current read position.
	SetPosition(pos int64) error

	// Reversed Returns true if this reader uses reversed bytes under-the-hood.
	Reversed() bool
}

var _ BytesReader = &ReverseBytesReader{}

type ReverseBytesReader struct {
	*store.Reader

	bytes []byte
	pos   int64
}

func newReverseBytesReader(bytes []byte) *ReverseBytesReader {
	reader := &ReverseBytesReader{
		bytes: bytes,
		pos:   int64(len(bytes) - 1),
	}
	reader.Reader = store.NewReader(reader)
	return reader
}

func (r *ReverseBytesReader) ReadByte() (byte, error) {
	b := r.bytes[r.pos]
	r.pos--
	return b, nil
}

func (r *ReverseBytesReader) Read(b []byte) (int, error) {
	for i := 0; i < len(b); i++ {
		b[i] = r.bytes[r.pos]
		r.pos--
	}
	return len(b), nil
}

func (r *ReverseBytesReader) SkipBytes(ctx context.Context, numBytes int) error {
	r.pos -= int64(numBytes)
	return nil
}

func (r *ReverseBytesReader) GetPosition() int64 {
	return r.pos
}

func (r *ReverseBytesReader) SetPosition(pos int64) error {
	r.pos = pos
	return nil
}

func (r *ReverseBytesReader) Reversed() bool {
	return true
}

type ReverseRandomAccessReader struct {
	*store.Reader

	in  store.RandomAccessInput
	pos int64
}

func newReverseRandomAccessReader(in store.RandomAccessInput) *ReverseRandomAccessReader {
	reader := &ReverseRandomAccessReader{
		in: in,
	}
	reader.Reader = store.NewReader(reader)
	return reader
}

func (r *ReverseRandomAccessReader) ReadByte() (byte, error) {
	pos := r.pos
	r.pos--
	return r.in.RUint8(pos)
}

func (r *ReverseRandomAccessReader) Read(b []byte) (int, error) {
	pos := r.pos - int64(len(b)) + 1
	n, err := r.in.ReadAt(b, pos)
	if err != nil {
		return 0, err
	}
	slices.Reverse(b)
	return n, nil
}

func (r *ReverseRandomAccessReader) SkipBytes(ctx context.Context, numBytes int) error {
	r.pos -= int64(numBytes)
	return nil
}

func (r *ReverseRandomAccessReader) GetPosition() int64 {
	return r.pos
}

func (r *ReverseRandomAccessReader) SetPosition(pos int64) error {
	r.pos = pos
	return nil
}

func (r *ReverseRandomAccessReader) Reversed() bool {
	return true
}

var _ BytesReader = &builderBytesReader{}

type builderBytesReader struct {
	*store.Reader

	bs         *ByteStore
	current    []byte
	nextBuffer int
	nextRead   int
}

func newBuilderBytesReader(bs *ByteStore) (*builderBytesReader, error) {
	var current []byte
	if bs.blocks.Size() != 0 {
		v, ok := bs.blocks.Get(0)
		if !ok {
			return nil, ErrItemNotFound
		}
		current = v
	}

	reader := &builderBytesReader{
		current:    current,
		bs:         bs,
		nextBuffer: -1,
		nextRead:   0,
	}

	reader.Reader = store.NewReader(reader)
	return reader, nil
}

func (b *builderBytesReader) ReadByte() (byte, error) {
	if b.nextRead == -1 {
		current, ok := b.bs.blocks.Get(b.nextBuffer)
		if !ok {
			return 0, ErrItemNotFound
		}
		b.current = current
		b.nextBuffer++
		b.nextRead = int(b.bs.blockSize - 1)
	}
	v := b.current[b.nextRead]
	b.nextRead--
	return v, nil
}

func (b *builderBytesReader) Read(bs []byte) (int, error) {
	for i := range bs {
		v, err := b.ReadByte()
		if err != nil {
			return 0, err
		}
		bs[i] = v
	}
	return len(bs), nil
}

func (b *builderBytesReader) SkipBytes(ctx context.Context, numBytes int) error {
	return b.SetPosition(b.GetPosition() - int64(numBytes))
}

func (b *builderBytesReader) GetPosition() int64 {
	return int64(b.nextBuffer+1)*b.bs.blockSize + int64(b.nextRead)
}

func (b *builderBytesReader) SetPosition(pos int64) error {
	// NOTE: a little weird because if you
	// setPosition(0), the next byte you read is
	// bytes[0] ... but I would expect bytes[-1] (ie,
	// isEOF)...?
	bufferIndex := (int)(pos >> b.bs.blockBits)
	if b.nextBuffer != bufferIndex-1 {
		b.nextBuffer = bufferIndex - 1
		v, ok := b.bs.blocks.Get(bufferIndex)
		if !ok {
			return ErrItemNotFound
		}
		b.current = v
	}
	b.nextRead = int(pos & b.bs.blockMask)
	return nil
}

func (b *builderBytesReader) Reversed() bool {
	return true
}
