package store

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"sync/atomic"

	"golang.org/x/exp/maps"
)

var _ Directory = &RAMDirectory{}

type RAMDirectory struct {
	sync.RWMutex

	fileMap map[string]*RAMFile

	// Used to generate temp file names in createTempOutput.
	nextTempFileCounter *atomic.Int64
}

func (d *RAMDirectory) Sync(files map[string]struct{}) error {
	return nil
}

func (d *RAMDirectory) ListAll(ctx context.Context) ([]string, error) {
	files := maps.Keys(d.fileMap)
	return files, nil
}

func (d *RAMDirectory) DeleteFile(ctx context.Context, name string) error {
	if _, ok := d.fileMap[name]; !ok {
		return os.ErrNotExist
	}
	delete(d.fileMap, name)
	return nil
}

func (d *RAMDirectory) FileLength(ctx context.Context, name string) (int64, error) {
	file, ok := d.fileMap[name]
	if !ok {
		return 0, os.ErrNotExist
	}
	return file.GetLength(), nil
}

func (d *RAMDirectory) CreateOutput(ctx context.Context, name string) (IndexOutput, error) {
	if _, ok := d.fileMap[name]; ok {
		return nil, os.ErrExist
	}

	file := NewRAMFile(d)
	d.fileMap[name] = file
	return NewRAMOutputStream(name, file, true), nil
}

func (d *RAMDirectory) CreateTempOutput(ctx context.Context, prefix, suffix string) (IndexOutput, error) {
	// Make the file first...
	file := NewRAMFile(d)

	// ... then try to find a unique name for it:
	for {
		suffixName := suffix + "_" + strconv.FormatInt(d.nextTempFileCounter.Add(1), 36)
		name := SegmentFileName(prefix, suffixName, "tmp")
		if _, ok := d.fileMap[name]; !ok {
			d.fileMap[name] = file
			return NewRAMOutputStream(name, file, true), nil
		}
	}
}

func (d *RAMDirectory) Rename(ctx context.Context, source, dest string) error {
	d.Lock()
	defer d.Unlock()

	if _, ok := d.fileMap[dest]; ok {
		return fmt.Errorf("dest file:%s exist", dest)
	}

	file, ok := d.fileMap[source]
	if !ok {
		return fmt.Errorf("source file:%s not exist", source)
	}

	d.fileMap[dest] = file
	delete(d.fileMap, source)

	return nil
}

func (d *RAMDirectory) OpenInput(ctx context.Context, name string) (IndexInput, error) {
	d.RLock()
	defer d.RUnlock()

	file, ok := d.fileMap[name]
	if !ok {
		return nil, fmt.Errorf("file:%s not found", name)
	}
	return NewRAMInputStream(name, file, int(file.length.Load()))
}

func (d *RAMDirectory) ObtainLock(name string) (Lock, error) {
	return &NoLock{}, nil
}

func (d *RAMDirectory) Close() error {
	clear(d.fileMap)
	return nil
}

func (d *RAMDirectory) CopyFrom(ctx context.Context, from Directory, src, dest string, ioContext *IOContext) error {
	d.Lock()
	defer d.Unlock()

	fromDir, ok := from.(*RAMDirectory)
	if !ok {
		return fmt.Errorf("fromDir is not *RAMDirectory")
	}

	if _, ok := fromDir.fileMap[src]; !ok {
		return fmt.Errorf("src:%s not found", src)
	}

	if _, ok := d.fileMap[dest]; ok {
		return fmt.Errorf("dest:%s exist", dest)
	}

	file := fromDir.fileMap[src]
	newFile := file.Clone()
	d.fileMap[dest] = newFile
	return nil
}

func (d *RAMDirectory) EnsureOpen() error {
	return nil
}
