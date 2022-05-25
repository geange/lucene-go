package core

var (
	_ AttributeImpl           = &BytesTermAttributeImpl{}
	_ BytesTermAttribute      = &BytesTermAttributeImpl{}
	_ TermToBytesRefAttribute = &BytesTermAttributeImpl{}
)

func NewBytesTermAttributeImpl() *BytesTermAttributeImpl {
	return &BytesTermAttributeImpl{bytes: make([]byte, 0)}
}

// BytesTermAttributeImpl Implementation class for BytesTermAttribute.
type BytesTermAttributeImpl struct {
	bytes []byte
}

func (b *BytesTermAttributeImpl) GetBytesRef() []byte {
	return b.bytes
}

func (b *BytesTermAttributeImpl) SetBytesRef(bytes []byte) error {
	b.bytes = bytes
	return nil
}

func (b *BytesTermAttributeImpl) Interfaces() []string {
	return []string{
		"BytesTerm",
		"TermToBytesRef",
	}
}

func (b *BytesTermAttributeImpl) Clear() error {
	b.bytes = nil
	return nil
}

func (b *BytesTermAttributeImpl) End() error {
	return b.Clear()
}

func (b *BytesTermAttributeImpl) CopyTo(target AttributeImpl) error {
	impl, ok := target.(*BytesTermAttributeImpl)
	if ok {
		bytes := make([]byte, len(b.bytes))
		copy(bytes, b.bytes)
		return impl.SetBytesRef(bytes)
	}
	return nil
}

func (b *BytesTermAttributeImpl) Clone() AttributeImpl {
	bytes := make([]byte, len(b.bytes))
	copy(bytes, b.bytes)
	return &BytesTermAttributeImpl{bytes: bytes}
}
