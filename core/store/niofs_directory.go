package store

import (
	"bytes"
	"errors"
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
	*BufferedIndexInputDefault

	file    *os.File
	pointer int64
}

func NewNIOFSIndexInput(file *os.File, ctx *IOContext) *NIOFSIndexInput {
	_, err := file.Stat()
	if err != nil {
		return nil
	}

	input := &NIOFSIndexInput{
		file: file,
	}

	cfg := &BufferedIndexInputDefaultConfig{
		IndexInputDefaultConfig: IndexInputDefaultConfig{
			DataInputDefaultConfig: DataInputDefaultConfig{
				ReadByte: input.ReadByte,
				Read:     input.Read,
			},
			Close:          input.Close,
			GetFilePointer: input.GetFilePointer,
			Seek:           input.Seek,
			Slice:          input.Slice,
			Length:         input.Length,
		},
		ReadInternal: input.ReadInternal,
		SeekInternal: input.SeekInternal,
	}

	input.BufferedIndexInputDefault = NewBufferedIndexInputDefault(cfg)

	return input
}

func (n *NIOFSIndexInput) ReadInternal(buf *bytes.Buffer, size int) error {
	bs := make([]byte, size)
	num, err := n.Read(bs)
	if err != nil {
		if num > 0 && errors.Is(err, io.EOF) {
			buf.Write(bs)
			return nil
		}
	}
	return err
}

func (n *NIOFSIndexInput) SeekInternal(pos int) error {
	stat, err := n.file.Stat()
	if err != nil {
		return err
	}
	if pos > int(stat.Size()) {
		return errors.New("pos too large")
	}
	return nil
}

func (n *NIOFSIndexInput) Clone() IndexInput {
	input := &NIOFSIndexInput{
		file:    n.file,
		pointer: n.pointer,
	}

	cfg := &BufferedIndexInputDefaultConfig{
		IndexInputDefaultConfig: IndexInputDefaultConfig{
			DataInputDefaultConfig: DataInputDefaultConfig{
				ReadByte: input.ReadByte,
				Read:     input.Read,
			},
			Close:          input.Close,
			GetFilePointer: input.GetFilePointer,
			Seek:           input.Seek,
			Slice:          input.Slice,
			Length:         input.Length,
		},
		ReadInternal: input.readInternal,
		SeekInternal: input.seekInternal,
	}
	input.BufferedIndexInputDefault = n.BufferedIndexInputDefault.Clone(cfg)
	return input
}

func (n *NIOFSIndexInput) ReadByte() (byte, error) {
	bs := [1]byte{}
	_, err := n.file.ReadAt(bs[:], n.pointer)
	if err != nil {
		return 0, err
	}
	n.pointer++
	return bs[0], nil
}

func (n *NIOFSIndexInput) Read(b []byte) (int, error) {
	num, err := n.file.ReadAt(b, n.pointer)
	if err != nil {
		if err != io.EOF {
			return 0, err
		}
	}
	n.pointer += int64(num)
	return len(b), err
}

func (n *NIOFSIndexInput) Close() error {
	return n.file.Close()
}

func (n *NIOFSIndexInput) Seek(pos int64, whence int) (int64, error) {
	return n.file.Seek(pos, io.SeekStart)
}

func (n *NIOFSIndexInput) Length() int64 {
	info, _ := n.file.Stat()
	return info.Size()
}

func (n *NIOFSIndexInput) Slice(sliceDescription string, offset, length int64) (IndexInput, error) {
	//TODO implement me
	panic("implement me")
}
