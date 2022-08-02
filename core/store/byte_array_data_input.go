package store

var _ DataInput = &ByteArrayDataInput{}

// ByteArrayDataInput DataInput backed by a byte array. WARNING: This class omits all low-level checks.
type ByteArrayDataInput struct {
}

func (b *ByteArrayDataInput) Close() error {
	//TODO implement me
	panic("implement me")
}
