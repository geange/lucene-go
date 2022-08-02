package store

var (
	_ IndexInput        = &ByteBuffersIndexInput{}
	_ RandomAccessInput = &ByteBuffersIndexInput{}
)

// ByteBuffersIndexInput An IndexInput implementing RandomAccessInput and backed by a ByteBuffersDataInput.
type ByteBuffersIndexInput struct {
}
