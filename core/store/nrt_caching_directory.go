package store

// NRTCachingDirectory Wraps a RAMDirectory around any provided delegate directory, to be used during NRT search.
// This class is likely only useful in a near-real-time context, where indexing rate is lowish but reopen rate is highish, resulting in many tiny files being written. This directory keeps such segments (as well as the segments produced by merging them, as long as they are small enough), in RAM.
// This is safe to use: when your app calls {IndexWriter#commit}, all cached files will be flushed from the cached and sync'd.
// Here's a simple example usage:
//
//	Directory fsDir = FSDirectory.open(new File("/path/to/index").toPath());
//	NRTCachingDirectory cachedFSDir = new NRTCachingDirectory(fsDir, 5.0, 60.0);
//	IndexWriterConfig conf = new IndexWriterConfig(analyzer);
//	IndexWriter writer = new IndexWriter(cachedFSDir, conf);
//
// This will cache all newly flushed segments, all merges whose expected segment size is <= 5 MB, unless the net cached bytes exceeds 60 MB at which point all writes will not be cached (until the net bytes falls below 60 MB).
type NRTCachingDirectory struct {
}

func (n *NRTCachingDirectory) ListAll() ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (n *NRTCachingDirectory) DeleteFile(name string) error {
	//TODO implement me
	panic("implement me")
}

func (n *NRTCachingDirectory) FileLength(name string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (n *NRTCachingDirectory) CreateOutput(name string, context *IOContext) (IndexOutput, error) {
	//TODO implement me
	panic("implement me")
}

func (n *NRTCachingDirectory) CreateTempOutput(prefix, suffix string, context *IOContext) (IndexOutput, error) {
	//TODO implement me
	panic("implement me")
}

func (n *NRTCachingDirectory) Sync(names []string) error {
	//TODO implement me
	panic("implement me")
}

func (n *NRTCachingDirectory) SyncMetaData() error {
	//TODO implement me
	panic("implement me")
}

func (n *NRTCachingDirectory) Rename(source, dest string) error {
	//TODO implement me
	panic("implement me")
}

func (n *NRTCachingDirectory) OpenInput(name string, context *IOContext) (IndexInput, error) {
	//TODO implement me
	panic("implement me")
}

func (n *NRTCachingDirectory) ObtainLock(name string) (Lock, error) {
	//TODO implement me
	panic("implement me")
}

func (n *NRTCachingDirectory) Close() error {
	//TODO implement me
	panic("implement me")
}

func (n *NRTCachingDirectory) EnsureOpen() error {
	//TODO implement me
	panic("implement me")
}

func (n *NRTCachingDirectory) GetPendingDeletions() (map[string]struct{}, error) {
	//TODO implement me
	panic("implement me")
}
