package bytesref

// Builder A builder for BytesRef instances.
type Builder struct {
	bytes []byte
}

// NewBytesRefBuilder Sole constructor.
func NewBytesRefBuilder() *Builder {
	return &Builder{bytes: make([]byte, 0, 4096)}
}

func (r *Builder) Get() []byte {
	return r.bytes
}

// Bytes Return a reference to the bytes of this builder.
func (r *Builder) Bytes() []byte {
	return r.bytes
}

// Length Return the number of bytes in this buffer.
func (r *Builder) Length() int {
	return len(r.bytes)
}

// SetLength Set the length.
func (r *Builder) SetLength(length int) {
	if len(r.bytes) < length {
		r.bytes = append(r.bytes, make([]byte, length-len(r.bytes))...)
		return
	}
	r.bytes = r.bytes[:length]
}

// ByteAt Return the byte at the given offset.
func (r *Builder) ByteAt(offset int) byte {
	if offset < len(r.bytes) {
		return r.bytes[offset]
	}
	return 0
}

// SetByteAt Set a byte.
func (r *Builder) SetByteAt(offset int, b byte) {
	if offset < len(r.bytes) {
		r.bytes[offset] = b
	}
}

// Grow Ensure that this builder can hold at least capacity bytes without resizing.
func (r *Builder) Grow(capacity int) {
	if capacity <= 0 {
		return
	}

	r.bytes = append(r.bytes, make([]byte, capacity)...)
}

// AppendByte Append a single byte to this builder.
func (r *Builder) AppendByte(b byte) {
	r.bytes = append(r.bytes, b)
}

// AppendBytes Append the provided bytes to this builder.
func (r *Builder) AppendBytes(b []byte) {
	r.bytes = append(r.bytes, b...)
}

// AppendRef Append the provided bytes to this builder.
//func (r *BytesRefBuilder) AppendRef(ref *BytesRef) {
//	r.AppendBytes(ref.NewBytes, ref.Offset, ref.Len)
//}

// AppendBuilder Append the provided bytes to this builder.
func (r *Builder) AppendBuilder(builder *Builder) {
	r.AppendBytes(builder.Get())
}

func (r *Builder) Clear() {
	r.SetLength(0)
}

// CopyBytes Replace the content of this builder with the provided bytes. Equivalent to calling clear() and
// then append(byte[], int, int).
func (r *Builder) CopyBytes(b []byte, off, length int) {
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
// and then append(Builder).
func (r *Builder) CopyBytesBuilder(builder *Builder) {
	r.Clear()
	r.AppendBuilder(builder)
}
