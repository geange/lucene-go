package store

var _ DataOutput = &ByteArrayDataOutput{}

// ByteArrayDataOutput DataOutput backed by a byte array. WARNING: This class omits most low-level checks, so be sure to test heavily with assertions enabled.
type ByteArrayDataOutput struct {
	*Writer

	bytes []byte
	pos   int
	limit int
}

func NewByteArrayDataOutput(bytes []byte) *ByteArrayDataOutput {
	output := &ByteArrayDataOutput{bytes: bytes, pos: 0, limit: len(bytes)}
	output.Writer = NewWriter(output)
	return output
}

func (r *ByteArrayDataOutput) Write(b []byte) (int, error) {
	copy(r.bytes[r.pos:], b)
	r.pos += len(b)
	return len(b), nil
}

func (r *ByteArrayDataOutput) Reset(bytes []byte) error {
	return r.Reset3(bytes, 0, len(bytes))
}

func (r *ByteArrayDataOutput) Reset3(bytes []byte, offset, size int) error {
	r.bytes = bytes
	r.pos = offset
	r.limit = offset + size
	return nil
}

func (r *ByteArrayDataOutput) GetPosition() int {
	return r.pos
}
