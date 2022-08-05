package store

// ChecksumIndexInput Extension of IndexInput, computing checksum as it goes. Callers can retrieve the checksum via getChecksum().
type ChecksumIndexInput interface {
	IndexInput

	// GetChecksum Returns the current checksum value
	GetChecksum() uint32
}
