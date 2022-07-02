package tokenattributes

import (
	"errors"
	"github.com/geange/lucene-go/core/util"
)

var (
	_ TypeAttribute              = &PackedTokenAttributeImpl{}
	_ PositionIncrementAttribute = &PackedTokenAttributeImpl{}
	_ PositionLengthAttribute    = &PackedTokenAttributeImpl{}
	_ OffsetAttribute            = &PackedTokenAttributeImpl{}
	_ TermFrequencyAttribute     = &PackedTokenAttributeImpl{}
	_ util.AttributeImpl         = &PackedTokenAttributeImpl{}
)

func NewPackedTokenAttributeImpl() *PackedTokenAttributeImpl {
	return &PackedTokenAttributeImpl{
		CharTermAttributeImpl: NewCharTermAttributeImpl(),
		startOffset:           0,
		endOffset:             0,
		_type:                 DEFAULT_TYPE,
		positionIncrement:     1,
		positionLength:        1,
		termFrequency:         1,
	}
}

type PackedTokenAttributeImpl struct {
	*CharTermAttributeImpl

	startOffset       int
	endOffset         int
	_type             string
	positionIncrement int
	positionLength    int
	termFrequency     int
}

func (p *PackedTokenAttributeImpl) SetTermFrequency(termFrequency int) error {
	if termFrequency < 1 {
		return errors.New("term frequency must be 1 or greater")
	}
	p.termFrequency = termFrequency
	return nil
}

func (p *PackedTokenAttributeImpl) GetTermFrequency() int {
	return p.termFrequency
}

func (p *PackedTokenAttributeImpl) StartOffset() int {
	return p.startOffset
}

func (p *PackedTokenAttributeImpl) EndOffset() int {
	return p.endOffset
}

func (p *PackedTokenAttributeImpl) SetOffset(startOffset, endOffset int) error {
	if startOffset < 0 || startOffset > endOffset {
		return errors.New("startOffset must be non-negative, and endOffset must be >= startOffset")
	}
	p.startOffset = startOffset
	p.endOffset = endOffset
	return nil
}

func (p *PackedTokenAttributeImpl) SetPositionLength(positionLength int) error {
	if positionLength < 1 {
		return errors.New("position length must be 1 or greater")
	}
	p.positionLength = positionLength
	return nil
}

func (p *PackedTokenAttributeImpl) GetPositionLength() int {
	return p.positionLength
}

func (p *PackedTokenAttributeImpl) SetPositionIncrement(positionIncrement int) error {
	if positionIncrement < 0 {
		return errors.New("increment must be zero or greater")
	}
	p.positionIncrement = positionIncrement
	return nil
}

func (p *PackedTokenAttributeImpl) GetPositionIncrement() int {
	return p.positionIncrement
}

func (p *PackedTokenAttributeImpl) Type() string {
	return p._type
}

func (p *PackedTokenAttributeImpl) SetType(_type string) {
	p._type = _type
}

func (p *PackedTokenAttributeImpl) Interfaces() []string {
	values := []string{
		"Type",
		"PositionIncrement",
		"PositionLength",
		"Offset",
		"TermFrequency",
	}
	return append(p.CharTermAttributeImpl.Interfaces(), values...)
}

func (p *PackedTokenAttributeImpl) Clear() error {
	p.positionIncrement, p.positionLength = 1, 1
	p.termFrequency = 1
	p.startOffset, p.endOffset = 0, 0
	p._type = "word"
	return nil
}

func (p *PackedTokenAttributeImpl) End() error {
	p.positionIncrement = 0
	return nil
}

func (p *PackedTokenAttributeImpl) CopyTo(target util.AttributeImpl) error {
	if impl, ok := target.(*PackedTokenAttributeImpl); ok {
		impl.startOffset = p.startOffset
		impl.endOffset = p.endOffset
		impl._type = p._type
		impl.positionIncrement = p.positionIncrement
		impl.positionLength = p.positionLength
		impl.termFrequency = p.termFrequency
		return nil
	}
	return errors.New("target is not PackedTokenAttributeImpl")
}

func (p *PackedTokenAttributeImpl) Clone() util.AttributeImpl {
	return &PackedTokenAttributeImpl{
		startOffset:       p.startOffset,
		endOffset:         p.endOffset,
		_type:             p._type,
		positionIncrement: p.positionIncrement,
		positionLength:    p.positionLength,
		termFrequency:     p.termFrequency,
	}
}
