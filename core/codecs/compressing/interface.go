package compressing

import (
	"bytes"
	"context"
	"io"

	"github.com/geange/lucene-go/core/store"
)

type FieldsIndex interface {
	io.Closer

	// GetStartPointer
	// Get the start pointer for the block that contains the given docID.
	GetStartPointer(docId int) (int64, error)

	// CheckIntegrity
	// Check the integrity of the index.
	CheckIntegrity() error

	Clone() (FieldsIndex, error)
}

type Compressor interface {
	Compress(ctx context.Context, bytes []byte, out store.DataOutput) error
}

type Decompressor interface {
	io.Closer

	Decompress(ctx context.Context, in store.DataInput, offset int64, length int64, buf *bytes.Buffer) error

	Clone() Decompressor
}

type CompressionMode interface {
	NewCompressor() Compressor
	NewDecompressor() Decompressor
}
