package store

// FilterDirectory Directory implementation that delegates calls to another directory. This class can be used
// to add limitations on top of an existing Directory implementation such as NRTCachingDirectory or to add
// additional sanity checks for tests. However, if you plan to write your own Directory implementation,
// you should consider extending directly Directory or BaseDirectory rather than try to reuse functionality
// of existing Directorys by extending this class.
type FilterDirectory interface {
	Directory
}
