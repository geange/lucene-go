package structure

type BytesRef struct {
	Bytes []byte
}

func NewBytesRef(values []byte) *BytesRef {
	return &BytesRef{Bytes: values}
}

func (r *BytesRef) Len() int {
	return len(r.Bytes)
}
