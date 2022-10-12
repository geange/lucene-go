package store

import "io"

// IndexInput Abstract base class for input from a file in a Directory. A random-access input stream. Used for
// all Lucene index input operations.
//
// IndexInput may only be used from one thread, because it is not thread safe (it keeps internal state like
// file position). To allow multithreaded use, every IndexInput instance must be cloned before it is used in
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

	// GetFilePointer Returns the current position in this file, where the next read will occur.
	// See Also: seek(long)
	GetFilePointer() int64

	// Seek Sets current position in this file, where the next read will occur. If this is beyond the end
	// of the file then this will throw EOFException and then the stream is in an undetermined state.
	// See Also: getFilePointer()
	Seek(pos int64) error

	Clone() IndexInput

	// Slice Creates a slice of this index input, with the given description, offset, and length. The slice is sought to the beginning.
	Slice(sliceDescription string, offset, length int64) (IndexInput, error)

	// Length The number of bytes in the file.
	Length() int64

	// RandomAccessSlice Creates a random-access slice of this index input, with the given offset and length.
	// The default implementation calls slice, and it doesn't support random access, it implements absolute
	// reads as seek+read.
	RandomAccessSlice(offset int64, length int64) (RandomAccessInput, error)
}

type IndexInputImp struct {
	in IndexInput
	*DataInputImp
}

func NewIndexInputImp(in IndexInput) *IndexInputImp {
	return &IndexInputImp{
		in:           in,
		DataInputImp: NewDataInputImp(in),
	}
}

func (i *IndexInputImp) RandomAccessSlice(offset int64, length int64) (RandomAccessInput, error) {
	slice, err := i.in.Slice("randomaccess", offset, length)
	if err != nil {
		return nil, err
	}

	if random, ok := slice.(RandomAccessInput); ok {
		return random, nil
	}
	return &randomAccessIndexInput{in: slice}, nil
}

var _ RandomAccessInput = &randomAccessIndexInput{}

type randomAccessIndexInput struct {
	in IndexInput
}

func (r *randomAccessIndexInput) RUint8(pos int64) (byte, error) {
	err := r.in.Seek(pos)
	if err != nil {
		return 0, err
	}
	return r.in.ReadByte()
}

func (r *randomAccessIndexInput) RUint16(pos int64) (uint16, error) {
	err := r.in.Seek(pos)
	if err != nil {
		return 0, err
	}
	return r.in.ReadUint16()
}

func (r *randomAccessIndexInput) RUint32(pos int64) (uint32, error) {
	err := r.in.Seek(pos)
	if err != nil {
		return 0, err
	}
	return r.in.ReadUint32()
}

func (r *randomAccessIndexInput) RUint64(pos int64) (uint64, error) {
	err := r.in.Seek(pos)
	if err != nil {
		return 0, err
	}
	return r.in.ReadUint64()
}
