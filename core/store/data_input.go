package store

import "io"

// DataInput Abstract base class for performing read operations of Lucene's low-level data types.
// DataInput may only be used from one thread, because it is not thread safe (it keeps internal state
// like file position). To allow multithreaded use, every DataInput instance must be cloned before used
// in another thread. Subclasses must therefore implement clone(), returning a new DataInput which operates
// on the same underlying resource, but positioned independently.
type DataInput interface {
	io.Closer
}
