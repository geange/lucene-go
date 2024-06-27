package store

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
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
	sync.Mutex

	open                *atomic.Bool
	lockFactory         LockFactory   // Holds the LockFactory instance (implements locking for this Directory instance).
	dir                 string        // The underlying filesystem directory
	nextTempFileCounter *atomic.Int64 // Used to generate temp file names in createTempOutput.
}

func (n *NIOFSDirectory) Sync(names map[string]struct{}) error {
	return nil
}

func (n *NIOFSDirectory) CopyFrom(ctx context.Context, from Directory, src, dest string, ioContext *IOContext) error {
	return CopyFrom(ctx, n, from, src, dest, ioContext)
}

func NewNIOFSDirectory(path string) (*NIOFSDirectory, error) {
	dirPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	stat, err := os.Stat(dirPath)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}

		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return nil, err
		}
	} else {
		if !stat.IsDir() {
			return nil, fmt.Errorf("%s is not dir", dirPath)
		}
	}

	dir := &NIOFSDirectory{
		open:                &atomic.Bool{},
		lockFactory:         NewSimpleFSLockFactory(),
		dir:                 dirPath,
		nextTempFileCounter: &atomic.Int64{},
	}
	dir.open.Store(true)
	return dir, nil
}

func (n *NIOFSDirectory) OpenInput(ctx context.Context, name string) (IndexInput, error) {
	n.Lock()
	defer n.Unlock()

	if err := n.EnsureOpen(); err != nil {
		return nil, err
	}

	path := n.resolveFilePath(name)

	file, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	return NewNIOFSIndexInput(file)
}

func (n *NIOFSDirectory) ObtainLock(name string) (Lock, error) {
	n.Lock()
	defer n.Unlock()

	return n.lockFactory.ObtainLock(n, name)
}

func (n *NIOFSDirectory) ListAll(context.Context) ([]string, error) {
	n.Lock()
	defer n.Unlock()

	stat, err := os.Stat(n.dir)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
		if err := os.MkdirAll(n.dir, 0755); err != nil {
			return nil, err
		}
		stat, err = os.Stat(n.dir)
		if err != nil {
			return nil, err
		}
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf("%s is not dir", n.dir)
	}

	entries, err := os.ReadDir(n.dir)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names, nil
}

func (n *NIOFSDirectory) DeleteFile(ctx context.Context, name string) error {
	n.Lock()
	defer n.Unlock()

	return n.privateDeleteFile(name)
}

func (n *NIOFSDirectory) FileLength(ctx context.Context, name string) (int64, error) {
	n.Lock()
	defer n.Unlock()

	filePath := filepath.Join(n.dir, name)
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	return fileInfo.Size(), nil
}

func (n *NIOFSDirectory) CreateOutput(ctx context.Context, name string) (IndexOutput, error) {
	n.Lock()
	defer n.Unlock()

	if err := n.EnsureOpen(); err != nil {
		return nil, err
	}

	return n.NewFSIndexOutput(name)
}

func (n *NIOFSDirectory) CreateTempOutput(ctx context.Context, prefix, suffix string) (IndexOutput, error) {
	n.Lock()
	defer n.Unlock()

	if err := n.EnsureOpen(); err != nil {
		return nil, err
	}

	for {
		name := genTempFileName(prefix, suffix, n.nextTempFileCounter.Add(1))
		output, err := n.NewFSIndexOutput(name)
		if err != nil {
			if errors.Is(err, fs.ErrExist) {
				continue
			}
			return nil, err
		}
		return output, nil
	}
}

func (n *NIOFSDirectory) Rename(ctx context.Context, source, dest string) error {
	n.Lock()
	defer n.Unlock()

	if err := n.EnsureOpen(); err != nil {
		return err
	}
	return os.Rename(n.resolveFilePath(source), n.resolveFilePath(dest))
}

func (n *NIOFSDirectory) Close() error {
	n.open.Store(false)
	return nil
}

func (n *NIOFSDirectory) EnsureOpen() error {
	if n.open.Load() {
		return nil
	}
	return errors.New("directory is closed")
}

func (n *NIOFSDirectory) resolveFilePath(name string) string {
	return filepath.Join(n.dir, name)
}

func (n *NIOFSDirectory) privateDeleteFile(name string) error {
	if err := os.Remove(n.resolveFilePath(name)); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	return nil
}

func (n *NIOFSDirectory) GetLockFactory() LockFactory {
	return n.lockFactory
}

var _ IndexOutput = &FSIndexOutput{}

type FSIndexOutput struct {
	*OutputStream
}

func (n *NIOFSDirectory) NewFSIndexOutput(name string) (*FSIndexOutput, error) {
	if _, err := os.Stat(n.resolveFilePath(name)); err == nil {
		return nil, fs.ErrExist
	}

	file, err := os.Create(n.resolveFilePath(name))
	if err != nil {
		return nil, err
	}

	return &FSIndexOutput{
		OutputStream: NewOutputStream(name, file),
	}, nil
}

func (n *NIOFSDirectory) GetDirectory() (string, error) {
	if err := n.EnsureOpen(); err != nil {
		return "", err
	}
	return n.dir, nil
}

var _ IndexInput = &NIOFSIndexInput{}

type NIOFSIndexInput struct {
	*BaseIndexInput

	desc  string
	file  *os.File
	off   int64
	end   int64
	pos   int64
	clone *atomic.Bool
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

func NewNIOFSIndexInput(file *os.File) (*NIOFSIndexInput, error) {
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	input := &NIOFSIndexInput{
		file:  file,
		off:   0,
		pos:   0,
		end:   info.Size(),
		clone: &atomic.Bool{},
	}

	input.BaseIndexInput = NewBaseIndexInput(input)

	return input, nil
}

func newNIOFSIndexInput(desc string, file *os.File, off, length int64) (*NIOFSIndexInput, error) {
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	end := off + length
	if info.Size() < end {
		return nil, errors.New("length exceeds the file size")
	}

	input := &NIOFSIndexInput{
		desc:  desc,
		file:  file,
		off:   off,
		pos:   off,
		end:   end,
		clone: &atomic.Bool{},
	}

	input.clone.Store(true)
	input.BaseIndexInput = NewBaseIndexInput(input)

	return input, nil
}

func (n *NIOFSIndexInput) Clone() CloneReader {
	input := &NIOFSIndexInput{
		file:  n.file,
		clone: &atomic.Bool{},
		off:   n.off,
		pos:   n.pos,
		end:   n.end,
	}

	input.clone.Store(true)
	input.BaseIndexInput = NewBaseIndexInput(input)
	return input
}

func (n *NIOFSIndexInput) Close() error {
	if n.clone.Load() {
		return nil
	}
	return n.file.Close()
}

func (n *NIOFSIndexInput) Seek(pos int64, whence int) (int64, error) {
	nextPos := int64(0)

	switch whence {
	case io.SeekStart:
		nextPos = n.off + pos
	case io.SeekCurrent:
		nextPos = n.pos + pos
	case io.SeekEnd:
		nextPos = n.end - pos
	}

	if nextPos < n.off || nextPos > n.end {
		return 0, errors.New("out of NIOFSIndexInput range")
	}

	n.pos = nextPos
	return n.file.Seek(n.pos, whence)
}

func (n *NIOFSIndexInput) Length() int64 {
	return n.end - n.off
}

func (n *NIOFSIndexInput) Slice(desc string, offset, length int64) (IndexInput, error) {
	return newNIOFSIndexInput(desc, n.file, n.off+offset, length)
}

func (n *NIOFSIndexInput) GetFilePointer() int64 {
	return n.pos - n.off
}
