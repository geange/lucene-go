package util

// BytesRefBuilder A builder for BytesRef instances.
type BytesRefBuilder struct {
	ref *BytesRef
}

// NewBytesRefBuilder Sole constructor.
func NewBytesRefBuilder() *BytesRefBuilder {
	return &BytesRefBuilder{ref: NewBytesRefDefault()}
}

func (r *BytesRefBuilder) Get() *BytesRef {
	return r.ref
}

// Bytes Return a reference to the bytes of this builder.
func (r *BytesRefBuilder) Bytes() []byte {
	return r.ref.Bytes
}

// Length Return the number of bytes in this buffer.
func (r *BytesRefBuilder) Length() int {
	return r.ref.Length
}

// SetLength Set the length.
func (r *BytesRefBuilder) SetLength(length int) {
	r.ref.Length = length
}

// ByteAt Return the byte at the given offset.
func (r *BytesRefBuilder) ByteAt(offset int) byte {
	return r.ref.Bytes[offset]
}

// SetByteAt Set a byte.
func (r *BytesRefBuilder) SetByteAt(offset int, b byte) {
	r.ref.Bytes[offset] = b
}

// Grow Ensure that this builder can hold at least capacity bytes without resizing.
func (r *BytesRefBuilder) Grow(capacity int) {
	r.ref.Bytes = append(r.ref.Bytes, make([]byte, capacity)...)
}

// AppendByte Append a single byte to this builder.
func (r *BytesRefBuilder) AppendByte(b byte) {
	r.Grow(r.ref.Length + 1)
	r.ref.Bytes[r.ref.Length] = b
	r.ref.Length++
}

// AppendBytes Append the provided bytes to this builder.
func (r *BytesRefBuilder) AppendBytes(b []byte, off, length int) {
	r.Grow(r.ref.Length + length)
	copy(r.ref.Bytes[r.ref.Offset:], b[off:off+length])
	r.ref.Length += length
}

// AppendRef Append the provided bytes to this builder.
func (r *BytesRefBuilder) AppendRef(ref *BytesRef) {
	r.AppendBytes(ref.Bytes, ref.Offset, ref.Length)
}

// AppendBuilder Append the provided bytes to this builder.
func (r *BytesRefBuilder) AppendBuilder(builder *BytesRefBuilder) {
	r.AppendRef(builder.Get())
}

func (r *BytesRefBuilder) Clear() {
	r.SetLength(0)
}

// CopyBytes Replace the content of this builder with the provided bytes. Equivalent to calling clear() and
// then append(byte[], int, int).
func (r *BytesRefBuilder) CopyBytes(b []byte, off, length int) {
	r.Clear()
	r.AppendBytes(b, off, length)
}

// CopyBytesRef Replace the content of this builder with the provided bytes. Equivalent to calling clear() and
// then append(BytesRef).
func (r *BytesRefBuilder) CopyBytesRef(ref *BytesRef) {
	r.Clear()
	r.AppendRef(ref)
}

// CopyBytesBuilder Replace the content of this builder with the provided bytes. Equivalent to calling clear()
// and then append(BytesRefBuilder).
func (r *BytesRefBuilder) CopyBytesBuilder(builder *BytesRefBuilder) {
	r.Clear()
	r.AppendBuilder(builder)
}
