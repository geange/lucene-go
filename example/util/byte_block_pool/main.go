package main

import (
	"fmt"
	"github.com/geange/lucene-go/core/util"
)

func main() {
	pool := util.NewByteBlockPool(util.NewBytesAllocator(util.BYTE_BLOCK_SIZE, &util.DirectBytesAllocator{}))
	pool.NewSlice(2)
	pool.Append([]byte("abcdefg"))

	pool.Append([]byte("abcdefg"))

	fmt.Println(pool)
}
