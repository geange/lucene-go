package store

var _ ChecksumIndexInput = &BufferedChecksumIndexInput{}

// BufferedChecksumIndexInput Simple implementation of ChecksumIndexInput that wraps another input and delegates calls.
type BufferedChecksumIndexInput struct {
}
