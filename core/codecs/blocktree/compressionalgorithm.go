package blocktree

import "github.com/geange/lucene-go/core/store"

type CompressionAlgorithm interface {
	Read(in store.DataInput, bs []byte) error
}

func ByCode(code int) (CompressionAlgorithm, error) {
	panic("")
}
