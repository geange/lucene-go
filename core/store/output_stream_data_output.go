package store

import "io"

var (
	_ DataOutput = &OutputStreamDataOutput{}
	_ io.Closer  = &OutputStreamDataOutput{}
)

type OutputStreamDataOutput struct {
}

func (o *OutputStreamDataOutput) Close() error {
	//TODO implement me
	panic("implement me")
}
