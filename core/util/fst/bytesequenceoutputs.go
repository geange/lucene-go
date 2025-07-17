package fst

var _ Output = &ByteSequenceOutput{}

type ByteSequenceOutput []byte

func (b ByteSequenceOutput) Common(v Output) (Output, error) {
	//TODO implement me
	panic("implement me")
}

func (b ByteSequenceOutput) Sub(v Output) (Output, error) {
	//TODO implement me
	panic("implement me")
}

func (b ByteSequenceOutput) Add(v Output) (Output, error) {
	//TODO implement me
	panic("implement me")
}

func (b ByteSequenceOutput) Merge(v Output) (Output, error) {
	//TODO implement me
	panic("implement me")
}

func (b ByteSequenceOutput) IsNoOutput() bool {
	//TODO implement me
	panic("implement me")
}

func (b ByteSequenceOutput) Equal(v Output) bool {
	//TODO implement me
	panic("implement me")
}

func (b ByteSequenceOutput) Hash() int64 {
	//TODO implement me
	panic("implement me")
}
