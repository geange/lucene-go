package store

var _ Directory = &FilterDirectory{}

// FilterDirectory Directory implementation that delegates calls to another directory. This class can be used
// to add limitations on top of an existing Directory implementation such as NRTCachingDirectory or to add
// additional sanity checks for tests. However, if you plan to write your own Directory implementation,
// you should consider extending directly Directory or BaseDirectory rather than try to reuse functionality
// of existing Directorys by extending this class.
type FilterDirectory struct {
	*DirectoryDefault
	in Directory
}

func (f *FilterDirectory) ListAll() ([]string, error) {
	return f.in.ListAll()
}

func (f *FilterDirectory) DeleteFile(name string) error {
	return f.in.DeleteFile(name)
}

func (f *FilterDirectory) FileLength(name string) (int64, error) {
	return f.in.FileLength(name)
}

func (f *FilterDirectory) CreateOutput(name string, context *IOContext) (IndexOutput, error) {
	return f.in.CreateOutput(name, context)
}

func (f *FilterDirectory) CreateTempOutput(prefix, suffix string, context *IOContext) (IndexOutput, error) {
	return f.in.CreateTempOutput(prefix, suffix, context)
}

func (f *FilterDirectory) Sync(names []string) error {
	return f.in.Sync(names)
}

func (f *FilterDirectory) SyncMetaData() error {
	return f.in.SyncMetaData()
}

func (f *FilterDirectory) Rename(source, dest string) error {
	return f.in.Rename(source, dest)
}

func (f *FilterDirectory) OpenInput(name string, context *IOContext) (IndexInput, error) {
	return f.in.OpenInput(name, context)
}

func (f *FilterDirectory) ObtainLock(name string) (Lock, error) {
	return f.in.ObtainLock(name)
}

func (f *FilterDirectory) Close() error {
	return f.in.Close()
}

func (f *FilterDirectory) EnsureOpen() error {
	return f.in.EnsureOpen()
}

func (f *FilterDirectory) GetPendingDeletions() (map[string]struct{}, error) {
	return f.in.GetPendingDeletions()
}
