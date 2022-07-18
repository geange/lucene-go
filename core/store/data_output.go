package store

// DataOutput Abstract base class for performing write operations of Lucene's low-level data types.
// DataOutput may only be used from one thread, because it is not thread safe (it keeps internal state like file position).
type DataOutput interface {
}
