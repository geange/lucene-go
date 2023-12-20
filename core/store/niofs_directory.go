package store

import (
	"io"
	"os"
)

var _ FSDirectory = &NIOFSDirectory{}

// NIOFSDirectory
// An FSDirectory implementation that uses java.nio's FileChannel's positional read,
// which allows multiple threads to read from the same file without synchronizing.
// This class only uses FileChannel when reading; writing is achieved with FSDirectory.FSIndexOutput.
// NOTE: NIOFSDirectory is not recommended on Windows because of a bug in how FileChannel.read is
// implemented in Sun's JRE. Inside of the implementation the pos is apparently synchronized.
// See here  for details.
// NOTE: Accessing this class either directly or indirectly from a thread while it's interrupted can
// close the underlying file descriptor immediately if at the same time the thread is blocked on IO.
// The file descriptor will remain closed and subsequent access to NIOFSDirectory will throw a ClosedChannelException.
// If your application uses either Thread.interrupt() or Future.cancel(boolean) you should use the legacy
// RAFDirectory from the Lucene misc module in favor of NIOFSDirectory.
type NIOFSDirectory struct {
	*FSDirectoryBase
}

func NewNIOFSDirectory(path string) (*NIOFSDirectory, error) {
	base, err := NewFSDirectoryBase(path, NewSimpleFSLockFactory())
	if err != nil {
		return nil, err
	}
	dir := &NIOFSDirectory{base}
	base.BaseDirectoryBase = &BaseDirectoryBase{dir: dir}
	return dir, nil
}

func (n *NIOFSDirectory) OpenInput(name string, context *IOContext) (IndexInput, error) {
	if err := n.EnsureOpen(); err != nil {
		return nil, err
	}
	if err := n.ensureCanRead(name); err != nil {
		return nil, err
	}
	path := n.resolveFilePath(name)

	file, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}

	return NewNIOFSIndexInput(file, context), nil
}

var _ IndexInput = &NIOFSIndexInput{}

type NIOFSIndexInput struct {
	*IndexInputBase

	file    *os.File
	off     int64
	end     int64
	pos     int64
	isClone bool
}

func (n *NIOFSIndexInput) Read(p []byte) (size int, err error) {
	if n.pos >= n.end {
		return 0, io.EOF
	}

	size = len(p)
	left := int(n.end - n.pos)
	if left < size {
		size = left
	}

	num, err := n.file.ReadAt(p[:size], n.pos)
	if err != nil {
		return 0, err
	}
	n.pos += int64(num)
	return num, nil
}

func NewNIOFSIndexInput(file *os.File, ctx *IOContext) *NIOFSIndexInput {
	info, err := file.Stat()
	if err != nil {
		return nil
	}

	input := &NIOFSIndexInput{
		file: file,
		off:  0,
		pos:  0,
		end:  info.Size(),
	}

	input.IndexInputBase = NewIndexInputBase(input)
	return input
}

func NewNIOFSIndexInputV1(file *os.File, off, length int64) *NIOFSIndexInput {
	input := &NIOFSIndexInput{
		file:    file,
		off:     off,
		pos:     off,
		end:     off + length,
		isClone: true,
	}

	input.IndexInputBase = NewIndexInputBase(input)

	return input
}

func (n *NIOFSIndexInput) Clone() IndexInput {
	input := &NIOFSIndexInput{
		file:    n.file,
		isClone: true,
		off:     n.off,
		pos:     n.pos,
		end:     n.end,
	}

	input.IndexInputBase = NewIndexInputBase(input)
	return input
}

func (n *NIOFSIndexInput) Close() error {
	if n.isClone {
		return nil
	}
	return n.file.Close()
}

func (n *NIOFSIndexInput) Seek(pos int64, whence int) (int64, error) {
	n.pos = n.off + pos
	return n.file.Seek(n.pos, whence)
}

func (n *NIOFSIndexInput) Length() int64 {
	return n.end - n.off
}

func (n *NIOFSIndexInput) Slice(sliceDescription string, offset, length int64) (IndexInput, error) {
	return NewNIOFSIndexInputV1(n.file, n.off+offset, length), nil
}

func (n *NIOFSIndexInput) GetFilePointer() int64 {
	return n.pos - n.off
}
