package store

var _ FilterDirectory = &TrackingDirectoryWrapper{}

// TrackingDirectoryWrapper A delegating Directory that records which files were written to and deleted.
type TrackingDirectoryWrapper struct {
}
