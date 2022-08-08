package store

//var _ FilterDirectory = &LockValidatingDirectoryWrapper{}

// LockValidatingDirectoryWrapper This class makes a best-effort check that a provided Lock is valid
// before any destructive filesystem operation.
type LockValidatingDirectoryWrapper struct {
}
