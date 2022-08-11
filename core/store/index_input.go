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
}
