package store

// RandomAccessInput Random Access Index API. Unlike IndexInput, this has no concept of file position,
// all reads are absolute. However, like IndexInput, it is only intended for use by a single thread.
type RandomAccessInput interface {
}
