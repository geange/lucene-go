package store

import (
	"context"
	"fmt"
	"io"
)

// A Directory provides an abstraction layer for storing a list of files. A directory contains only files
// (no sub-folder hierarchy). Implementing classes must comply with the following:
//
//   - A file in a directory can be created (createOutput), appended to, then closed.
//   - A file open for writing may not be available for read access until the corresponding IndexOutput is closed.
//   - Once a file is created it must only be opened for input (openInput), or deleted (deleteFile). Calling
//     createOutput on an existing file must throw java.nio.file.FileAlreadyExistsException.
//
// See Also: 	* FSDirectory
//   - RAMDirectory
//   - FilterDirectory
type Directory interface {

	// ListAll Returns names of all files stored in this directory. The output must be in sorted
	// (UTF-16, java's String.compareTo) order.
	// Throws: IOException – in case of I/O error
	ListAll(ctx context.Context) ([]string, error)

	// DeleteFile Removes an existing file in the directory. This method must throw either
	// NoSuchFileException or FileNotFoundException if name points to a non-existing file.
	// Params: name – the name of an existing file.
	// Throws: IOException – in case of I/O error
	DeleteFile(ctx context.Context, name string) error

	// FileLength Returns the byte length of a file in the directory. This method must throw either
	// NoSuchFileException or FileNotFoundException if name points to a non-existing file.
	// Params: name – the name of an existing file.
	// Throws: IOException – in case of I/O error
	FileLength(ctx context.Context, name string) (int64, error)

	// CreateOutput Creates a new, empty file in the directory and returns an IndexOutput instance for
	// appending data to this file. This method must throw java.nio.file.FileAlreadyExistsException if
	// the file already exists.
	// Params: name – the name of the file to create.
	// Throws: IOException – in case of I/O error
	CreateOutput(ctx context.Context, name string) (IndexOutput, error)

	// CreateTempOutput Creates a new, empty, temporary file in the directory and returns an IndexOutput
	// instance for appending data to this file. The temporary file name
	// (accessible via IndexOutput.getName()) will start with prefix, end with suffix and have a reserved
	// file extension .tmp.
	CreateTempOutput(ctx context.Context, prefix, suffix string) (IndexOutput, error)

	// Sync Ensures that any writes to these files are moved to stable storage (made durable).
	// Lucene uses this to properly commit changes to the index, to prevent a machine/OS crash
	// from corrupting the index.
	// See Also: syncMetaData()
	//Sync(ctx context.Context, names []string) error

	// SyncMetaData Ensures that directory metadata, such as recent file renames,
	// are moved to stable storage.
	// See Also: sync(Collection)
	//SyncMetaData(ctx context.Context) error

	// Rename Renames source file to dest file where dest must not already exist in the directory.
	// It is permitted for this operation to not be truly atomic, for example both source and dest can
	// be visible temporarily in listAll(). However, the implementation of this method must ensure the
	// content of dest appears as the entire source atomically. So once dest is visible for readers,
	// the entire content of previous source is visible. This method is used by IndexWriter to publish commits.
	Rename(ctx context.Context, source, dest string) error

	// OpenInput Opens a stream for reading an existing file. This method must throw either
	// NoSuchFileException or FileNotFoundException if name points to a non-existing file.
	// Params: name – the name of an existing file.
	// Throws: IOException – in case of I/O error
	OpenInput(ctx context.Context, name string) (IndexInput, error)

	// OpenChecksumInput Opens a checksum-computing stream for reading an existing file. This method must
	// throw either NoSuchFileException or FileNotFoundException if name points to a non-existing file.
	// Params: name – the name of an existing file.
	// Throws: IOException – in case of I/O error
	//OpenChecksumInput(name string, context *IOContext) (ChecksumIndexInput, error)

	// ObtainLock Acquires and returns a Lock for a file with the given name.
	// Params: name – the name of the lock file
	// Throws:  LockObtainFailedException – (optional specific exception) if the lock could not be obtained
	//			because it is currently held elsewhere.
	// IOException – if any i/o error occurs attempting to gain the lock
	ObtainLock(name string) (Lock, error)

	// Closer Closes the directory.
	io.Closer

	// CopyFrom Copies an existing src file from directory from to a non-existent file dest in this directory.
	//CopyFrom(from Directory, src, dest string, context *IOContext) error

	// EnsureOpen Ensures this directory is still open.
	// Throws: AlreadyClosedException – if this directory is closed.
	EnsureOpen() error

	// GetPendingDeletions Returns a set of files currently pending deletion in this directory.
	//GetPendingDeletions() (map[string]struct{}, error)
}

func OpenChecksumInput(dir Directory, name string) (ChecksumIndexInput, error) {
	input, err := dir.OpenInput(nil, name)
	if err != nil {
		return nil, err
	}
	return NewBufferedChecksumIndexInput(input), nil
}

type DirectoryDefault struct {
	DeleteFile func(ctx context.Context, name string) error
}

type DirectorySPI interface {
	DeleteFile(ctx context.Context, name string) error
}

func (d *DirectoryDefault) CopyFrom(ctx context.Context, from Directory, src, dest string) error {
	is, err := from.OpenInput(ctx, src)
	if err != nil {
		return err
	}

	os, err := from.CreateOutput(ctx, dest)
	if err != nil {
		return err
	}

	if err := os.CopyBytes(ctx, is, int(is.Length())); err != nil {
		// IOUtils.deleteFilesIgnoringExceptions(this, dest)
		// TODO: 删除目标文件
		return err
	}

	if err := d.DeleteFile(ctx, dest); err != nil {
		return err
	}

	return nil
}

// Creates a file name for a temporary file. The name will start with prefix, end with suffix and have a reserved file extension .tmp.
// See Also: createTempOutput(String, String)
func genTempFileName(prefix, suffix string, counter int64) string {
	return fmt.Sprintf("%s_%s_%d.tmp", prefix, suffix, counter)
}

type FSDirectory interface {
	Directory

	// GetDirectory Returns: the underlying filesystem directory
	GetDirectory() (string, error)
}
