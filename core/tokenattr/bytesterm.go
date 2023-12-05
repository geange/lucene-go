package tokenattr

var (
	_ Attribute               = &BytesTermAttrBase{}
	_ BytesTermAttribute      = &BytesTermAttrBase{}
	_ TermToBytesRefAttribute = &BytesTermAttrBase{}
)

func NewBytesTermAttr() *BytesTermAttrBase {
	return &BytesTermAttrBase{bytes: make([]byte, 0)}
}

// BytesTermAttrBase Implementation class for BytesTermAttribute.
type BytesTermAttrBase struct {
	bytes []byte
}

func (b *BytesTermAttrBase) GetBytes() []byte {
	return b.bytes
}

func (b *BytesTermAttrBase) SetBytes(bytes []byte) error {
	b.bytes = bytes
	return nil
}

func (b *BytesTermAttrBase) Interfaces() []string {
	return []string{
		"BytesTerm",
		"TermToBytesRef",
	}
}

func (b *BytesTermAttrBase) Clear() error {
	b.bytes = nil
	return nil
}

func (b *BytesTermAttrBase) End() error {
	return b.Clear()
}

func (b *BytesTermAttrBase) CopyTo(target Attribute) error {
	impl, ok := target.(*BytesTermAttrBase)
	if ok {
		bytes := make([]byte, len(b.bytes))
		copy(bytes, b.bytes)
		return impl.SetBytes(bytes)
	}
	return nil
}

func (b *BytesTermAttrBase) Clone() Attribute {
	bytes := make([]byte, len(b.bytes))
	copy(bytes, b.bytes)
	return &BytesTermAttrBase{bytes: bytes}
}
