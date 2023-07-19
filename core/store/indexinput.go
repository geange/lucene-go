package store

import "io"

// IndexInput Abstract base class for input from a file in a Directory. A random-access input stream. Used for
// all Lucene index input operations.
//
// IndexInput may only be used from one thread, because it is not thread safe (it keeps internal state like
// file pos). To allow multithreaded use, every IndexInput instance must be cloned before it is used in
// another thread. Subclasses must therefore implement clone(), returning a new IndexInput which operates on
// the same underlying resource, but positioned independently.
//
// Warning: Lucene never closes cloned IndexInputs, it will only call close() on the original object.
// If you access the cloned IndexInput after closing the original object, any readXXX methods will throw
// AlreadyClosedException.
// See Also: Directory
type IndexInput interface {
	DataInput

	io.Closer

	// Seeker Sets current pos in this file, where the next read will occur. If this is beyond the end
	// of the file then this will throw EOFException and then the stream is in an undetermined state.
	// See Also: getFilePointer()
	//Seek(pos int64, whence int) (int64, error)
	io.Seeker

	IndexInputInner

	Clone() IndexInput
}

type IndexInputInner interface {
	// GetFilePointer Returns the current pos in this file, where the next read will occur.
	// See Also: seek(long)
	GetFilePointer() int64

	// Slice Creates a slice of this index input, with the given description, offset, and length.
	// The slice is sought to the beginning.
	Slice(sliceDescription string, offset, length int64) (IndexInput, error)

	// Length The number of bytes in the file.
	Length() int64

	// RandomAccessSlice Creates a random-access slice of this index input, with the given offset and length.
	// The default implementation calls slice, and it doesn't support random access, it implements absolute
	// reads as seek+read.
	RandomAccessSlice(offset int64, length int64) (RandomAccessInput, error)
}

type IndexInputBase struct {
	*Reader

	inner IndexInputInner
}

func (i *IndexInputBase) RandomAccessSlice(offset int64, length int64) (RandomAccessInput, error) {
	slice, err := i.inner.Slice("randomaccess", offset, length)
	if err != nil {
		return nil, err
	}

	if random, ok := slice.(RandomAccessInput); ok {
		return random, nil
	}
	return &randomAccessIndexInput{in: slice}, nil
}

type IndexInputDefaultConfig struct {
	Reader         io.Reader
	Close          func() error
	GetFilePointer func() int64
	Seek           func(pos int64, whence int) (int64, error)
	Slice          func(sliceDescription string, offset, length int64) (IndexInput, error)
	Length         func() int64
}

func NewIndexInputBase(input IndexInput) *IndexInputBase {
	return &IndexInputBase{
		Reader: NewReader(input),
		inner:  input,
	}
}

func (i *IndexInputBase) Clone(inner IndexInput) *IndexInputBase {
	reader, ok := inner.(io.Reader)
	if !ok {
		return nil
	}

	return &IndexInputBase{
		Reader: i.Reader.Clone(reader),
		inner:  inner,
	}
}

// RandomAccessInput Random Access Index API. Unlike IndexInput, this has no concept of file pos,
// all reads are absolute. However, like IndexInput, it is only intended for use by a single thread.
type RandomAccessInput interface {

	// RUint8 Reads a byte at the given pos in the file
	// See Also: DataInput.readByte
	RUint8(pos int64) (byte, error)

	// RUint16 Reads a short at the given pos in the file
	// See Also: DataInput.readShort
	RUint16(pos int64) (uint16, error)

	// RUint32 Reads an integer at the given pos in the file
	// See Also: DataInput.readInt
	RUint32(pos int64) (uint32, error)

	// RUint64 Reads a long at the given pos in the file
	// See Also: DataInput.readLong
	RUint64(pos int64) (uint64, error)
}

var _ RandomAccessInput = &randomAccessIndexInput{}

type randomAccessIndexInput struct {
	in IndexInput
}

func (r *randomAccessIndexInput) RUint8(pos int64) (byte, error) {
	_, err := r.in.Seek(pos, 0)
	if err != nil {
		return 0, err
	}
	return r.in.ReadByte()
}

func (r *randomAccessIndexInput) RUint16(pos int64) (uint16, error) {
	_, err := r.in.Seek(pos, 0)
	if err != nil {
		return 0, err
	}
	return r.in.ReadUint16()
}

func (r *randomAccessIndexInput) RUint32(pos int64) (uint32, error) {
	_, err := r.in.Seek(pos, 0)
	if err != nil {
		return 0, err
	}
	return r.in.ReadUint32()
}

func (r *randomAccessIndexInput) RUint64(pos int64) (uint64, error) {
	_, err := r.in.Seek(pos, 0)
	if err != nil {
		return 0, err
	}
	return r.in.ReadUint64()
}
