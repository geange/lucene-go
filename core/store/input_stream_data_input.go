package store

var _ DataInput = &InputStreamDataInput{}

// InputStreamDataInput A DataInput wrapping a plain InputStream.
type InputStreamDataInput struct {
}

func (i *InputStreamDataInput) Close() error {
	//TODO implement me
	panic("implement me")
}
