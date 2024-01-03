package index

// A FilterLeafReader contains another LeafReader, which it uses as its basic source of data,
// possibly transforming the data along the way or providing additional functionality.
// The class FilterLeafReader itself simply implements all abstract methods of Reader
// with versions that pass all requests to the contained index reader. Subclasses of FilterLeafReader
// may further override some of these methods and may also provide additional methods and fields.
// NOTE: If you override getLiveDocs(), you will likely need to override numDocs() as well and vice-versa.
// NOTE: If this FilterLeafReader does not change the content the contained reader, you could consider
// delegating calls to getCoreCacheHelper() and getReaderCacheHelper().
type FilterLeafReader struct {
}
