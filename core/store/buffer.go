package store

import (
	"bytes"
	"errors"
)

var (
	_ IndexOutput = &BufferDataOutput{}
)

type BufferDataOutput struct {
	*BaseDataOutput

	buf *bytes.Buffer
}

func (b *BufferDataOutput) Close() error {
	b.buf.Reset()
	return nil
}

func (b *BufferDataOutput) GetName() string {
	return ""
}

func (b *BufferDataOutput) GetFilePointer() int64 {
	return int64(b.buf.Len())
}

func (b *BufferDataOutput) GetChecksum() (uint32, error) {
	return 0, errors.New("todo")
}

func NewBufferDataOutput() *BufferDataOutput {
	buf := new(bytes.Buffer)
	return &BufferDataOutput{
		BaseDataOutput: NewBaseDataOutput(buf),
		buf:            buf,
	}
}

func (b *BufferDataOutput) Write(p []byte) (n int, err error) {
	return b.writer.Write(p)
}

func (b *BufferDataOutput) CopyTo(output DataOutput) error {
	if _, err := output.Write(b.buf.Bytes()); err != nil {
		return err
	}
	return nil
}

func (b *BufferDataOutput) Bytes() []byte {
	return b.buf.Bytes()
}

func (b *BufferDataOutput) Reset() {
	b.buf.Reset()
}

func (b *BufferDataOutput) Size() int {
	return b.buf.Len()
}

var _ IndexInput = &BufferInput{}

type BufferInput struct {
	*BaseDataInput

	buf *bytes.Buffer
}

func (b *BufferInput) Seek(offset int64, whence int) (int64, error) {
	return -1, errors.New("unsupported func")
}

func (b *BufferInput) GetFilePointer() int64 {
	return int64(b.buf.Len())
}

func (b *BufferInput) Slice(sliceDescription string, offset, length int64) (IndexInput, error) {
	return nil, errors.New("unsupported func")
}

func (b *BufferInput) Length() int64 {
	return int64(b.buf.Len())
}

func (b *BufferInput) RandomAccessSlice(offset int64, length int64) (RandomAccessInput, error) {
	return nil, errors.New("unsupported RandomAccessSlice")
}

func NewBufferDataInput(buf *bytes.Buffer) *BufferInput {
	input := &BufferInput{
		buf: buf,
	}
	input.BaseDataInput = NewBaseDataInput(input)
	return input
}

func (b *BufferInput) Read(p []byte) (n int, err error) {
	return b.buf.Read(p)
}

func (b *BufferInput) Clone() CloneReader {
	newBuf := new(bytes.Buffer)
	newBuf.Write(b.buf.Bytes())
	return NewBufferDataInput(newBuf)
}
