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

	IndexInputSPI

	//Clone() IndexInput
}

type IndexInputSPI interface {
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

type BaseIndexInput struct {
	*BaseDataInput

	spi IndexInputSPI
}

func (i *BaseIndexInput) RandomAccessSlice(offset int64, length int64) (RandomAccessInput, error) {
	slice, err := i.spi.Slice("randomaccess", offset, length)
	if err != nil {
		return nil, err
	}

	if random, ok := slice.(RandomAccessInput); ok {
		return random, nil
	}
	return &randomAccessIndexInput{in: slice}, nil
}

func NewBaseIndexInput(input IndexInput) *BaseIndexInput {
	return &BaseIndexInput{
		BaseDataInput: NewBaseDataInput(input),
		spi:           input,
	}
}

// IndexOutput A DataOutput for appending data to a file in a Directory.
// Instances of this class are not thread-safe.
// See Also: Directory, IndexInput
type IndexOutput interface {
	io.Closer

	DataOutput

	GetName() string

	// GetFilePointer Returns the current pos in this file,
	// where the next write will occur.
	GetFilePointer() int64

	GetChecksum() (uint32, error)
}

type BaseIndexOutput struct {
	*BaseDataOutput

	name string
}

func NewBaseIndexOutput(name string, writer io.Writer) *BaseIndexOutput {
	return &BaseIndexOutput{
		BaseDataOutput: NewBaseDataOutput(writer),
		name:           name,
	}
}

func (r *BaseIndexOutput) GetName() string {
	return r.name
}
