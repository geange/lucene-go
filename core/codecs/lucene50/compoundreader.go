package lucene50

import (
	"context"
	"errors"
	"fmt"

	"github.com/samber/lo"

	"github.com/geange/lucene-go/core/codecs"
	index2 "github.com/geange/lucene-go/core/index"
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

// NewCompoundReader
// Create a new CompoundFileDirectory.
func NewCompoundReader(ctx context.Context, directory store.Directory, si index.SegmentInfo) (*CompoundReader, error) {
	segmentName := si.Name()
	reader := CompoundReader{
		directory:   directory,
		segmentName: segmentName,
	}

	dataFileName := store.SegmentFileName(segmentName, "", COMPOUND_FORMAT_DATA_EXTENSION)
	entriesFileName := store.SegmentFileName(segmentName, "", COMPOUND_FORMAT_ENTRIES_EXTENSION)

	entries, err := reader.readEntries(ctx, si.GetID(), directory, entriesFileName)
	if err != nil {
		return nil, err
	}
	reader.entries = entries

	expectedLength := codecs.IndexHeaderLength(COMPOUND_FORMAT_DATA_CODEC, "")
	for _, entry := range reader.entries {
		expectedLength += entry.Length
	}
	expectedLength += codecs.FooterLength()

	handle, err := directory.OpenInput(ctx, dataFileName)
	if err != nil {
		return nil, err
	}
	reader.handle = handle

	if _, err := codecs.CheckIndexHeader(ctx, handle, COMPOUND_FORMAT_DATA_CODEC,
		reader.version, reader.version, si.GetID(), ""); err != nil {
		return nil, err
	}

	// NOTE: data file is too costly to verify checksum against all the bytes on open,
	// but for now we at least verify proper structure of the checksum footer: which looks
	// for FOOTER_MAGIC + algorithmID. This is cheap and can detect some forms of corruption
	// such as file truncation.
	if _, err := codecs.RetrieveChecksum(ctx, handle); err != nil {
		return nil, err
	}

	// We also validate length, because e.g. if you strip 16 bytes off the .cfs we otherwise
	// would not detect it:
	if int(handle.Length()) != expectedLength {
		return nil, fmt.Errorf("length should be %d", expectedLength)
	}

	return &reader, nil
}

func (r *CompoundReader) readEntries(ctx context.Context, segmentID []byte, dir store.Directory, entriesFileName string) (map[string]*FileEntry, error) {

	entriesStream, err := store.OpenChecksumInput(ctx, dir, entriesFileName)
	if err != nil {
		return nil, err
	}

	version, err := codecs.CheckIndexHeader(ctx, entriesStream, COMPOUND_FORMAT_ENTRY_CODEC,
		COMPOUND_FORMAT_VERSION_START, COMPOUND_FORMAT_VERSION_CURRENT, segmentID, "")
	if err != nil {
		return nil, err
	}
	r.version = version

	numEntries, err := entriesStream.ReadUvarint(ctx)
	if err != nil {
		return nil, err
	}
	mapping := make(map[string]*FileEntry, numEntries)

	for i := 0; i < int(numEntries); i++ {
		id, err := entriesStream.ReadString(ctx)
		if err != nil {
			return nil, err
		}
		_, exist := mapping[id]
		if exist {
			return nil, fmt.Errorf("duplicate cfs entry id=%s", id)
		}

		offset, err := entriesStream.ReadUint64(ctx)
		if err != nil {
			return nil, err
		}
		length, err := entriesStream.ReadUint64(ctx)
		if err != nil {
			return nil, err
		}

		mapping[id] = &FileEntry{
			Offset: int(offset),
			Length: int(length),
		}
	}

	if _, err := codecs.CheckFooter(ctx, entriesStream); err != nil {
		return nil, err
	}
	return mapping, nil
}

type FileEntry struct {
	Offset int
	Length int
}

// ListAll
// Returns an array of strings, one for each file in the directory.
func (r *CompoundReader) ListAll(ctx context.Context) ([]string, error) {
	res := lo.Keys(r.entries)

	for i := range res {
		res[i] = r.segmentName + res[i]
	}
	return res, nil
}

// FileLength
// Returns the length of a file in the directory.
func (r *CompoundReader) FileLength(ctx context.Context, name string) (int64, error) {
	e, ok := r.entries[index2.StripSegmentName(name)]
	if !ok {
		return 0, errors.New("file no exist")
	}
	return int64(e.Length), nil
}

func (r *CompoundReader) OpenInput(ctx context.Context, name string) (store.IndexInput, error) {
	id := index2.StripSegmentName(name)
	entry, ok := r.entries[id]
	if !ok {
		store.SegmentFileName(r.segmentName, "", COMPOUND_FORMAT_DATA_EXTENSION)
		return nil, fmt.Errorf("no sub-file with id:'%s' found in compound file", name)
	}
	return r.handle.Slice(name, int64(entry.Offset), int64(entry.Length))
}

func (r *CompoundReader) Close() error {
	return r.handle.Close()
}

func (r *CompoundReader) DeleteFile(ctx context.Context, name string) error {
	return errors.New("unsupported operation")
}

func (r *CompoundReader) CreateOutput(ctx context.Context, name string) (store.IndexOutput, error) {
	return nil, errors.New("unsupported operation")
}

func (r *CompoundReader) CreateTempOutput(ctx context.Context, prefix, suffix string) (store.IndexOutput, error) {
	return nil, errors.New("unsupported operation")
}

func (r *CompoundReader) Rename(ctx context.Context, source, dest string) error {
	return errors.New("unsupported operation")
}

func (r *CompoundReader) ObtainLock(name string) (store.Lock, error) {
	return nil, errors.New("unsupported operation")
}

func (r *CompoundReader) CopyFrom(ctx context.Context, from store.Directory, src, dest string, ioContext *store.IOContext) error {
	return errors.New("unsupported operation")
}

func (r *CompoundReader) EnsureOpen() error {
	return nil
}

func (r *CompoundReader) Sync(files map[string]struct{}) error {
	return errors.New("unsupported operation")
}

func (r *CompoundReader) CheckIntegrity() error {
	_, err := codecs.ChecksumEntireFile(context.Background(), r.handle)
	return err
}
