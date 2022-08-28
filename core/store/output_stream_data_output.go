package store

import "io"

var (
	_ DataOutput = &OutputStreamDataOutput{}
	_ io.Closer  = &OutputStreamDataOutput{}
)

type OutputStreamDataOutput struct {
	*DataOutputImp

	os io.WriteCloser
}

func NewOutputStreamDataOutput(os io.WriteCloser) *OutputStreamDataOutput {
	output := &OutputStreamDataOutput{os: os}
	output.DataOutputImp = NewDataOutputImp(output)
	return output
}

func (o *OutputStreamDataOutput) Close() error {
	return o.os.Close()
}

func (o *OutputStreamDataOutput) WriteBytes(b []byte) error {
	_, err := o.os.Write(b)
	return err
}
