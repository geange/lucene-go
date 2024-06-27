package hash

import (
	"bytes"
	"encoding/gob"
	"sync"
)

func Compare[T any](a, b T) int {
	b1 := hash(a)
	b2 := hash(b)
	return bytes.Compare(b1, b2)
}

var localPool = &sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

func hash(s any) []byte {
	b := localPool.Get().(*bytes.Buffer)
	defer localPool.Put(b)

	gob.NewEncoder(b).Encode(s)
	return b.Bytes()
}
