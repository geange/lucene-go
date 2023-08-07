package store

import (
	"hash"
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

// BufferedChecksumIndexInput Simple implementation of ChecksumIndexInput that wraps another input and delegates calls.
type BufferedChecksumIndexInput struct {
	*IndexInputBase

	main   IndexInput
	digest hash.Hash32
}

func NewBufferedChecksumIndexInput(main IndexInput) *BufferedChecksumIndexInput {
	input := &BufferedChecksumIndexInput{
		main:   main,
		digest: crc32.NewIEEE(),
	}
	input.IndexInputBase = NewIndexInputBase(input)
	return input
}

func (b *BufferedChecksumIndexInput) Clone() IndexInput {
	panic("")
}

func (b *BufferedChecksumIndexInput) Slice(sliceDescription string, offset, length int64) (IndexInput, error) {
	//TODO implement me
	panic("implement me")
}

func (b *BufferedChecksumIndexInput) GetFilePointer() int64 {
	return b.main.GetFilePointer()
}

func (b *BufferedChecksumIndexInput) Seek(pos int64, whence int) (int64, error) {
	return b.main.Seek(pos, io.SeekStart)
}

func (b *BufferedChecksumIndexInput) Length() int64 {
	return b.main.Length()
}

func (b *BufferedChecksumIndexInput) ReadByte() (byte, error) {
	readByte, err := b.main.ReadByte()
	if err != nil {
		return 0, err
	}
	if _, err = b.digest.Write([]byte{readByte}); err != nil {
		return 0, err
	}
	return readByte, nil
}

func (b *BufferedChecksumIndexInput) Read(bs []byte) (int, error) {
	n, err := b.main.Read(bs)
	if err != nil {
		return 0, err
	}
	if _, err := b.digest.Write(bs[:n]); err != nil {
		return 0, err
	}
	return len(bs), nil
}

func (b *BufferedChecksumIndexInput) GetChecksum() uint32 {
	return b.digest.Sum32()
}

func (b *BufferedChecksumIndexInput) Close() error {
	return b.main.Close()
}
