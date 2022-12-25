package store

import "io"

var (
	_ DataOutput = &OutputStreamDataOutput{}
	_ io.Closer  = &OutputStreamDataOutput{}
)

type OutputStreamDataOutput struct {
	*DataOutputDefault

	os io.WriteCloser
}

func NewOutputStreamDataOutput(os io.WriteCloser) *OutputStreamDataOutput {
	output := &OutputStreamDataOutput{os: os}
	output.DataOutputDefault = NewDataOutputDefault(&DataOutputDefaultConfig{
		WriteByte:  nil,
		WriteBytes: output.Write,
	})
	return output
}

func (o *OutputStreamDataOutput) Close() error {
	return o.os.Close()
}

func (o *OutputStreamDataOutput) Write(b []byte) (int, error) {
	return o.os.Write(b)
}
