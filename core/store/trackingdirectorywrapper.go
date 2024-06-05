package store

import (
	"context"
	"sync"
)

var _ Directory = &TrackingDirectoryWrapper{}

type TrackingDirectoryWrapper struct {
	sync.RWMutex

	Directory

	createdFileNames map[string]struct{}
}

func NewTrackingDirectoryWrapper(directory Directory) *TrackingDirectoryWrapper {
	return &TrackingDirectoryWrapper{
		RWMutex:          sync.RWMutex{},
		Directory:        directory,
		createdFileNames: make(map[string]struct{}),
	}
}

func (t *TrackingDirectoryWrapper) DeleteFile(ctx context.Context, name string) error {
	if err := t.Directory.DeleteFile(ctx, name); err != nil {
		return err
	}

	t.Lock()
	defer t.Unlock()

	delete(t.createdFileNames, name)

	return nil
}

func (t *TrackingDirectoryWrapper) CreateOutput(ctx context.Context, name string) (IndexOutput, error) {
	output, err := t.Directory.CreateOutput(ctx, name)
	if err != nil {
		return nil, err
	}

	t.Lock()
	defer t.Unlock()

	t.createdFileNames[output.GetName()] = struct{}{}

	return output, nil
}

func (t *TrackingDirectoryWrapper) CreateTempOutput(ctx context.Context, prefix, suffix string) (IndexOutput, error) {
	tempOutput, err := t.Directory.CreateTempOutput(ctx, prefix, suffix)
	if err != nil {
		return nil, err
	}

	t.Lock()
	defer t.Unlock()

	t.createdFileNames[tempOutput.GetName()] = struct{}{}

	return tempOutput, nil
}

func (t *TrackingDirectoryWrapper) CopyFrom(ctx context.Context, from Directory, src, dest string, ioContext *IOContext) error {
	if err := t.Directory.CopyFrom(ctx, from, src, dest, ioContext); err != nil {
		return err
	}

	t.Lock()
	defer t.Unlock()

	t.createdFileNames[dest] = struct{}{}
	return nil
}

func (t *TrackingDirectoryWrapper) Rename(ctx context.Context, source, dest string) error {
	if err := t.Directory.Rename(ctx, source, dest); err != nil {
		return err
	}

	t.Lock()
	defer t.Unlock()

	t.createdFileNames[dest] = struct{}{}
	delete(t.createdFileNames, source)

	return nil
}

func (t *TrackingDirectoryWrapper) GetCreatedFiles() map[string]struct{} {
	t.RLock()
	defer t.RUnlock()

	return t.createdFileNames
}

func (t *TrackingDirectoryWrapper) ClearCreatedFiles() {
	t.Lock()
	defer t.Unlock()

	clear(t.createdFileNames)
}
