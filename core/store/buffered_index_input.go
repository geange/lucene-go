package store

// BufferedIndexInput Base implementation class for buffered IndexInput.
type BufferedIndexInput interface {
	IndexInput
	RandomAccessInput
}
