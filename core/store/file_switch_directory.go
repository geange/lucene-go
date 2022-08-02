package store

var _ Directory = &FileSwitchDirectory{}

// FileSwitchDirectory Expert: A Directory instance that switches files between two other Directory instances.
//Files with the specified extensions are placed in the primary directory; others are placed in the secondary directory. The provided Set must not change once passed to this class, and must allow multiple threads to call contains at once.
//Locks with a name having the specified extensions are delegated to the primary directory; others are delegated to the secondary directory. Ideally, both Directory instances should use the same lock factory
type FileSwitchDirectory struct {
}
