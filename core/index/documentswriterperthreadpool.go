package index

// DocumentsWriterPerThreadPool controls DocumentsWriterPerThread instances and their thread assignments
// during indexing. Each DocumentsWriterPerThread is once a obtained from the pool exclusively used for
// indexing a single document or list of documents by the obtaining thread. Each indexing thread must
// obtain such a DocumentsWriterPerThread to make progress. Depending on the DocumentsWriterPerThreadPool
// implementation DocumentsWriterPerThread assignments might differ from document to document.
// Once a DocumentsWriterPerThread is selected for Flush the DocumentsWriterPerThread will be checked out
// of the thread pool and won't be reused for indexing. See checkout(DocumentsWriterPerThread).
type DocumentsWriterPerThreadPool struct {
}
