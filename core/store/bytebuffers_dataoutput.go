package store

import "bytes"

var _ DataOutput = &ByteBuffersDataOutput{}

type ByteBuffersDataOutput struct {
	*Writer
	buf *bytes.Buffer
}

func NewByteBuffersDataOutput() *ByteBuffersDataOutput {
	buf := new(bytes.Buffer)
	return &ByteBuffersDataOutput{
		Writer: NewWriter(buf),
		buf:    buf,
	}
}

func (b *ByteBuffersDataOutput) Write(p []byte) (n int, err error) {
	return b.writer.Write(p)
}

func (b *ByteBuffersDataOutput) CopyTo(output DataOutput) error {
	_, err := output.Write(b.buf.Bytes())
	return err
}

func (b *ByteBuffersDataOutput) Bytes() []byte {
	return b.buf.Bytes()
}

func (b *ByteBuffersDataOutput) Reset() {
	b.buf.Reset()
}
