package tokenattributes

type AttributeSourceV1 struct {
	packed  *PackedTokenAttributeIMP
	bytes   *BytesTermAttributeImpl
	payload *PayloadAttributeImpl
}

func NewAttributeSourceV1() *AttributeSourceV1 {
	return &AttributeSourceV1{
		packed:  NewPackedTokenAttributeIMP(),
		bytes:   NewBytesTermAttributeImpl(),
		payload: NewPayloadAttributeImpl(),
	}
}

func (r *AttributeSourceV1) PackedTokenAttribute() PackedTokenAttribute {
	return r.packed
}

func (r *AttributeSourceV1) BytesTermAttribute() BytesTermAttribute {
	return r.bytes
}

func (r *AttributeSourceV1) PayloadAttribute() PayloadAttribute {
	return r.payload
}

func (r *AttributeSourceV1) Clear() error {
	return r.packed.Clear()
}
