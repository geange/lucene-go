package store

import (
	"errors"
	"fmt"
	"go.uber.org/atomic"
	"hash"
	"sort"
)

var _ BaseDirectory = &RAMDirectory{}

// RAMDirectory A memory-resident Directory implementation. Locking implementation is by default the
// SingleInstanceLockFactory.
// Warning: This class is not intended to work with huge indexes. Everything beyond several hundred
// megabytes will waste resources (GC cycles), because it uses an internal buffer size of 1024 bytes,
// producing millions of byte[1024] arrays. This class is optimized for small memory-resident indexes.
// It also has bad concurrency on multithreaded environments.
// It is recommended to materialize large indexes on disk and use MMapDirectory, which is a
// high-performance directory implementation working directly on the file system cache of the
// operating system, so copying data to Java heap space is not useful.
// Deprecated
// This class uses inefficient synchronization and is discouraged in favor of MMapDirectory. It will
// be removed in future versions of Lucene.
type RAMDirectory struct {
	*BaseDirectoryBase

	fileMap     map[string]*RAMFile
	sizeInBytes int64

	// Used to generate temp file names in createTempOutput.
	nextTempFileCounter atomic.Int64

	digest hash.Hash32
}

func NewRAMDirectory() *RAMDirectory {
	return &RAMDirectory{}
}

func (r *RAMDirectory) ListAll() ([]string, error) {
	err := r.EnsureOpen()
	if err != nil {
		return nil, err
	}

	names := make([]string, 0)
	for s, _ := range r.fileMap {
		names = append(names, s)
	}

	sort.Strings(names)

	return names, nil
}

func (r *RAMDirectory) DeleteFile(name string) error {
	err := r.EnsureOpen()
	if err != nil {
		return err
	}
	_, ok := r.fileMap[name]
	if ok {
		delete(r.fileMap, name)
		return nil
	}
	return errors.New("file not found")
}

func (r *RAMDirectory) FileLength(name string) (int64, error) {
	err := r.EnsureOpen()
	if err != nil {
		return 0, err
	}

	if file, ok := r.fileMap[name]; ok {
		return file.GetLength(), nil
	}
	return 0, errors.New("file not found")
}

func (r *RAMDirectory) CreateOutput(name string, context *IOContext) (IndexOutput, error) {
	err := r.EnsureOpen()
	if err != nil {
		return nil, err
	}

	file := NewRAMFile()

	if _, ok := r.fileMap[name]; ok {
		return nil, errors.New("file already exist")
	}
	r.fileMap[name] = file
	return NewRAMOutputStreamV1(name, file, true), nil
}

func (r *RAMDirectory) CreateTempOutput(prefix, suffix string, context *IOContext) (IndexOutput, error) {
	err := r.EnsureOpen()
	if err != nil {
		return nil, err
	}

	file := NewRAMFile()

	// ... then try to find a unique name for it:
	for {
		name := SegmentFileName(prefix,
			fmt.Sprintf("%s_%d", suffix, r.nextTempFileCounter.Inc()), "tmp")

		if _, ok := r.fileMap[name]; !ok {
			return NewRAMOutputStreamV1(name, file, true), nil
		}
	}
}

func (r *RAMDirectory) Sync(names []string) error {
	return nil
}

func (r *RAMDirectory) SyncMetaData() error {
	return nil
}

func (r *RAMDirectory) Rename(source, dest string) error {
	err := r.EnsureOpen()
	if err != nil {
		return err
	}
	_, ok := r.fileMap[source]
	if ok {
		return errors.New("file already exist")
	}

	file, ok := r.fileMap[source]
	if !ok {
		return errors.New("file was unexpectedly replaced")
	}

	r.fileMap[source] = file
	delete(r.fileMap, source)
	return nil
}

func (r *RAMDirectory) OpenInput(name string, context *IOContext) (IndexInput, error) {
	err := r.EnsureOpen()
	if err != nil {
		return nil, err
	}

	file, ok := r.fileMap[name]
	if !ok {
		return nil, errors.New("file not found")
	}
	return NewRAMInputStream(name, file)
}

func (r *RAMDirectory) Close() error {
	r.isOpen = false

	keys := make([]string, 0, len(r.fileMap))
	for k := range r.fileMap {
		keys = append(keys, k)
	}

	for _, key := range keys {
		delete(r.fileMap, key)
	}

	return nil
}

func (r *RAMDirectory) GetPendingDeletions() (map[string]struct{}, error) {
	return map[string]struct{}{}, nil
}
