package store

import "io"

var _ DataInput = &InputStreamDataInput{}

// InputStreamDataInput A DataInput wrapping a plain InputStream.
type InputStreamDataInput struct {
	*DataInputImp

	eof bool
	is  io.ReadCloser
}

func NewInputStreamDataInput(is io.ReadCloser) *InputStreamDataInput {
	input := &InputStreamDataInput{is: is}
	input.DataInputImp = NewDataInputImp(input)
	return input
}

func (i *InputStreamDataInput) ReadByte() (byte, error) {
	bs := []byte{0}
	err := i.ReadBytes(bs)
	if err != nil {
		return 0, err
	}
	return bs[0], nil
}

func (i *InputStreamDataInput) ReadBytes(b []byte) error {
	if i.eof {
		return io.EOF
	}

	_, err := i.is.Read(b)
	if err != nil {
		if err == io.EOF {
			i.eof = true
			return nil
		}
		return err
	}
	return err
}

func (i *InputStreamDataInput) Close() error {
	return i.is.Close()
}
