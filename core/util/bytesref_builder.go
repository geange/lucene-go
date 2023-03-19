package util

// BytesRefBuilder A builder for BytesRef instances.
type BytesRefBuilder struct {
	bytes []byte
}

// NewBytesRefBuilder Sole constructor.
func NewBytesRefBuilder() *BytesRefBuilder {
	return &BytesRefBuilder{bytes: make([]byte, 0, 4096)}
}

func (r *BytesRefBuilder) Get() []byte {
	return r.bytes
}

// Bytes Return a reference to the bytes of this builder.
func (r *BytesRefBuilder) Bytes() []byte {
	return r.bytes
}

// Length Return the number of bytes in this buffer.
func (r *BytesRefBuilder) Length() int {
	return len(r.bytes)
}

// SetLength Set the length.
func (r *BytesRefBuilder) SetLength(length int) {
	if len(r.bytes) < length {
		r.bytes = append(r.bytes, make([]byte, length-len(r.bytes))...)
		return
	}
	r.bytes = r.bytes[:length]
}

// ByteAt Return the byte at the given offset.
func (r *BytesRefBuilder) ByteAt(offset int) byte {
	if offset < len(r.bytes) {
		return r.bytes[offset]
	}
	return 0
}

// SetByteAt Set a byte.
func (r *BytesRefBuilder) SetByteAt(offset int, b byte) {
	if offset < len(r.bytes) {
		r.bytes[offset] = b
	}
}

// Grow Ensure that this builder can hold at least capacity bytes without resizing.
func (r *BytesRefBuilder) Grow(capacity int) {
	r.bytes = append(r.bytes, make([]byte, capacity)...)
}

// AppendByte Append a single byte to this builder.
func (r *BytesRefBuilder) AppendByte(b byte) {
	r.bytes = append(r.bytes, b)
}

// AppendBytes Append the provided bytes to this builder.
func (r *BytesRefBuilder) AppendBytes(b []byte) {
	r.bytes = append(r.bytes, b...)
}

// AppendRef Append the provided bytes to this builder.
//func (r *BytesRefBuilder) AppendRef(ref *BytesRef) {
//	r.AppendBytes(ref.Bytes, ref.Offset, ref.Len)
//}

// AppendBuilder Append the provided bytes to this builder.
func (r *BytesRefBuilder) AppendBuilder(builder *BytesRefBuilder) {
	r.AppendBytes(builder.Get())
}

func (r *BytesRefBuilder) Clear() {
	r.SetLength(0)
}

// CopyBytes Replace the content of this builder with the provided bytes. Equivalent to calling clear() and
// then append(byte[], int, int).
func (r *BytesRefBuilder) CopyBytes(b []byte, off, length int) {
	r.Clear()
	r.AppendBytes(b[off : off+length])
}

// CopyBytesRef Replace the content of this builder with the provided bytes. Equivalent to calling clear() and
// then append(BytesRef).
//func (r *BytesRefBuilder) CopyBytesRef(ref *BytesRef) {
//	r.Clear()
//	r.AppendRef(ref)
//}

// CopyBytesBuilder Replace the content of this builder with the provided bytes. Equivalent to calling clear()
// and then append(BytesRefBuilder).
func (r *BytesRefBuilder) CopyBytesBuilder(builder *BytesRefBuilder) {
	r.Clear()
	r.AppendBuilder(builder)
}
