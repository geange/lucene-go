package store

import (
	"io"
	"os"
)

var _ FSDirectory = &NIOFSDirectory{}

// NIOFSDirectory An FSDirectory implementation that uses java.nio's FileChannel's positional read, which allows multiple threads to read from the same file without synchronizing.
// This class only uses FileChannel when reading; writing is achieved with FSDirectory.FSIndexOutput.
// NOTE: NIOFSDirectory is not recommended on Windows because of a bug in how FileChannel.read is implemented in Sun's JRE. Inside of the implementation the position is apparently synchronized. See here  for details.
// NOTE: Accessing this class either directly or indirectly from a thread while it's interrupted can close the underlying file descriptor immediately if at the same time the thread is blocked on IO. The file descriptor will remain closed and subsequent access to NIOFSDirectory will throw a ClosedChannelException. If your application uses either Thread.interrupt() or Future.cancel(boolean) you should use the legacy RAFDirectory from the Lucene misc module in favor of NIOFSDirectory.
type NIOFSDirectory struct {
	*FSDirectoryImp
}

func NewNIOFSDirectory(path string) (*NIOFSDirectory, error) {
	directory, err := NewFSDirectoryImp(path, NewSimpleFSLockFactory())
	if err != nil {
		return nil, err
	}
	return &NIOFSDirectory{directory}, nil
}

func (n *NIOFSDirectory) OpenInput(name string, context *IOContext) (IndexInput, error) {
	if err := n.EnsureOpen(); err != nil {
		return nil, err
	}
	if err := n.ensureCanRead(name); err != nil {
		return nil, err
	}
	path := n.resolveFilePath(name)

	file, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}

	return NewNIOFSIndexInput(file, context), nil
}

var _ BufferedIndexInput = &NIOFSIndexInput{}

type NIOFSIndexInput struct {
	*IndexInputImp

	file    *os.File
	pointer int64

	buffer []byte
}

func NewNIOFSIndexInput(file *os.File, ctx *IOContext) *NIOFSIndexInput {
	_, err := file.Stat()
	if err != nil {
		return nil
	}

	input := &NIOFSIndexInput{
		file:   file,
		buffer: make([]byte, 48),
	}
	input.IndexInputImp = NewIndexInputImp(input)
	return input
}

func (n *NIOFSIndexInput) Clone() IndexInput {
	//TODO implement me
	panic("implement me")
}

func (n *NIOFSIndexInput) SetBufferSize(newSize int) {
	//TODO implement me
	panic("implement me")
}

func (n *NIOFSIndexInput) ReadByte() (byte, error) {
	_, err := n.file.Read(n.buffer[:1])
	if err != nil {
		return 0, err
	}
	n.pointer++
	return n.buffer[0], nil
}

func (n *NIOFSIndexInput) ReadBytes(b []byte) error {
	num, err := n.file.Read(b)
	n.pointer += int64(num)
	return err
}

func (n *NIOFSIndexInput) Close() error {
	return n.file.Close()
}

func (n *NIOFSIndexInput) GetFilePointer() int64 {
	return 0
}

func (n *NIOFSIndexInput) Seek(pos int64) error {
	_, err := n.file.Seek(pos, io.SeekStart)
	return err
}

func (n *NIOFSIndexInput) Length() int64 {
	info, _ := n.file.Stat()
	return info.Size()
}

func (n *NIOFSIndexInput) Slice(sliceDescription string, offset, length int64) (IndexInput, error) {
	//TODO implement me
	panic("implement me")
}
