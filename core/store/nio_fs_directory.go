package store

// NIOFSDirectory An FSDirectory implementation that uses java.nio's FileChannel's positional read, which allows multiple threads to read from the same file without synchronizing.
// This class only uses FileChannel when reading; writing is achieved with FSDirectory.FSIndexOutput.
// NOTE: NIOFSDirectory is not recommended on Windows because of a bug in how FileChannel.read is implemented in Sun's JRE. Inside of the implementation the position is apparently synchronized. See here  for details.
// NOTE: Accessing this class either directly or indirectly from a thread while it's interrupted can close the underlying file descriptor immediately if at the same time the thread is blocked on IO. The file descriptor will remain closed and subsequent access to NIOFSDirectory will throw a ClosedChannelException. If your application uses either Thread.interrupt() or Future.cancel(boolean) you should use the legacy RAFDirectory from the Lucene misc module in favor of NIOFSDirectory.
type NIOFSDirectory interface {
	FSDirectory
}
