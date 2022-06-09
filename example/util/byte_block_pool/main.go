package main

import (
	"fmt"
	"github.com/geange/lucene-go/core/util"
)

func main() {
	pool := util.NewByteBlockPool(util.NewBytesAllocator(util.BYTE_BLOCK_SIZE, &util.DirectBytesAllocator{}))
	pool.NewSlice(2)
	pool.Append(util.NewBytesRef([]byte("abcdefg"), 0, 7))

	pool.Append(util.NewBytesRef([]byte("abcdefg"), 0, 7))

	fmt.Println(pool)
}
