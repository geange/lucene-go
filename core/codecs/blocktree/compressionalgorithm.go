package blocktree

import (
	"context"
	"errors"

	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/compress"
)

var (
	NO_COMPRESSION  = &noCompression{}
	LOWERCASE_ASCII = &lowercaseAscii{}
	LZ4             = &lz4Algorithm{}
)

type CompressionAlgorithm interface {
	Code() int
	Read(ctx context.Context, in store.DataInput, bs []byte) error
}

type noCompression struct {
}

func (*noCompression) Code() int {
	return 0
}

func (n *noCompression) Read(ctx context.Context, in store.DataInput, out []byte) error {
	_, err := in.Read(out)
	return err
}

type lowercaseAscii struct {
}

func (*lowercaseAscii) Code() int {
	return 1
}

func (*lowercaseAscii) Read(ctx context.Context, in store.DataInput, out []byte) error {
	return compress.LowercaseAsciiCompression.Decompress(ctx, in, out)
}

type lz4Algorithm struct {
}

func (*lz4Algorithm) Code() int {
	return 2
}

func (*lz4Algorithm) Read(ctx context.Context, in store.DataInput, out []byte) error {
	return compress.LZ4Compression.Decompress(in, out)
}

func ByCode(code int) (CompressionAlgorithm, error) {
	switch code {
	case 0:
		return &noCompression{}, nil
	case 1:
		return &lowercaseAscii{}, nil
	case 2:
		return &lz4Algorithm{}, nil
	default:
		return nil, errors.New("unsupported compression")
	}
}
