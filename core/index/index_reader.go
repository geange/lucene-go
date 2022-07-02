package index

// IndexReader is an abstract class, providing an interface for accessing a point-in-time view of an index. Any changes made to the index via IndexWriter will not be visible until a new IndexReader is opened. It's best to use DirectoryReader.open(IndexWriter) to obtain an IndexReader, if your IndexWriter is in-process. When you need to re-open to see changes to the index, it's best to use DirectoryReader.openIfChanged(DirectoryReader) since the new reader will share resources with the previous one when possible. Search of an index is done entirely through this abstract interface, so that any subclass which implements it is searchable.
// There are two different types of IndexReaders:
// LeafReader: These indexes do not consist of several sub-readers, they are atomic. They support retrieval of stored fields, doc values, terms, and postings.
// CompositeReader: Instances (like DirectoryReader) of this reader can only be used to get stored fields from the underlying LeafReaders, but it is not possible to directly retrieve postings. To do that, get the sub-readers via CompositeReader.getSequentialSubReaders.
// IndexReader instances for indexes on disk are usually constructed with a call to one of the static DirectoryReader.open() methods, e.g. DirectoryReader.open(org.apache.lucene.store.Directory). DirectoryReader implements the CompositeReader interface, it is not possible to directly get postings.
// For efficiency, in this API documents are often referred to via document numbers, non-negative integers which each name a unique document in the index. These document numbers are ephemeral -- they may change as documents are added to and deleted from an index. Clients should thus not rely on a given document having the same number between sessions.
//
// NOTE: IndexReader instances are completely thread safe, meaning multiple threads can call any of its methods, concurrently. If your application requires external synchronization, you should not synchronize on the IndexReader instance; use your own (non-Lucene) objects instead.
type IndexReader interface {
}
