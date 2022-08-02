package store

// ByteBufferGuard A guard that is created for every ByteBufferIndexInput that tries on best effort to reject
// any access to the ByteBuffer behind, once it is unmapped. A single instance of this is used for the original
// and all clones, so once the original is closed and unmapped all clones also throw AlreadyClosedException,
// triggered by a NullPointerException.
// This code tries to hopefully flush any CPU caches using a store-store barrier. It also yields the current
// thread to give other threads a chance to finish in-flight requests...
type ByteBufferGuard struct {
}
