package blocktree

import "github.com/geange/lucene-go/core/interface/index"

var _ index.Term = &FieldReader{}

type FieldReader struct {
}

func (f *FieldReader) Field() string {
	//TODO implement me
	panic("implement me")
}

func (f *FieldReader) Text() string {
	//TODO implement me
	panic("implement me")
}

func (f *FieldReader) Bytes() []byte {
	//TODO implement me
	panic("implement me")
}
