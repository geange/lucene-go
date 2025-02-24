package lucene50

import (
	"context"

	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
)

var _ index.CompoundDirectory = &CompoundReader{}

type CompoundReader struct {
	directory   store.Directory
	segmentName string
	entries     map[string]*FileEntry
	handle      store.IndexInput
	version     int
}

func NewCompoundReader(directory store.Directory, si index.SegmentInfo) *CompoundReader {
	panic("")
}

type FileEntry struct {
}

func (r *CompoundReader) ListAll(ctx context.Context) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (r *CompoundReader) DeleteFile(ctx context.Context, name string) error {
	//TODO implement me
	panic("implement me")
}

func (r *CompoundReader) FileLength(ctx context.Context, name string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (r *CompoundReader) CreateOutput(ctx context.Context, name string) (store.IndexOutput, error) {
	//TODO implement me
	panic("implement me")
}

func (r *CompoundReader) CreateTempOutput(ctx context.Context, prefix, suffix string) (store.IndexOutput, error) {
	//TODO implement me
	panic("implement me")
}

func (r *CompoundReader) Rename(ctx context.Context, source, dest string) error {
	//TODO implement me
	panic("implement me")
}

func (r *CompoundReader) OpenInput(ctx context.Context, name string) (store.IndexInput, error) {
	//TODO implement me
	panic("implement me")
}

func (r *CompoundReader) ObtainLock(name string) (store.Lock, error) {
	//TODO implement me
	panic("implement me")
}

func (r *CompoundReader) Close() error {
	//TODO implement me
	panic("implement me")
}

func (r *CompoundReader) CopyFrom(ctx context.Context, from store.Directory, src, dest string, ioContext *store.IOContext) error {
	//TODO implement me
	panic("implement me")
}

func (r *CompoundReader) EnsureOpen() error {
	//TODO implement me
	panic("implement me")
}

func (r *CompoundReader) Sync(files map[string]struct{}) error {
	//TODO implement me
	panic("implement me")
}

func (r *CompoundReader) CheckIntegrity() error {
	//TODO implement me
	panic("implement me")
}
