package structure

type ByteRef struct {
	Bytes []byte
}

func NewByteRef(values []byte) *ByteRef {
	return &ByteRef{Bytes: values}
}

func (r *ByteRef) Len() int {
	return len(r.Bytes)
}
