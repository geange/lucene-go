package core

// PayloadAttributeImpl Default implementation of PayloadAttribute.
type PayloadAttributeImpl struct {
	payload []byte
}

func NewPayloadAttributeImpl() *PayloadAttributeImpl {
	return &PayloadAttributeImpl{payload: make([]byte, 0)}
}

func (p *PayloadAttributeImpl) Interfaces() []string {
	return []string{"Payload"}
}

func (p *PayloadAttributeImpl) Clear() error {
	p.payload = nil
	return nil
}

func (p *PayloadAttributeImpl) End() error {
	return p.Clear()
}

func (p *PayloadAttributeImpl) CopyTo(target AttributeImpl) error {
	attr, ok := target.(*PayloadAttributeImpl)
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

func (p *PayloadAttributeImpl) Clone() AttributeImpl {
	attr := &PayloadAttributeImpl{payload: make([]byte, len(p.payload))}
	copy(attr.payload, p.payload)
	return attr
}

func (p *PayloadAttributeImpl) GetPayload() []byte {
	return p.payload
}

func (p *PayloadAttributeImpl) SetPayload(payload []byte) error {
	p.payload = payload
	return nil
}
