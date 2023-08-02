package attribute

type Source struct {
	packed   *packedTokenAttr
	termAttr *bytesAttr
	payload  *bytesAttr
}

func NewSource() *Source {
	return &Source{
		packed:   newPackedTokenAttr(),
		termAttr: newBytesAttr(ClassBytesTerm, ClassTermToBytesRef),
		payload:  newBytesAttr(ClassPayload),
	}
}

func (r *Source) Type() TypeAttr {
	return r.packed
}

func (r *Source) PackedTokenAttribute() PackedTokenAttr {
	return r.packed
}

func (r *Source) BytesTerm() BytesTermAttr {
	return r.termAttr
}

func (r *Source) Payload() PayloadAttr {
	return r.payload
}

func (r *Source) CharTerm() CharTermAttr {
	return r.packed.bytesAttr
}

func (r *Source) Offset() OffsetAttr {
	return r.packed
}

func (r *Source) PositionIncrement() PositionIncrAttr {
	return r.packed
}

func (r *Source) PositionLength() PositionLengthAttr {
	return r.packed
}

func (r *Source) TermFrequency() TermFreqAttr {
	return r.packed
}

func (r *Source) Term2Bytes() Term2BytesAttr {
	return r.termAttr
}

func (r *Source) Reset() error {
	if err := r.packed.Reset(); err != nil {
		return err
	}

	if err := r.termAttr.Reset(); err != nil {
		return err
	}

	return r.payload.Reset()
}
