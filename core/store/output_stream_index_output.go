package store

var _ IndexOutput = &OutputStreamIndexOutput{}

// OutputStreamIndexOutput Implementation class for buffered IndexOutput that writes to an OutputStream.
type OutputStreamIndexOutput struct {
}

func (o *OutputStreamIndexOutput) Close() error {
	//TODO implement me
	panic("implement me")
}
