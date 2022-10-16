package main

import (
	"fmt"
	"github.com/geange/lucene-go/core/util/packed"
)

func main() {
	writer, _ := packed.NewPagedGrowableWriter(16, 1<<27, 8, packed.COMPACT)
	writer.Set(12, 9)
	fmt.Println(writer.Get(12))

	writer.Grow(20)
	fmt.Println(writer.Get(12))
}
