package store

import "io"

var _ IndexOutput = &OutputStreamIndexOutput{}

// OutputStreamIndexOutput Implementation class for buffered IndexOutput that writes to an OutputStream.
type OutputStreamIndexOutput struct {
	*DataOutputImp
	out io.WriteCloser

	bytesWritten   int64
	flushedOnClose bool
}

func NewOutputStreamIndexOutput(out io.WriteCloser) *OutputStreamIndexOutput {
	output := &OutputStreamIndexOutput{
		out: out,
	}
	output.DataOutputImp = NewDataOutputImp(output)
	return output
}

func (o *OutputStreamIndexOutput) WriteByte(b byte) error {
	o.bytesWritten++
	_, err := o.out.Write([]byte{b})
	return err
}

func (o *OutputStreamIndexOutput) WriteBytes(b []byte) error {
	o.bytesWritten += int64(len(b))
	_, err := o.out.Write(b)
	return err
}

func (o *OutputStreamIndexOutput) Close() error {
	return o.out.Close()
}

func (o *OutputStreamIndexOutput) GetFilePointer() int64 {
	return o.bytesWritten
}
