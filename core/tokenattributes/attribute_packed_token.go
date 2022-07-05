package tokenattributes

import (
	"errors"
)

var (
	_ TypeAttribute              = &PackedTokenAttributeIMP{}
	_ PositionIncrementAttribute = &PackedTokenAttributeIMP{}
	_ PositionLengthAttribute    = &PackedTokenAttributeIMP{}
	_ OffsetAttribute            = &PackedTokenAttributeIMP{}
	_ TermFrequencyAttribute     = &PackedTokenAttributeIMP{}
	_ AttributeImpl              = &PackedTokenAttributeIMP{}
)

type PackedTokenAttribute interface {
	TypeAttribute
	PositionIncrementAttribute
	PositionLengthAttribute
	OffsetAttribute
	TermFrequencyAttribute
	AttributeImpl
}

func NewPackedTokenAttributeIMP() *PackedTokenAttributeIMP {
	return &PackedTokenAttributeIMP{
		CharTermAttributeIMP: NewCharTermAttributeImpl(),
		startOffset:          0,
		endOffset:            0,
		_type:                DEFAULT_TYPE,
		positionIncrement:    1,
		positionLength:       1,
		termFrequency:        1,
	}
}

type PackedTokenAttributeIMP struct {
	*CharTermAttributeIMP

	startOffset       int
	endOffset         int
	_type             string
	positionIncrement int
	positionLength    int
	termFrequency     int
}

func (p *PackedTokenAttributeIMP) SetTermFrequency(termFrequency int) error {
	if termFrequency < 1 {
		return errors.New("term frequency must be 1 or greater")
	}
	p.termFrequency = termFrequency
	return nil
}

func (p *PackedTokenAttributeIMP) GetTermFrequency() int {
	return p.termFrequency
}

func (p *PackedTokenAttributeIMP) StartOffset() int {
	return p.startOffset
}

func (p *PackedTokenAttributeIMP) EndOffset() int {
	return p.endOffset
}

func (p *PackedTokenAttributeIMP) SetOffset(startOffset, endOffset int) error {
	if startOffset < 0 || startOffset > endOffset {
		return errors.New("startOffset must be non-negative, and endOffset must be >= startOffset")
	}
	p.startOffset = startOffset
	p.endOffset = endOffset
	return nil
}

func (p *PackedTokenAttributeIMP) SetPositionLength(positionLength int) error {
	if positionLength < 1 {
		return errors.New("position length must be 1 or greater")
	}
	p.positionLength = positionLength
	return nil
}

func (p *PackedTokenAttributeIMP) GetPositionLength() int {
	return p.positionLength
}

func (p *PackedTokenAttributeIMP) SetPositionIncrement(positionIncrement int) error {
	if positionIncrement < 0 {
		return errors.New("increment must be zero or greater")
	}
	p.positionIncrement = positionIncrement
	return nil
}

func (p *PackedTokenAttributeIMP) GetPositionIncrement() int {
	return p.positionIncrement
}

func (p *PackedTokenAttributeIMP) Type() string {
	return p._type
}

func (p *PackedTokenAttributeIMP) SetType(_type string) {
	p._type = _type
}

func (p *PackedTokenAttributeIMP) Interfaces() []string {
	values := []string{
		"Type",
		"PositionIncrement",
		"PositionLength",
		"Offset",
		"TermFrequency",
	}
	return append(p.CharTermAttributeIMP.Interfaces(), values...)
}

func (p *PackedTokenAttributeIMP) Clear() error {
	p.positionIncrement, p.positionLength = 1, 1
	p.termFrequency = 1
	p.startOffset, p.endOffset = 0, 0
	p._type = "word"
	return p.CharTermAttributeIMP.Clear()
}

func (p *PackedTokenAttributeIMP) End() error {
	p.positionIncrement = 0
	return nil
}

func (p *PackedTokenAttributeIMP) CopyTo(target AttributeImpl) error {
	if impl, ok := target.(*PackedTokenAttributeIMP); ok {
		impl.startOffset = p.startOffset
		impl.endOffset = p.endOffset
		impl._type = p._type
		impl.positionIncrement = p.positionIncrement
		impl.positionLength = p.positionLength
		impl.termFrequency = p.termFrequency
		return nil
	}
	return errors.New("target is not PackedTokenAttributeIMP")
}

func (p *PackedTokenAttributeIMP) Clone() AttributeImpl {
	return &PackedTokenAttributeIMP{
		startOffset:       p.startOffset,
		endOffset:         p.endOffset,
		_type:             p._type,
		positionIncrement: p.positionIncrement,
		positionLength:    p.positionLength,
		termFrequency:     p.termFrequency,
	}
}
