package tokenattr

type AttributeSource struct {
	packed  *PackedTokenAttrBase
	bytes   *BytesTermAttrBase
	payload *PayloadAttrBase
}

func NewAttributeSource() *AttributeSource {
	return &AttributeSource{
		packed:  NewPackedTokenAttr(),
		bytes:   NewBytesTermAttrBase(),
		payload: NewPayloadAttrBase(),
	}
}

func (r *AttributeSource) Type() TypeAttribute {
	return r.packed
}

func (r *AttributeSource) PackedTokenAttribute() PackedTokenAttribute {
	return r.packed
}

func (r *AttributeSource) BytesTerm() BytesTermAttribute {
	return r.bytes
}

func (r *AttributeSource) Payload() PayloadAttribute {
	return r.payload
}

func (r *AttributeSource) CharTerm() CharTermAttribute {
	return r.packed
}

func (r *AttributeSource) Offset() OffsetAttribute {
	return r.packed
}

func (r *AttributeSource) PositionIncrement() PositionIncrementAttribute {
	return r.packed
}

func (r *AttributeSource) PositionLength() PositionLengthAttribute {
	return r.packed
}

func (r *AttributeSource) TermFrequency() TermFrequencyAttribute {
	return r.packed
}

func (r *AttributeSource) TermToBytesRef() TermToBytesRefAttribute {
	return r.bytes
}

func (r *AttributeSource) Clear() error {
	err := r.packed.Clear()
	if err != nil {
		return err
	}
	err = r.bytes.Clear()
	if err != nil {
		return err
	}
	return r.payload.Clear()
}
