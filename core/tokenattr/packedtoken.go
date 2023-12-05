package tokenattr

import (
	"errors"
)

var (
	_ TypeAttribute              = &PackedTokenAttrBase{}
	_ PositionIncrementAttribute = &PackedTokenAttrBase{}
	_ PositionLengthAttribute    = &PackedTokenAttrBase{}
	_ OffsetAttribute            = &PackedTokenAttrBase{}
	_ TermFrequencyAttribute     = &PackedTokenAttrBase{}
	_ Attribute                  = &PackedTokenAttrBase{}
)

type PackedTokenAttribute interface {
	TypeAttribute
	PositionIncrementAttribute
	PositionLengthAttribute
	OffsetAttribute
	TermFrequencyAttribute
	Attribute
}

func NewPackedTokenAttr() *PackedTokenAttrBase {
	return &PackedTokenAttrBase{
		CharTermAttrBase:  NewCharTermAttr(),
		startOffset:       0,
		endOffset:         0,
		_type:             DEFAULT_TYPE,
		positionIncrement: 1,
		positionLength:    1,
		termFrequency:     1,
	}
}

type PackedTokenAttrBase struct {
	*CharTermAttrBase

	startOffset       int
	endOffset         int
	_type             string
	positionIncrement int
	positionLength    int
	termFrequency     int
}

func (p *PackedTokenAttrBase) SetTermFrequency(termFrequency int) error {
	if termFrequency < 1 {
		return errors.New("term frequency must be 1 or greater")
	}
	p.termFrequency = termFrequency
	return nil
}

func (p *PackedTokenAttrBase) GetTermFrequency() int {
	return p.termFrequency
}

func (p *PackedTokenAttrBase) StartOffset() int {
	return p.startOffset
}

func (p *PackedTokenAttrBase) EndOffset() int {
	return p.endOffset
}

func (p *PackedTokenAttrBase) SetOffset(startOffset, endOffset int) error {
	if startOffset < 0 || startOffset > endOffset {
		return errors.New("startOffset must be non-negative, and endOffset must be >= startOffset")
	}
	p.startOffset = startOffset
	p.endOffset = endOffset
	return nil
}

func (p *PackedTokenAttrBase) SetPositionLength(positionLength int) error {
	if positionLength < 1 {
		return errors.New("position length must be 1 or greater")
	}
	p.positionLength = positionLength
	return nil
}

func (p *PackedTokenAttrBase) GetPositionLength() int {
	return p.positionLength
}

func (p *PackedTokenAttrBase) SetPositionIncrement(positionIncrement int) error {
	if positionIncrement < 0 {
		return errors.New("increment must be zero or greater")
	}
	p.positionIncrement = positionIncrement
	return nil
}

func (p *PackedTokenAttrBase) GetPositionIncrement() int {
	return p.positionIncrement
}

func (p *PackedTokenAttrBase) Type() string {
	return p._type
}

func (p *PackedTokenAttrBase) SetType(_type string) {
	p._type = _type
}

func (p *PackedTokenAttrBase) Interfaces() []string {
	values := []string{
		"Type",
		"PositionIncrement",
		"PositionLength",
		"Offset",
		"TermFrequency",
	}
	return append(p.CharTermAttrBase.Interfaces(), values...)
}

func (p *PackedTokenAttrBase) Clear() error {
	p.positionIncrement, p.positionLength = 1, 1
	p.termFrequency = 1
	p.startOffset, p.endOffset = 0, 0
	p._type = "word"
	return p.CharTermAttrBase.Clear()
}

func (p *PackedTokenAttrBase) End() error {
	p.positionIncrement = 0
	return nil
}

func (p *PackedTokenAttrBase) CopyTo(target Attribute) error {
	if impl, ok := target.(*PackedTokenAttrBase); ok {
		impl.startOffset = p.startOffset
		impl.endOffset = p.endOffset
		impl._type = p._type
		impl.positionIncrement = p.positionIncrement
		impl.positionLength = p.positionLength
		impl.termFrequency = p.termFrequency
		return nil
	}
	return errors.New("target is not PackedTokenAttrBase")
}

func (p *PackedTokenAttrBase) Clone() Attribute {
	return &PackedTokenAttrBase{
		startOffset:       p.startOffset,
		endOffset:         p.endOffset,
		_type:             p._type,
		positionIncrement: p.positionIncrement,
		positionLength:    p.positionLength,
		termFrequency:     p.termFrequency,
	}
}
