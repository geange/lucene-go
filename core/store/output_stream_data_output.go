package store

import "io"

var (
	//_ DataOutput = &OutputStreamDataOutput{}
	_ io.Closer = &OutputStreamDataOutput{}
)

type OutputStreamDataOutput struct {
}

func (o *OutputStreamDataOutput) WriteByte(b byte) error {
	//TODO implement me
	panic("implement me")
}

func (o *OutputStreamDataOutput) WriteBytes(b []byte) error {
	//TODO implement me
	panic("implement me")
}

func (o *OutputStreamDataOutput) CopyBytes(input DataInput, numBytes int) error {
	//TODO implement me
	panic("implement me")
}

func (o *OutputStreamDataOutput) Close() error {
	//TODO implement me
	panic("implement me")
}
