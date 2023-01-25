package index

import (
	"errors"
	"github.com/geange/lucene-go/core/store"
)

// CompoundDirectory A read-only Directory that consists of a view over a compound file.
// See Also: CompoundFormat
// lucene.experimental
type CompoundDirectory interface {
	store.Directory

	// CheckIntegrity Checks consistency of this directory.
	// Note that this may be costly in terms of I/O, e.g. may involve computing a checksum
	// value against large data files.
	CheckIntegrity() error
}

type CompoundDirectoryDefault struct {
}

var (
	ErrUnsupportedOperation = errors.New("unsupported operation exception")
)

func (*CompoundDirectoryDefault) DeleteFile(name string) error {
	return ErrUnsupportedOperation
}

func (*CompoundDirectoryDefault) Rename(source, dest string) error {
	return ErrUnsupportedOperation
}

func (*CompoundDirectoryDefault) SyncMetaData() error {
	return nil
}

func (*CompoundDirectoryDefault) CreateOutput(name string, context *store.IOContext) (store.IndexOutput, error) {
	return nil, ErrUnsupportedOperation
}

func (*CompoundDirectoryDefault) CreateTempOutput(prefix, suffix string,
	context *store.IOContext) (store.IndexOutput, error) {
	return nil, ErrUnsupportedOperation
}

func (*CompoundDirectoryDefault) Sync(names []string) error {
	return ErrUnsupportedOperation
}

func (*CompoundDirectoryDefault) ObtainLock(name string) (store.Lock, error) {
	return nil, ErrUnsupportedOperation
}
