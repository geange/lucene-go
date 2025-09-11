package compress

import (
	"os"
	"testing"

	"github.com/geange/lucene-go/core/store"
)

func TestCompress(t *testing.T) {
	ht := NewFastCompressionHashTable()

	buf := store.NewBufferDataOutput()

	Compress([]byte("{\"data\":[{\"name\":\"xxx\"},{\"name\":\"www\"}]}"), buf, ht)

	t.Log(len(buf.Bytes()))

	f, _ := os.Create("lz4.out")
	f.Write(buf.Bytes())
	f.Close()
}
