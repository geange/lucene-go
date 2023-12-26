package tokenattr

import (
	"bytes"
	"errors"
	"slices"
)

// BytesTermAttr
// This attribute can be used if you have the raw term bytes to be indexed.
// It can be used as replacement for CharTermAttr, if binary terms should be indexed.
type BytesTermAttr interface {
	Attribute

	GetBytes() []byte

	// SetBytes
	// Sets the bytes of the term
	SetBytes(bytes []byte) error

	Reset() error
}

// Term2BytesAttr
// This attribute is requested by TermsHashPerField to index the contents.
// This attribute can be used to customize the final byte[] encoding of terms.
// Consumers of this attribute call getBytesRef() for each term.
type Term2BytesAttr interface {
	Attribute

	GetBytes() []byte
}

// PayloadAttr
// The payload of a Token.
// The payload is stored in the index at each position, and can be used to influence scoring when using
// Payload-based queries.
// NOTE: because the payload will be stored at each position, it's usually best to use the minimum number of
// bytes necessary. Some codec implementations may optimize payload storage when all payloads have the same length.
// See Also: org.apache.lucene.index.PostingsEnum
type PayloadAttr interface {
	Attribute

	// GetPayload Returns this Token's payload.
	// See Also: setPayload(BytesRef)
	GetPayload() []byte

	// SetPayload Sets this Token's payload.
	// See Also: getPayload()
	SetPayload(payload []byte) error

	Reset() error
}

// CharTermAttr The term text of a Token.
type CharTermAttr interface {
	Attribute

	GetBytes() []byte

	// GetString
	// Returns the internal termBuffer character array which you can then directly alter. If the array is
	// too small for your token, use resizeBuffer(int) to increase it. After altering the buffer be sure to call
	// setLength to record the number of valid characters that were placed into the termBuffer.
	GetString() string

	// AppendString
	// Appends the specified String to this character sequence.
	// The characters of the String argument are appended, in order, increasing the length of this sequence by the
	// length of the argument. If argument is null, then the four characters "null" are appended.
	AppendString(s string) error

	AppendRune(r rune) error

	// Reset
	// Sets the length of the termBuffer to zero. Use this method before appending contents
	// using the Appendable interface.
	Reset() error
}

var (
	_ Attribute = &bytesAttr{}
)

type bytesAttr struct {
	classes []string
	buf     *bytes.Buffer
}

func (b *bytesAttr) Interfaces() []string {
	return b.classes
}

func newBytesAttr(classes ...string) *bytesAttr {
	return &bytesAttr{
		classes: classes,
		buf:     new(bytes.Buffer),
	}
}

func newBytesTermAttr() *bytesAttr {
	return newBytesAttr(ClassBytesTerm, ClassTermToBytesRef)
}

func newCharTermAttr() *bytesAttr {
	return newBytesAttr(ClassCharTerm, ClassTermToBytesRef)
}

func newPayloadAttr() *bytesAttr {
	return newBytesAttr(ClassPayload)
}

func (b *bytesAttr) SetBytes(p []byte) error {
	b.buf.Reset()
	b.buf.Write(p)
	return nil
}

func (b *bytesAttr) SetString(s string) error {
	b.buf.Reset()
	b.buf.WriteString(s)
	return nil
}

func (b *bytesAttr) AppendRune(r rune) error {
	b.buf.WriteRune(r)
	return nil
}

func (b *bytesAttr) AppendString(s string) error {
	b.buf.WriteString(s)
	return nil
}

func (b *bytesAttr) GetString() string {
	return b.buf.String()
}

func (b *bytesAttr) GetBytes() []byte {
	return b.buf.Bytes()
}

func (b *bytesAttr) Reset() error {
	b.buf.Reset()
	return nil
}

func (b *bytesAttr) GetPayload() []byte {
	return b.GetBytes()
}

func (b *bytesAttr) SetPayload(payload []byte) error {
	return b.SetBytes(payload)
}

func (b *bytesAttr) CopyTo(target Attribute) error {
	attr, ok := target.(*bytesAttr)
	if ok {
		attr.buf.Reset()
		attr.classes = slices.Clone(b.classes)
		return attr.SetBytes(b.GetBytes())
	}
	return errors.New("target is not *bytesAttr")
}

func (b *bytesAttr) Clone() Attribute {
	attr := &bytesAttr{
		classes: slices.Clone(b.classes),
		buf:     new(bytes.Buffer),
	}
	_ = attr.SetBytes(b.GetBytes())
	return attr
}
