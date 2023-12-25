package store

import (
	"hash/crc32"
	"io"
)

// ChecksumIndexInput Extension of IndexInput, computing checksum as it goes. Callers can retrieve the checksum via getChecksum().
type ChecksumIndexInput interface {
	IndexInput

	// GetChecksum Returns the current checksum value
	GetChecksum() uint32
}

var _ ChecksumIndexInput = &BufferedChecksumIndexInput{}

// BufferedChecksumIndexInput
// Simple implementation of ChecksumIndexInput that wraps another input and delegates calls.
type BufferedChecksumIndexInput struct {
	*BaseIndexInput

	input  IndexInput
	digest Hash
}

func NewBufferedChecksumIndexInput(in IndexInput) *BufferedChecksumIndexInput {
	input := &BufferedChecksumIndexInput{
		input:  in,
		digest: NewHash(),
	}
	input.BaseIndexInput = NewBaseIndexInput(input)
	return input
}

type Hash interface {
	Write(p []byte)
	Sum() uint32
	Clone() Hash
}

// 自定义
type crc32Hash struct {
	crc uint32
	tab *crc32.Table
}

func NewHash() Hash {
	return &crc32Hash{tab: crc32.IEEETable}
}

func (c *crc32Hash) Write(p []byte) {
	c.crc = crc32.Update(c.crc, c.tab, p)
}

func (c *crc32Hash) Sum() uint32 {
	return c.crc
}

func (c *crc32Hash) Clone() Hash {
	newTab := &crc32.Table{}

	copy(newTab[:], c.tab[:])

	return &crc32Hash{
		crc: c.crc,
		tab: newTab,
	}
}

func (b *BufferedChecksumIndexInput) Slice(sliceDescription string, offset, length int64) (IndexInput, error) {
	//TODO implement me
	panic("implement me")
}

func (b *BufferedChecksumIndexInput) GetFilePointer() int64 {
	return b.input.GetFilePointer()
}

func (b *BufferedChecksumIndexInput) Seek(pos int64, whence int) (int64, error) {
	return b.input.Seek(pos, io.SeekStart)
}

func (b *BufferedChecksumIndexInput) Length() int64 {
	return b.input.Length()
}

//func (b *BufferedChecksumIndexInput) ReadByte() (byte, error) {
//	readByte, err := b.input.ReadByte()
//	if err != nil {
//		return 0, err
//	}
//
//	b.digest.Write([]byte{readByte})
//
//	return readByte, nil
//}

func (b *BufferedChecksumIndexInput) Read(bs []byte) (int, error) {
	n, err := b.input.Read(bs)
	if err != nil {
		return 0, err
	}
	b.digest.Write(bs[:n])
	return n, nil
}

func (b *BufferedChecksumIndexInput) Clone() CloneReader {
	indexInput, ok := b.input.Clone().(IndexInput)
	if !ok {
		return b
	}

	input := &BufferedChecksumIndexInput{
		input:  indexInput,
		digest: b.digest.Clone(),
	}
	input.BaseIndexInput = NewBaseIndexInput(input)
	return input
}

func (b *BufferedChecksumIndexInput) GetChecksum() uint32 {
	return b.digest.Sum()
}

func (b *BufferedChecksumIndexInput) Close() error {
	return b.input.Close()
}
