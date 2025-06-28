package blocktree

import (
	"context"
	"errors"

	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/compress"
)

type CompressionAlgorithm interface {
	Read(ctx context.Context, in store.DataInput, bs []byte) error
}

type noCompression struct {
}

func (n *noCompression) Read(ctx context.Context, in store.DataInput, out []byte) error {
	_, err := in.Read(out)
	return err
}

type lowercaseAscii struct {
}

func (*lowercaseAscii) Read(ctx context.Context, in store.DataInput, out []byte) error {
	return compress.LowercaseAsciiCompression.Decompress(ctx, in, out)
}

type lz4 struct {
}

func (*lz4) Read(ctx context.Context, in store.DataInput, out []byte) error {
	return compress.LZ4Compression.Decompress(in, out)
}

func ByCode(code int) (CompressionAlgorithm, error) {
	switch code {
	case 0:
		return &noCompression{}, nil
	case 1:
		return &lowercaseAscii{}, nil
	case 2:
		return &lz4{}, nil
	default:
		return nil, errors.New("unsupported compression")
	}
}
