package store

import "io"

var _ IndexOutput = &OutputStreamIndexOutput{}

type Writer interface {
	io.Writer
	io.Closer
}

// OutputStreamIndexOutput Implementation class for buffered IndexOutput that writes to an OutputStream.
type OutputStreamIndexOutput struct {
	*DataOutputImp
	out Writer
}

func (o *OutputStreamIndexOutput) GetFilePointer() int64 {
	//TODO implement me
	panic("implement me")
}

func NewOutputStreamIndexOutput(out Writer) *OutputStreamIndexOutput {
	output := &OutputStreamIndexOutput{
		out: out,
	}
	output.DataOutputImp = NewDataOutputImp(output)
	return output
}

func (o *OutputStreamIndexOutput) WriteByte(b byte) error {
	_, err := o.out.Write([]byte{b})
	return err
}

func (o *OutputStreamIndexOutput) WriteBytes(b []byte) error {
	_, err := o.out.Write(b)
	return err
}

func (o *OutputStreamIndexOutput) Close() error {
	return o.out.Close()
}
