package store

import (
	"io"
	"os"
)

var _ FSDirectory = &NIOFSDirectory{}

// NIOFSDirectory An FSDirectory implementation that uses java.nio's FileChannel's positional read, which allows multiple threads to read from the same file without synchronizing.
// This class only uses FileChannel when reading; writing is achieved with FSDirectory.FSIndexOutput.
// NOTE: NIOFSDirectory is not recommended on Windows because of a bug in how FileChannel.read is implemented in Sun's JRE. Inside of the implementation the pos is apparently synchronized. See here  for details.
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

var _ IndexInput = &NIOFSIndexInput{}

type NIOFSIndexInput struct {
	*IndexInputDefault

	file *os.File

	off     int64
	end     int64
	pos     int64
	isClone bool
}

func (n *NIOFSIndexInput) Read(p []byte) (size int, err error) {
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

	input.IndexInputDefault = NewIndexInputDefault(&IndexInputDefaultConfig{
		Reader:         input,
		Close:          input.Close,
		GetFilePointer: input.GetFilePointer,
		Seek:           input.Seek,
		Slice:          input.Slice,
		Length:         input.Length,
	})

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

	input.IndexInputDefault = NewIndexInputDefault(&IndexInputDefaultConfig{
		Reader:         input,
		Close:          input.Close,
		GetFilePointer: input.GetFilePointer,
		Seek:           input.Seek,
		Slice:          input.Slice,
		Length:         input.Length,
	})

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

	//cfg := &BufferedIndexInputDefaultConfig{
	//	IndexInputDefaultConfig: IndexInputDefaultConfig{
	//		DataInputDefaultConfig: DataInputDefaultConfig{
	//			ReadByte: input.ReadByte,
	//			Read:     input.Read,
	//		},
	//		Close:          input.Close,
	//		GetFilePointer: input.GetFilePointer,
	//		Seek:           input.Seek,
	//		Slice:          input.Slice,
	//		Length:         input.Length,
	//	},
	//	ReadInternal: input.ReadInternal,
	//	SeekInternal: input.SeekInternal,
	//}
	input.IndexInputDefault = NewIndexInputDefault(&IndexInputDefaultConfig{
		Reader:         input,
		Close:          input.Close,
		GetFilePointer: input.GetFilePointer,
		Seek:           input.Seek,
		Slice:          input.Slice,
		Length:         input.Length,
	})
	return input
}

//func (n *NIOFSIndexInput) ReadByte() (byte, error) {
//	bs := [1]byte{}
//	_, err := n.file.ReadAt(bs[:], n.pointer)
//	if err != nil {
//		return 0, err
//	}
//	n.pointer++
//	return bs[0], nil
//}
//
//func (n *NIOFSIndexInput) Read(b []byte) (int, error) {
//	num, err := n.file.ReadAt(b, n.pointer)
//	if err != nil {
//		if err != io.isEof {
//			return 0, err
//		}
//	}
//	n.pointer += int64(num)
//	return len(b), err
//}

func (n *NIOFSIndexInput) Close() error {
	if n.isClone {
		return nil
	}
	return n.file.Close()
}

func (n *NIOFSIndexInput) Seek(pos int64, whence int) (int64, error) {
	n.pos = pos
	return n.file.Seek(pos, io.SeekStart)
}

func (n *NIOFSIndexInput) Length() int64 {
	return n.end - n.off
}

func (n *NIOFSIndexInput) Slice(sliceDescription string, offset, length int64) (IndexInput, error) {
	return NewNIOFSIndexInputV1(n.file, offset, length), nil
}

func (n *NIOFSIndexInput) GetFilePointer() int64 {
	return n.pos
}
