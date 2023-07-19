package store

import "io"

var (
	_ DataOutput = &OutputStreamDataOutput{}
	_ io.Closer  = &OutputStreamDataOutput{}
)

type OutputStreamDataOutput struct {
	*Writer

	os io.WriteCloser
}

func NewOutputStreamDataOutput(os io.WriteCloser) *OutputStreamDataOutput {
	output := &OutputStreamDataOutput{os: os}
	output.Writer = NewWriter(output)
	return output
}

func (o *OutputStreamDataOutput) Close() error {
	return o.os.Close()
}

func (o *OutputStreamDataOutput) Write(b []byte) (int, error) {
	return o.os.Write(b)
}
