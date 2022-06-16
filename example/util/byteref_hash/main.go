package main

import (
	"fmt"
	"github.com/geange/lucene-go/core/util"
)

func main() {
	pool := util.NewByteBlockPool(util.NewBytesAllocator(util.BYTE_BLOCK_SIZE, &util.DirectBytesAllocator{}))
	hash := util.NewBytesRefHash(pool)

	id, err := hash.Add([]byte("hello"))
	if err != nil {
		panic(err)
	}

	id2, err := hash.Add([]byte("hello world"))
	if err != nil {
		panic(err)
	}

	find := hash.Find([]byte("hello"))

	if id != find {
		panic(fmt.Sprintf("%d != %d", id, find))
	}

	fmt.Println(string(hash.Get(id)))
	fmt.Println(string(hash.Get(id2)))

}
