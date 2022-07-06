package tokenattributes

import (
	"errors"
)

var (
	_ TypeAttribute              = &PackedTokenAttributeImp{}
	_ PositionIncrementAttribute = &PackedTokenAttributeImp{}
	_ PositionLengthAttribute    = &PackedTokenAttributeImp{}
	_ OffsetAttribute            = &PackedTokenAttributeImp{}
	_ TermFrequencyAttribute     = &PackedTokenAttributeImp{}
	_ AttributeImpl              = &PackedTokenAttributeImp{}
)

type PackedTokenAttribute interface {
	TypeAttribute
	PositionIncrementAttribute
	PositionLengthAttribute
	OffsetAttribute
	TermFrequencyAttribute
	AttributeImpl
}

func NewPackedTokenAttributeImp() *PackedTokenAttributeImp {
	return &PackedTokenAttributeImp{
		CharTermAttributeImp: NewCharTermAttributeImpl(),
		startOffset:          0,
		endOffset:            0,
		_type:                DEFAULT_TYPE,
		positionIncrement:    1,
		positionLength:       1,
		termFrequency:        1,
	}
}

type PackedTokenAttributeImp struct {
	*CharTermAttributeImp

	startOffset       int
	endOffset         int
	_type             string
	positionIncrement int
	positionLength    int
	termFrequency     int
}

func (p *PackedTokenAttributeImp) SetTermFrequency(termFrequency int) error {
	if termFrequency < 1 {
		return errors.New("term frequency must be 1 or greater")
	}
	p.termFrequency = termFrequency
	return nil
}

func (p *PackedTokenAttributeImp) GetTermFrequency() int {
	return p.termFrequency
}

func (p *PackedTokenAttributeImp) StartOffset() int {
	return p.startOffset
}

func (p *PackedTokenAttributeImp) EndOffset() int {
	return p.endOffset
}

func (p *PackedTokenAttributeImp) SetOffset(startOffset, endOffset int) error {
	if startOffset < 0 || startOffset > endOffset {
		return errors.New("startOffset must be non-negative, and endOffset must be >= startOffset")
	}
	p.startOffset = startOffset
	p.endOffset = endOffset
	return nil
}

func (p *PackedTokenAttributeImp) SetPositionLength(positionLength int) error {
	if positionLength < 1 {
		return errors.New("position length must be 1 or greater")
	}
	p.positionLength = positionLength
	return nil
}

func (p *PackedTokenAttributeImp) GetPositionLength() int {
	return p.positionLength
}

func (p *PackedTokenAttributeImp) SetPositionIncrement(positionIncrement int) error {
	if positionIncrement < 0 {
		return errors.New("increment must be zero or greater")
	}
	p.positionIncrement = positionIncrement
	return nil
}

func (p *PackedTokenAttributeImp) GetPositionIncrement() int {
	return p.positionIncrement
}

func (p *PackedTokenAttributeImp) Type() string {
	return p._type
}

func (p *PackedTokenAttributeImp) SetType(_type string) {
	p._type = _type
}

func (p *PackedTokenAttributeImp) Interfaces() []string {
	values := []string{
		"Type",
		"PositionIncrement",
		"PositionLength",
		"Offset",
		"TermFrequency",
	}
	return append(p.CharTermAttributeImp.Interfaces(), values...)
}

func (p *PackedTokenAttributeImp) Clear() error {
	p.positionIncrement, p.positionLength = 1, 1
	p.termFrequency = 1
	p.startOffset, p.endOffset = 0, 0
	p._type = "word"
	return p.CharTermAttributeImp.Clear()
}

func (p *PackedTokenAttributeImp) End() error {
	p.positionIncrement = 0
	return nil
}

func (p *PackedTokenAttributeImp) CopyTo(target AttributeImpl) error {
	if impl, ok := target.(*PackedTokenAttributeImp); ok {
		impl.startOffset = p.startOffset
		impl.endOffset = p.endOffset
		impl._type = p._type
		impl.positionIncrement = p.positionIncrement
		impl.positionLength = p.positionLength
		impl.termFrequency = p.termFrequency
		return nil
	}
	return errors.New("target is not PackedTokenAttributeImp")
}

func (p *PackedTokenAttributeImp) Clone() AttributeImpl {
	return &PackedTokenAttributeImp{
		startOffset:       p.startOffset,
		endOffset:         p.endOffset,
		_type:             p._type,
		positionIncrement: p.positionIncrement,
		positionLength:    p.positionLength,
		termFrequency:     p.termFrequency,
	}
}
