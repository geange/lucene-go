package store

var _ IndexOutput = &RAMOutputStream{}

// RAMOutputStream A memory-resident IndexOutput implementation.
// Deprecated This class uses inefficient synchronization and is discouraged in favor of MMapDirectory.
// It will be removed in future versions of Lucene.
type RAMOutputStream struct {
}

func (r *RAMOutputStream) Close() error {
	//TODO implement me
	panic("implement me")
}

func (r *RAMOutputStream) WriteByte(b byte) error {
	//TODO implement me
	panic("implement me")
}

func (r *RAMOutputStream) WriteBytes(b []byte) error {
	//TODO implement me
	panic("implement me")
}

func (r *RAMOutputStream) CopyBytes(input DataInput, numBytes int) error {
	//TODO implement me
	panic("implement me")
}

func (r *RAMOutputStream) GetFilePointer() int64 {
	//TODO implement me
	panic("implement me")
}
