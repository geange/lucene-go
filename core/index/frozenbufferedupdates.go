package index

// FrozenBufferedUpdates
// Holds buffered deletes and updates by term or query, once pushed. Pushed deletes/updates are write-once,
// so we shift to more memory efficient data structure to hold them. We don't hold docIDs because these are
// applied on flush.
type FrozenBufferedUpdates struct {
}
