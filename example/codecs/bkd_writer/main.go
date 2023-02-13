package main

import (
	"encoding/binary"
	"github.com/geange/lucene-go/codecs/bkd"
	"github.com/geange/lucene-go/codecs/simpletext"
	"github.com/geange/lucene-go/core/store"
)

func main() {
	dir, err := store.NewNIOFSDirectory("data")
	if err != nil {
		panic(err)
	}

	cfg, err := bkd.NewBKDConfig(2, 2, 4, 2)
	if err != nil {
		panic(err)
	}

	output, err := dir.CreateOutput("bkd.txt", nil)
	if err != nil {
		panic(err)
	}

	writer := simpletext.NewSimpleTextBKDWriter(
		100, dir, "demo", cfg, 16, 4)

	writer.Add(Point(5, 4), 1)
	writer.Add(Point(1, 2), 1)
	writer.Add(Point(1, 3), 1)
	writer.Add(Point(2, 9), 2)

	writer.Finish(output)
	output.Close()
}

func Point(values ...int) []byte {
	size := 4 * len(values)
	bs := make([]byte, size)
	for i := 0; i < len(values); i++ {
		binary.BigEndian.PutUint32(bs[i*4:], uint32(values[i]))
	}
	return bs
}
