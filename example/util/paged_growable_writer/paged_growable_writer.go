package main

import (
	"fmt"
	"github.com/geange/lucene-go/core/util/packed"
)

func main() {
	writer := packed.NewPagedGrowableWriter(16, 1<<27, 8, packed.COMPACT)
	writer.Grow(30)
	writer.Set(12, 9)
	fmt.Println(writer.Get(12))
}
