package store

import "io"

var _ DataInput = &InputStreamDataInput{}

// InputStreamDataInput A DataInput wrapping a plain InputStream.
type InputStreamDataInput struct {
	*DataInputDefault

	eof bool
	is  io.ReadCloser
}

func NewInputStreamDataInput(is io.ReadCloser) *InputStreamDataInput {
	input := &InputStreamDataInput{is: is}
	input.DataInputDefault = NewDataInputDefault(&DataInputDefaultConfig{
		ReadByte: input.ReadByte,
		Read:     input.Read,
	})
	return input
}

func (i *InputStreamDataInput) ReadByte() (byte, error) {
	bs := []byte{0}
	_, err := i.Read(bs)
	if err != nil {
		return 0, err
	}
	return bs[0], nil
}

func (i *InputStreamDataInput) Read(b []byte) (int, error) {
	if i.eof {
		return 0, io.EOF
	}

	return i.is.Read(b)
}

func (i *InputStreamDataInput) Close() error {
	return i.is.Close()
}
