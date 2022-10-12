package store

import (
	"hash"
	"hash/crc32"
)

var _ ChecksumIndexInput = &BufferedChecksumIndexInput{}

// BufferedChecksumIndexInput Simple implementation of ChecksumIndexInput that wraps another input and delegates calls.
type BufferedChecksumIndexInput struct {
	*IndexInputImp

	main   IndexInput
	digest hash.Hash32
}

func NewBufferedChecksumIndexInput(main IndexInput) *BufferedChecksumIndexInput {
	input := &BufferedChecksumIndexInput{
		main:   main,
		digest: crc32.NewIEEE(),
	}
	input.IndexInputImp = NewIndexInputImp(input)
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

func (b *BufferedChecksumIndexInput) Seek(pos int64) error {
	return b.main.Seek(pos)
}

func (b *BufferedChecksumIndexInput) Length() int64 {
	return b.Length()
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

func (b *BufferedChecksumIndexInput) ReadBytes(bs []byte) error {
	if err := b.main.ReadBytes(bs); err != nil {
		return err
	}
	if _, err := b.digest.Write(bs); err != nil {
		return err
	}
	return nil
}

func (b *BufferedChecksumIndexInput) GetChecksum() uint32 {
	return b.digest.Sum32()
}
