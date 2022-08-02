package store

var _ FSDirectory = &MMapDirectory{}

// MMapDirectory File-based Directory implementation that uses mmap for reading, and FSDirectory.FSIndexOutput for writing.
// NOTE: memory mapping uses up a portion of the virtual memory address space in your process equal to the size of the file being mapped. Before using this class, be sure your have plenty of virtual address space, e.g. by using a 64 bit JRE, or a 32 bit JRE with indexes that are guaranteed to fit within the address space. On 32 bit platforms also consult MMapDirectory(Path, LockFactory, int) if you have problems with mmap failing because of fragmented address space. If you get an OutOfMemoryException, it is recommended to reduce the chunk size, until it works.
// Due to this bug  in Sun's JRE, MMapDirectory's IndexInput.close is unable to close the underlying OS file handle. Only when GC finally collects the underlying objects, which could be quite some time later, will the file handle be closed.
// This will consume additional transient disk usage: on Windows, attempts to delete or overwrite the files will result in an exception; on other platforms, which typically have a "delete on last close" semantics, while such operations will succeed, the bytes are still consuming space on disk. For many applications this limitation is not a problem (e.g. if you have plenty of disk space, and you don't rely on overwriting files on Windows) but it's still an important limitation to be aware of.
// This class supplies the workaround mentioned in the bug report (see setUseUnmap), which may fail on non-Oracle/OpenJDK JVMs. It forcefully unmaps the buffer on close by using an undocumented internal cleanup functionality. If UNMAP_SUPPORTED is true, the workaround will be automatically enabled (with no guarantees; if you discover any problems, you can disable it).
// NOTE: Accessing this class either directly or indirectly from a thread while it's interrupted can close the underlying channel immediately if at the same time the thread is blocked on IO. The channel will remain closed and subsequent access to MMapDirectory will throw a ClosedChannelException. If your application uses either Thread.interrupt() or Future.cancel(boolean) you should use the legacy RAFDirectory from the Lucene misc module in favor of MMapDirectory.
// See Also: Blog post about MMapDirectory
type MMapDirectory struct {
}
