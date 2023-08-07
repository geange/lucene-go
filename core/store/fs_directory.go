package store

import (
	"errors"
	"fmt"
	"go.uber.org/atomic"
	"os"
	"path/filepath"
)

// FSDirectory Base class for Directory implementations that store index files in the file system.
// There are currently three core subclasses:
// SimpleFSDirectory is a straightforward implementation using Files.newByteChannel. However, it has poor concurrent performance (multiple threads will bottleneck) as it synchronizes when multiple threads read from the same file.
// NIOFSDirectory uses java.nio's FileChannel's positional io when reading to avoid synchronization when reading from the same file. Unfortunately, due to a Windows-only Sun JRE bug  this is a poor choice for Windows, but on all other platforms this is the preferred choice. Applications using Thread.interrupt() or Future.cancel(boolean) should use RAFDirectory instead. See NIOFSDirectory java doc for details.
// MMapDirectory uses memory-mapped IO when reading. This is a good choice if you have plenty of virtual memory relative to your index size, eg if you are running on a 64 bit JRE, or you are running on a 32 bit JRE but your index sizes are small enough to fit into the virtual memory space. Java has currently the limitation of not being able to unmap files from user code. The files are unmapped, when GC releases the byte buffers. Due to this bug  in Sun's JRE, MMapDirectory's IndexInput.close is unable to close the underlying OS file handle. Only when GC finally collects the underlying objects, which could be quite some time later, will the file handle be closed. This will consume additional transient disk usage: on Windows, attempts to delete or overwrite the files will result in an exception; on other platforms, which typically have a "delete on last close" semantics, while such operations will succeed, the bytes are still consuming space on disk. For many applications this limitation is not a problem (e.g. if you have plenty of disk space, and you don't rely on overwriting files on Windows) but it's still an important limitation to be aware of. This class supplies a (possibly dangerous) workaround mentioned in the bug report, which may fail on non-Sun JVMs.
// Unfortunately, because of system peculiarities, there is no single overall best implementation. Therefore, we've added the open method, to allow Lucene to choose the best FSDirectory implementation given your environment, and the known limitations of each implementation. For users who have no reason to prefer a specific implementation, it's best to simply use open. For all others, you should instantiate the desired implementation directly.
// NOTE: Accessing one of the above subclasses either directly or indirectly from a thread while it's interrupted can close the underlying channel immediately if at the same time the thread is blocked on IO. The channel will remain closed and subsequent access to the index will throw a ClosedChannelException. Applications using Thread.interrupt() or Future.cancel(boolean) should use the slower legacy RAFDirectory from the misc Lucene module instead.
// The locking implementation is by default NativeFSLockFactory, but can be changed by passing in a custom LockFactory instance.
// See Also: Directory
type FSDirectory interface {
	BaseDirectory

	// GetDirectory Returns: the underlying filesystem directory
	GetDirectory() (string, error)
}

//var _ Directory = &FSDirectoryBase{}

type FSDirectoryBase struct {
	*BaseDirectoryBase

	// The underlying filesystem directory
	directory string

	//Maps files that we are trying to delete (or we tried already but failed) before attempting to delete that key.
	pendingDeletes map[string]struct{}

	opsSinceLastDelete *atomic.Int64

	// Used to generate temp file names in createTempOutput.
	nextTempFileCounter *atomic.Int64
}

func NewFSDirectoryBase(path string, factory LockFactory) (*FSDirectoryBase, error) {
	directory, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	return &FSDirectoryBase{
		BaseDirectoryBase:   &BaseDirectoryBase{},
		directory:           directory,
		pendingDeletes:      map[string]struct{}{},
		opsSinceLastDelete:  atomic.NewInt64(0),
		nextTempFileCounter: atomic.NewInt64(0),
	}, nil
}

func (f *FSDirectoryBase) ListAll() ([]string, error) {
	stat, err := os.Stat(f.directory)
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() {
		return nil, errors.New("TODO")
	}
	dir, err := os.ReadDir(f.directory)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(dir))
	for _, entry := range dir {
		names = append(names, entry.Name())
	}
	return names, nil
}

func (f *FSDirectoryBase) DeleteFile(name string) error {
	if _, ok := f.pendingDeletes[name]; ok {
		return fmt.Errorf("file %s is pending delete", name)
	}

	if err := f.privateDeleteFile(name, false); err != nil {
		return err
	}
	return f.maybeDeletePendingFiles()
}

func (f *FSDirectoryBase) FileLength(name string) (int64, error) {
	if _, ok := f.pendingDeletes[name]; ok {
		return 0, fmt.Errorf("file %s is pending delete", name)
	}
	filePath := filepath.Join(f.directory, name)
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	return fileInfo.Size(), nil
}

