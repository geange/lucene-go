package core

// BytesRefArray A simple append only random-access BytesRef array that stores full copies of the appended
// bytes in a ByteBlockPool. Note: This class is not Thread-Safe!
type BytesRefArray struct {
}

func (a BytesRefArray) Append(payload []byte) int {
	panic("")
}
