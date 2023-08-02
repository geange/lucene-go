package store

import (
	"errors"
	"io"
)

var _ IndexInput = &InputStream{}

// InputStream A DataInput wrapping a plain InputStream.
type InputStream struct {
	*BaseDataInput

	eof bool
	is  io.Reader
}

func NewInputStream(is io.Reader) *InputStream {
	input := &InputStream{is: is}
	input.BaseDataInput = NewBaseDataInput(input)
	return input
}

func (i *InputStream) Seek(offset int64, whence int) (int64, error) {
	return 0, errors.New("unsupported operate")
}

func (i *InputStream) GetFilePointer() int64 {
	return -1
}

func (i *InputStream) Slice(sliceDescription string, offset, length int64) (IndexInput, error) {
	return nil, errors.New("unsupported operate")
}

func (i *InputStream) Length() int64 {
	return 0
}

func (i *InputStream) RandomAccessSlice(offset int64, length int64) (RandomAccessInput, error) {
	return nil, errors.New("unsupported operate")
}

func (i *InputStream) Clone() CloneReader {
	return i
}

func (i *InputStream) ReadByte() (byte, error) {
	bs := [1]byte{}
	if _, err := i.Read(bs[:]); err != nil {
		return 0, err
	}
	return bs[0], nil
}

func (i *InputStream) Read(b []byte) (int, error) {
	return i.is.Read(b)
}

func (i *InputStream) Close() error {
	if closer, ok := i.is.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