func (f *FSDirectoryBase) CreateOutput(name string, context *IOContext) (IndexOutput, error) {
	if err := f.EnsureOpen(); err != nil {
		return nil, err
	}

	if err := f.maybeDeletePendingFiles(); err != nil {
		return nil, err
	}

	if _, ok := f.pendingDeletes[name]; ok {
		delete(f.pendingDeletes, name)
		if err := f.privateDeleteFile(name, true); err != nil {
			return nil, err
		}
	}

	return f.NewFSIndexOutput(name)
}

func (f *FSDirectoryBase) CreateTempOutput(prefix, suffix string, context *IOContext) (IndexOutput, error) {
	if err := f.EnsureOpen(); err != nil {
		return nil, err
	}

	if err := f.maybeDeletePendingFiles(); err != nil {
		return nil, err
	}

	for {
		name := getTempFileName(prefix, suffix, f.nextTempFileCounter.Inc())
		if _, ok := f.pendingDeletes[name]; ok {
			continue
		}
		return f.NewFSIndexOutput(name)
	}
}

func (f *FSDirectoryBase) Sync(names []string) error {
	if err := f.EnsureOpen(); err != nil {
		return err
	}

	for _, name := range names {
		if err := f.fsync(name); err != nil {
			return err
		}
	}
	return nil
}

func (f *FSDirectoryBase) fsync(name string) error {
	//  IOUtils.fsync(directory.resolve(name), false);
	// TODO:
	return nil
}

func (f *FSDirectoryBase) SyncMetaData() error {
	// TODO: to improve listCommits(), IndexFileDeleter could call this after deleting segments_Ns
	if err := f.EnsureOpen(); err != nil {
		return err
	}
	//IOUtils.fsync(directory, true);
	return f.maybeDeletePendingFiles()
}

func (f *FSDirectoryBase) Rename(source, dest string) error {
	if err := f.EnsureOpen(); err != nil {
		return err
	}
	if _, ok := f.pendingDeletes[source]; ok {
		return fmt.Errorf("file \"%s\" is pending delete and cannot be moved", source)
	}
	if err := f.maybeDeletePendingFiles(); err != nil {
		return err
	}
	if _, ok := f.pendingDeletes[dest]; ok {
		if err := f.privateDeleteFile(dest, true); err != nil {
			return err
		}
		delete(f.pendingDeletes, dest)
	}
	return os.Rename(f.resolveFilePath(source), f.resolveFilePath(dest))
}

//func (f *FSDirectoryBase) OpenInput(name string, context *IOContext) (IndexInput, error) {
//	//TODO implement me
//	panic("implement me")
//}

func (f *FSDirectoryBase) Close() error {
	f.isOpen = false
	return f.deletePendingFiles()
}

func (f *FSDirectoryBase) EnsureOpen() error {
	return nil
}

func (f *FSDirectoryBase) GetPendingDeletions() (map[string]struct{}, error) {
	if err := f.deletePendingFiles(); err != nil {
		return nil, err
	}

	return f.pendingDeletes, nil
}

func (f *FSDirectoryBase) resolveFilePath(name string) string {
	return filepath.Join(f.directory, name)
}

func (f *FSDirectoryBase) privateDeleteFile(name string, isPendingDelete bool) error {
	delete(f.pendingDeletes, name)
	err := os.Remove(f.resolveFilePath(name))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			delete(f.pendingDeletes, name)
			return err
		}

		f.pendingDeletes[name] = struct{}{}
		return err
	}
	return nil
}

func (f *FSDirectoryBase) maybeDeletePendingFiles() error {
	if len(f.pendingDeletes) > 0 {
		count := int(f.opsSinceLastDelete.Add(1))
		if count >= len(f.pendingDeletes) {
			return f.deletePendingFiles()
		}
	}
	return nil
}

// try to delete any pending files that we had previously tried to delete but failed because we are on
// Windows and the files were still held open.
func (f *FSDirectoryBase) deletePendingFiles() error {

	if len(f.pendingDeletes) == 0 {
		return nil
	}

	for name := range f.pendingDeletes {
		if err := f.privateDeleteFile(name, true); err != nil {
			return err
		}
	}
	return nil
}

var _ IndexOutput = &FSIndexOutput{}

type FSIndexOutput struct {
	*OutputStreamIndexOutput
}

func (f *FSDirectoryBase) NewFSIndexOutput(name string) (*FSIndexOutput, error) {
	file, err := os.OpenFile(f.resolveFilePath(name), os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	return &FSIndexOutput{
		OutputStreamIndexOutput: NewOutputStreamIndexOutput(name, file),
	}, nil
}

func (f *FSDirectoryBase) ensureCanRead(name string) error {
	if _, ok := f.pendingDeletes[name]; ok {
		return fmt.Errorf("file \"%s\" is pending delete and cannot be opened for read", name)
	}
	return nil
}

func (f *FSDirectoryBase) GetDirectory() (string, error) {
	if err := f.EnsureOpen(); err != nil {
		return "", err
	}
	return f.directory, nil
}
