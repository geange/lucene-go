package tokenattr

// PayloadAttrBase Default implementation of PayloadAttribute.
type PayloadAttrBase struct {
	payload []byte
}

func NewPayloadAttr() *PayloadAttrBase {
	return &PayloadAttrBase{payload: make([]byte, 0)}
}

func (p *PayloadAttrBase) Interfaces() []string {
	return []string{"Payload"}
}

func (p *PayloadAttrBase) Clear() error {
	p.payload = nil
	return nil
}

func (p *PayloadAttrBase) End() error {
	return p.Clear()
}

func (p *PayloadAttrBase) CopyTo(target Attribute) error {
	attr, ok := target.(*PayloadAttrBase)
	if ok {
		if len(p.payload) > len(attr.payload) {
			attr.payload = make([]byte, len(p.payload))
		} else {
			attr.payload = attr.payload[:len(p.payload)]
		}
		copy(attr.payload, p.payload)
	}
	return nil
}

func (p *PayloadAttrBase) Clone() Attribute {
	attr := &PayloadAttrBase{payload: make([]byte, len(p.payload))}
	copy(attr.payload, p.payload)
	return attr
}

func (p *PayloadAttrBase) GetPayload() []byte {
	return p.payload
}

func (p *PayloadAttrBase) SetPayload(payload []byte) error {
	p.payload = payload
	return nil
}
