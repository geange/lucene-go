package tokenattr

type AttributeSource struct {
	packed   *packedTokenAttr
	termAttr *bytesAttr
	payload  *bytesAttr
}

func NewAttributeSource() *AttributeSource {
	return &AttributeSource{
		packed:   newPackedTokenAttr(),
		termAttr: newBytesAttr(ClassBytesTerm, ClassTermToBytesRef),
		payload:  newBytesAttr(ClassPayload),
	}
}

func (r *AttributeSource) Type() TypeAttr {
	return r.packed
}

func (r *AttributeSource) PackedTokenAttribute() PackedTokenAttr {
	return r.packed
}

func (r *AttributeSource) BytesTerm() BytesTermAttr {
	return r.termAttr
}

func (r *AttributeSource) Payload() PayloadAttr {
	return r.payload
}

func (r *AttributeSource) CharTerm() CharTermAttr {
	return r.packed.bytesAttr
}

func (r *AttributeSource) Offset() OffsetAttr {
	return r.packed
}

func (r *AttributeSource) PositionIncrement() PositionIncrAttr {
	return r.packed
}

func (r *AttributeSource) PositionLength() PositionLengthAttr {
	return r.packed
}

func (r *AttributeSource) TermFrequency() TermFreqAttr {
	return r.packed
}

func (r *AttributeSource) Term2Bytes() Term2BytesAttr {
	return r.termAttr
}

func (r *AttributeSource) Reset() error {
	if err := r.packed.Reset(); err != nil {
		return err
	}

	if err := r.termAttr.Reset(); err != nil {
		return err
	}

	return r.payload.Reset()
}
